package main

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(errorHandler())

	router.GET("/ping", func(c *gin.Context) {
		router.GET("/ping", func(c *gin.Context) {
			if err := c.Error(errors.New("[pong] something went wrong")); err != nil {
				log.Println(err)
			}
			c.String(200, "pong")
		})
		c.String(200, "pong")
	})

	return router
}

func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			log.Println(err.Error())
		}
	}
}

func main() {
	router := setupRouter()

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
