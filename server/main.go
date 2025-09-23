package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var db *pgxpool.Pool
var jwtSecret []byte

func main() {
	env := os.Getenv("GO_ENV")
	filename := ".env"

	if env == "test" {
		filename = ".env.test"
	}
	if err := godotenv.Load(filename); err != nil {
		log.Printf("No %s file found, relying on system env vars", filename)
	}

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err = pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	db = pool

	r := gin.Default()
	api := r.Group("/api")
	{
		api.POST("/auth/register", register)
		api.POST("/auth/login", login)

		authed := api.Group("")
		authed.Use(authMiddleware())
		{
			authed.GET("/me/profile", getProfile)
			authed.POST("/me/profile:", putprofile)
			authed.POST("/match/find", matchFind) //v0.1 corefunction
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("sever running on port %s", port)
	_ = r.Run(":" + port)
}

/************ Auth *************/
type creds struct {
	Email    string
	Password string
}

func register(c *gin.Context) {
	var in creds
	err := c.ShouldBindJSON(&in)
	if err != nil {
		c.JSON(400, gin.H{"error": "Bad request"})
		return
	}
	if in.Email == "" || in.Password == "" {
		c.JSON(400, gin.H{"error": "email/password required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}
	var uid int
	row := db.QueryRow(c, "INSERT INTO app_user(email, password_hash) VALUES($1, $2) RETURNING user_id", in.Email, string(hash))
	err = row.Scan(&uid)
	if err != nil {
		c.JSON(409, gin.H{"error": "email already in use"})
		return
	}
	_, _ = db.Exec(c, "INSERT INTO tennis_profile(user_id) VALUES($1)", uid)
	token := jwtNew(uid)
	c.JSON(200, gin.H{"token": token})
}

func login(c *gin.Context) {
	var in creds
	err := c.ShouldBindJSON(&in)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad payload"})
		return
	}
	var uid int
	var hash string
	err = db.QueryRow(c, "SELECT user_id, password_hash FROM app_user WHERE email = $1", in.Email).Scan(&uid, &hash)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(200, gin.H{"token": jwtNew(uid)})
}

func jwtNew(uid int) string {
	claims := jwt.MapClaims{}
	claims["uid"] = uid
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("failed to sign token: %v", err)
		return ""
	}
	return signedToken
}

func authMiddleware(c *gin.Context) {
	// check if the authorization head is present
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// check if the authorization header is in the correct format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}

	// extract the token from the header
	tokenStr := strings.TrimPrefix(authHeader, bearerPrefix)

	// parse the token and verify the signature
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		// check if the signing method is HS256
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})

	// check if the token is valid
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// extract the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	// extract the user id from the claims
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}
	uid := int(uidFloat)

	// set the user id in the context
	c.Set("uid", uid)
	c.Next()
}

/************ Profile *************/
type Profile struct {
	UserID           int            `json:"-"`
	Handedness       string         `json:"Handedness"`
	LevelEst         float64        `json:"level_est"`
	Elo              int            `json:"elo"`
	PreferredFormats []string       `json:"preferred_formats"`
	RadiusKM         int            `json:"radiusKm"`
	Availability     map[string]any `json:"availability"`
	HomeLat          *float64       `json:"homeLat"`
	HomeLng          *float64       `json:"homeLng"`
}

func getProfile(c *gin.Context) {
	uid := c.GetInt("uid")

	var (
		p               Profile
		availabilityRaw []byte
	)

	p.UserID = uid

	const q = `
		SELECT 
			tp.handedness,
			tp.level_est,
			tp.elo,
			tp.preferred_formats,
			tp.radius_km,
			tp.availability,
			ST_Y(ul.geom::geometry) AS lat,
			ST_X(ul.geom::geometry) AS lng
		FROM tennis_profile tp
		LEFT JOIN user_locations ul
			  ON ul.user_id = tp.user_id
		     And ul.label = 'home'
		WHERE tp.user_id = $1
	`

	row := db.QueryRow(c, q, uid)
	if err := row.Scan(
		&p.Handedness,
		&p.LevelEst,
		&p.Elo,
		&p.PreferredFormats,
		&p.RadiusKM,
		&availabilityRaw,
		&p.HomeLat,
		&p.HomeLng,
	); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	if len(availabilityRaw) > 0 {
		err := json.Unmarshal(availabilityRaw, &p.Availability)
		if err != nil {
			p.Availability = map[string]any{}
		}
	} else {
		p.Availability = map[string]any{}
	}
	c.JSON(http.StatusOK, p)
}

func putprofile(c *gin.Context) {
	uid := c.GetInt("uid")
	var (
		p               Profile
		availabilityRaw []byte
	)

	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad payload"})
		return
	}

	err = db.QueryRow(c,
		`SELECT tp.handedness, tp.level_est, tp.elo, tp.preferred_formats, tp.radis_km, tp.availability,
		ST_Y(ul.geom::geometry) AS lon,
		ST_X(ul.geom::geometry) AS lat,
		FROM tennis_profile tp
		LEFT JOIN user_locations ul
			ON ul.user_id = tp.user_id
			AND ul.label = 'home'
		WHERE tp.user_id = $1`, uid).Scan(&p.Handedness, &p.LevelEst, &p.Elo, &p.PreferredFormats, &p.RadiusKM, &availabilityRaw, &p.HomeLng, &p.HomeLat)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	if p.Handedness != "" && p.Handedness != "L" && p.Handedness != "R" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "handedness must be R or L"})
		return
	}
	
	if p.RadiusKM < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "radius must be greater than 0"})
		return
	}

	if p.LevelEst < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "level must be greater than 0"})
		return
	}

	if p.Elo < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "elo must be greater than 0"})
		return
	}

	var availability map[string]any
	var err error
	if p.Availability != nil {
		availability, err = json.Marshal(p.Availability)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid availability"})
			return
		} else {
			availability = []byte("{}")
		}

}
