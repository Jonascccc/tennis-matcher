package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Jonascccc/tennis-matcher/server/config"
	"github.com/Jonascccc/tennis-matcher/server/db"
	"github.com/Jonascccc/tennis-matcher/server/handlers"
	"github.com/Jonascccc/tennis-matcher/server/middleware"
)

func main() {
	config.Load()

	ctx := context.Background()
	if err := db.Init(ctx, config.C.DatabaseURL); err != nil {
		log.Fatal("db init:", err)
	}

	jwtSecret := []byte(config.C.JWTSecret)

	r := gin.Default()
	api := r.Group("/api")

	api.POST("/auth/register", func(c *gin.Context) { middleware.Register(c, jwtSecret) })
	api.POST("/auth/login", func(c *gin.Context) { middleware.Login(c, jwtSecret) })

	authed := api.Group("")
	authed.Use(middleware.Auth(jwtSecret))
	{
		authed.GET("/me/profile", handlers.GetProfile)
		authed.PUT("/me/profile", handlers.PutProfile)
		authed.POST("/match/find", handlers.MatchFind)
	}

	log.Printf("server on :%s", config.C.Port)
	_ = r.Run(":" + config.C.Port)
}
