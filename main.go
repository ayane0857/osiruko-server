package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"osiruko/handlers/api"
	"osiruko/handlers/oidc"
	"osiruko/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}
	oidc.InitOIDC()
	c := cache.New(5*time.Minute, 10*time.Minute)
	r := gin.Default()
	dbUser := os.Getenv("POSTGRES_USER")
	log.Printf("DEBUG: dbUser is '%s'\n", dbUser)
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := "5432"
	log.Println("Connecting to database with user:", dbUser, "and db name:", dbName)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(&models.Users{})

	r.GET("/auth/login", oidc.HandleAuthLogin(c))
	r.GET("/callback", oidc.HandleCallback(c,db))

	r.GET("/api/status", api.StatusHandler(db))
	r.GET("/api/issue-code", api.IssueCodeHandler(c,db))

	r.Run(":8080")
}