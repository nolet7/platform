package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type CatalogService struct {
	db   *sql.DB
	nats *nats.Conn
}

func main() {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Connect to NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	svc := &CatalogService{db: db, nats: nc}

	// Initialize Gin router
	r := gin.Default()

	// Middleware
	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())

	// Health checks
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	r.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/entities", svc.RegisterEntity)
		v1.GET("/entities/:id", svc.GetEntity)
		v1.GET("/entities", svc.SearchEntities)
		v1.PATCH("/entities/:id", svc.UpdateEntity)
		v1.DELETE("/entities/:id", svc.DeleteEntity)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting catalog service on port %s", port)
	r.Run(":" + port)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	// Simple implementation - replace with UUID in production
	return "req-" + randomString(16)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
