package handlers

import (
	"encoding/json"

	"github.com/Jonascccc/tennis-matcher/server/db"

	"github.com/gin-gonic/gin"
)

type Profile struct {
	Handedness       string         `json:"handedness"` // "R"|"L"
	LevelEst         float64        `json:"levelEst"`
	Elo              int            `json:"elo"`
	PreferredFormats []string       `json:"preferredFormats"` // text[]
	RadiusKm         int            `json:"radiusKm"`
	Availability     map[string]any `json:"availability"`
	HomeLat          *float64       `json:"homeLat"`
	HomeLng          *float64       `json:"homeLng"`
}

func GetProfile(c *gin.Context) {
	uid := c.GetInt("uid")
	var p Profile
	var availability []byte
	err := db.Pool.QueryRow(c, `
SELECT tp.handedness, tp.level_est, tp.elo, tp.preferred_formats, tp.radius_km, tp.availability,
       ST_Y(ul.geom::geometry) AS lat, ST_X(ul.geom::geometry) AS lng
FROM tennis_profile tp
LEFT JOIN user_locations ul ON ul.user_id=tp.user_id AND ul.label='home'
WHERE tp.user_id=$1`, uid).Scan(
		&p.Handedness, &p.LevelEst, &p.Elo, &p.PreferredFormats, &p.RadiusKm, &availability, &p.HomeLat, &p.HomeLng,
	)
	if err != nil {
		c.JSON(404, gin.H{"error": "profile not found"})
		return
	}
	if len(availability) > 0 {
		_ = json.Unmarshal(availability, &p.Availability)
	}
	if p.Availability == nil {
		p.Availability = map[string]any{}
	}
	c.JSON(200, p)
}

func PutProfile(c *gin.Context) {
	uid := c.GetInt("uid")
	var in Profile
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "bad payload"})
		return
	}

	if in.Handedness != "" && in.Handedness != "R" && in.Handedness != "L" {
		c.JSON(400, gin.H{"error": "handedness must be R or L"})
		return
	}
	if in.RadiusKm < 0 {
		c.JSON(400, gin.H{"error": "radius must be >= 0"})
		return
	}
	if in.LevelEst < 0 {
		c.JSON(400, gin.H{"error": "levelEst must be >= 0"})
		return
	}
	if in.Elo < 0 {
		c.JSON(400, gin.H{"error": "elo must be >= 0"})
		return
	}

	availBytes, _ := json.Marshal(in.Availability)

	_, err := db.Pool.Exec(c, `
UPDATE tennis_profile
SET handedness=$1, level_est=$2, preferred_formats=$3, radius_km=$4, availability=$5
WHERE user_id=$6`, in.Handedness, in.LevelEst, in.PreferredFormats, in.RadiusKm, availBytes, uid)
	if err != nil {
		c.JSON(500, gin.H{"error": "update profile failed"})
		return
	}

	if in.HomeLat != nil && in.HomeLng != nil {
		lat, lng := *in.HomeLat, *in.HomeLng
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			c.JSON(400, gin.H{"error": "invalid home location"})
			return
		}
		_, err = db.Pool.Exec(c, `
INSERT INTO user_locations(user_id,label,geom)
VALUES ($1,'home', ST_SetSRID(ST_MakePoint($2,$3),4326)::geography)
ON CONFLICT (user_id,label) DO UPDATE SET geom=EXCLUDED.geom, updated_at=now()
`, uid, lng, lat) // 注意 lng, lat 顺序
		if err != nil {
			c.JSON(500, gin.H{"error": "update home location failed"})
			return
		}
	}
	c.JSON(200, gin.H{"ok": true})
}
