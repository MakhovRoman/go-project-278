package main

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}

func initSentry(router *gin.Engine) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Println("Sentry disabled: SENTRY_DSN is not found")
	} else {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			TracesSampleRate: 0.01,
		}); err != nil {
			log.Fatalf("failed to init Sentry: %v", err)
		}
		log.Println("✨ ~*~ wzhoooh ~*~ Sentry is ACTIVATED ~*~ ✨")
		router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := setupRouter()
	initSentry(router)
	defer sentry.Flush(2 * time.Second)

	router.GET("/ping", handlePing)
	// добавляю endpoint только для локальной разработки
	if gin.Mode() != gin.ReleaseMode {
		router.GET("/panic", handlePanic)
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
