package main

import (
	"database/sql"
	"go-project-278/internal/links"
	"go-project-278/internal/visits"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:5173"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Content-Length"},
		ExposeHeaders: []string{"Content-Range"},
	}))

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
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is missing")
	}

	conn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed connect to db: %v", err)
	}
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to close db connection: %v", err)
		}
	}(conn)

	router := setupRouter()
	initSentry(router)
	defer sentry.Flush(2 * time.Second)

	// добавляю endpoint только для локальной разработки
	if gin.Mode() != gin.ReleaseMode {
		router.GET("/panic", handlePanic)
	}
	// SERVICE
	router.GET("/ping", handlePing)
	// LINKS
	linkSvc := links.NewService(conn)
	links.NewHandler(linkSvc).Register(router)

	// VISITS
	visitSvc := visits.NewService(conn)
	visits.NewHandler(linkSvc, visitSvc).Register(router)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
