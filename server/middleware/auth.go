package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Jonascccc/tennis-matcher/server/db"
	"google.golang.org/api/idtoken"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type creds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Auth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		ah := c.GetHeader("Authorization")
		if !strings.HasPrefix(ah, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or bad Authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(ah, "Bearer ")

		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		uidf, ok := claims["uid"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid uid"})
			return
		}
		c.Set("uid", int(uidf))
		c.Next()
	}
}

func jwtNew(uid int, secret []byte) (string, error) {
	claims := jwt.MapClaims{"uid": uid, "exp": time.Now().Add(24 * time.Hour).Unix()}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func Register(c *gin.Context, jwtSecret []byte) {
	var in creds
	if err := c.ShouldBindJSON(&in); err != nil || in.Email == "" || in.Password == "" {
		c.JSON(400, gin.H{"error": "email/password required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "hash fail"})
		return
	}

	var uid int
	if err := db.Pool.QueryRow(c, `INSERT INTO app_user(email,password_hash) VALUES($1,$2) RETURNING user_id`, in.Email, string(hash)).Scan(&uid); err != nil {
		c.JSON(409, gin.H{"error": "email exists?"})
		return
	}
	_, _ = db.Pool.Exec(c, `INSERT INTO tennis_profile(user_id) VALUES($1)`, uid)

	tok, _ := jwtNew(uid, jwtSecret)
	c.JSON(200, gin.H{"token": tok})
}

func Login(c *gin.Context, jwtSecret []byte) {
	var in creds
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "bad payload"})
		return
	}
	var uid int
	var hash string
	err := db.Pool.QueryRow(c, `SELECT user_id,password_hash FROM app_user WHERE email=$1`, in.Email).Scan(&uid, &hash)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		log.Println("db err:", err)
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	tok, _ := jwtNew(uid, jwtSecret)
	c.JSON(200, gin.H{"token": tok})
}

func GoogleLogin(c *gin.Context, jwtSecret []byte, googleClientID string) {
	var in struct {
		IDToken string `json:"idToken"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || strings.TrimSpace(in.IDToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idToken required"})
		return
	}

	// Verify ID token with Google
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, in.IDToken, googleClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid id_token"})
		return
	}

	emailAny := payload.Claims["email"]
	email, _ := emailAny.(string)
	if strings.TrimSpace(email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email not present"})
		return
	}

	// Find or create user
	var uid int
	err = db.Pool.QueryRow(c, `SELECT user_id FROM app_user WHERE email=$1`, email).Scan(&uid)
	if err == pgx.ErrNoRows {
		err = db.Pool.QueryRow(c,
			`INSERT INTO app_user(email,password_hash) VALUES($1,$2) RETURNING user_id`,
			email, "google-oauth",
		).Scan(&uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
			return
		}
		_, _ = db.Pool.Exec(c, `INSERT INTO tennis_profile(user_id) VALUES($1)`, uid)
	} else if err != nil {
		log.Println("db err:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	tok, _ := jwtNew(uid, jwtSecret)
	c.JSON(http.StatusOK, gin.H{"token": tok})
}
