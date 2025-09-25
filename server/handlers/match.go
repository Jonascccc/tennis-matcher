package handlers

import (
	"math"
	"strings"
	"time"

	"github.com/Jonascccc/tennis-matcher/server/db"
	"github.com/gin-gonic/gin"
)

type FindReq struct {
	CenterLat   float64 `json:"centerLat"`
	CenterLng   float64 `json:"centerLng"`
	RadiusKm    float64 `json:"radiusKm"`
	Format      string  `json:"format"`      // "SINGLES"|"DOUBLES"
	WindowStart string  `json:"windowStart"` // RFC3339
	Limit       int     `json:"limit"`
}
type Candidate struct {
	UserID int     `json:"userId"`
	Elo    int     `json:"elo"`
	Meters float64 `json:"meters"`
	Score  float64 `json:"score"`
}
type Suggestion struct {
	CourtID int       `json:"courtId"`
	StartAt time.Time `json:"startAt"`
	EndAt   time.Time `json:"endAt"`
	Message string    `json:"message"`
}
type FindResp struct {
	Candidates  []Candidate  `json:"candidates"`
	Suggestions []Suggestion `json:"suggestions"`
}

func validateFindReq(in FindReq) error {
	if in.CenterLat < -90 || in.CenterLat > 90 {
		return errf("centerLat out of range")
	}
	if in.CenterLng < -180 || in.CenterLng > 180 {
		return errf("centerLng out of range")
	}
	if in.RadiusKm <= 0 || in.RadiusKm > 100 {
		return errf("radiusKm out of range (0,100]")
	}
	if strings.TrimSpace(in.Format) == "" {
		return errf("format required")
	}
	return nil
}
func errf(s string) error { return &gin.Error{Err: &simpleErr{s}} }

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }

func computeScore(myElo, theirElo int, meters float64) float64 {
	level := -math.Abs(float64(theirElo-myElo)) / 50.0
	geo := -meters / 1000.0
	return 0.6*level + 0.4*geo
}

func MatchFind(c *gin.Context) {
	uid := c.GetInt("uid")
	var in FindReq
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "bad payload"})
		return
	}
	if err := validateFindReq(in); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var myElo int
	if err := db.Pool.QueryRow(c, `SELECT elo FROM tennis_profile WHERE user_id=$1`, uid).Scan(&myElo); err != nil {
		c.JSON(500, gin.H{"error": "fetch my elo failed"})
		return
	}

	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := db.Pool.Query(c, `
SELECT u.user_id, tp.elo,
       ST_Distance(ul.geom, ST_SetSRID(ST_MakePoint($1,$2),4326)::geography) AS meters
FROM app_user u
JOIN tennis_profile tp ON tp.user_id=u.user_id
JOIN user_locations ul ON ul.user_id=u.user_id AND ul.label='home'
WHERE u.user_id<>$3
  AND ST_DWithin(ul.geom, ST_SetSRID(ST_MakePoint($1,$2),4326)::geography, $4*1000)
  AND $5 = ANY(tp.preferred_formats)
ORDER BY meters ASC
LIMIT $6
`, in.CenterLng, in.CenterLat, uid, in.RadiusKm, in.Format, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	var cs []Candidate
	for rows.Next() {
		var cd Candidate
		if err := rows.Scan(&cd.UserID, &cd.Elo, &cd.Meters); err == nil {
			cd.Score = computeScore(myElo, cd.Elo, cd.Meters)
			cs = append(cs, cd)
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{"error": "scan rows failed"})
		return
	}

	ws := time.Now()
	if t, err := time.Parse(time.RFC3339, in.WindowStart); err == nil {
		ws = t
	}
	sugs := []Suggestion{
		{CourtID: 1, StartAt: ws.Add(2 * time.Hour), EndAt: ws.Add(3 * time.Hour), Message: "周四晚 7-8 点@城市公园硬地场？"},
		{CourtID: 1, StartAt: ws.Add(26 * time.Hour), EndAt: ws.Add(27 * time.Hour), Message: "周五晚 9-10 点也行～"},
	}

	c.JSON(200, FindResp{Candidates: cs, Suggestions: sugs})
}
