package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type CatalogService struct {
	db   *sql.DB
	nats *nats.Conn
}

type Entity struct {
	ID                 string                 `json:"id"`
	Type               string                 `json:"type" binding:"required"`
	Name               string                 `json:"name" binding:"required"`
	OwnerTeam          string                 `json:"owner_team" binding:"required"`
	OwnerEmail         string                 `json:"owner_email" binding:"required"`
	Tier               string                 `json:"tier"`
	DataClassification string                 `json:"data_classification"`
	Metadata           map[string]interface{} `json:"metadata"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	CreatedBy          string                 `json:"created_by"`
	UpdatedBy          string                 `json:"updated_by"`
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Println("Warning: Failed to connect to NATS:", err)
		nc = nil
	}
	if nc != nil {
		defer nc.Close()
	}

	svc := &CatalogService{db: db, nats: nc}

	r := gin.Default()
	r.Use(CORSMiddleware())

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

	v1 := r.Group("/api/v1")
	{
		v1.POST("/entities", svc.RegisterEntity)
		v1.GET("/entities/:id", svc.GetEntity)
		v1.GET("/entities", svc.SearchEntities)
		v1.PATCH("/entities/:id", svc.UpdateEntity)
		v1.DELETE("/entities/:id", svc.DeleteEntity)
	}

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

func (s *CatalogService) RegisterEntity(c *gin.Context) {
	var entity Entity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}

	metadataJSON, err := json.Marshal(entity.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata format"})
		return
	}

	query := `INSERT INTO entities (id, type, name, owner_team, owner_email, tier, data_classification, metadata, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING created_at, updated_at`

	err = s.db.QueryRow(query, entity.ID, entity.Type, entity.Name, entity.OwnerTeam, entity.OwnerEmail,
		entity.Tier, entity.DataClassification, metadataJSON, "system", "system").Scan(&entity.CreatedAt, &entity.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	if s.nats != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"event_type": "entity.registered",
			"entity_id":  entity.ID,
			"timestamp":  time.Now(),
		})
		if err := s.nats.Publish("platform.events", eventData); err != nil {
			log.Printf("Warning: failed to publish NATS event for entity %s: %v", entity.ID, err)
		}
	}

	c.JSON(http.StatusCreated, entity)
}

func (s *CatalogService) GetEntity(c *gin.Context) {
	id := c.Param("id")
	var entity Entity
	var metadataJSON []byte

	query := `SELECT id, type, name, owner_team, owner_email, tier, data_classification, 
		metadata, created_at, updated_at, created_by, updated_by FROM entities WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(&entity.ID, &entity.Type, &entity.Name, &entity.OwnerTeam, &entity.OwnerEmail,
		&entity.Tier, &entity.DataClassification, &metadataJSON, &entity.CreatedAt, &entity.UpdatedAt, &entity.CreatedBy, &entity.UpdatedBy)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	json.Unmarshal(metadataJSON, &entity.Metadata)
	c.JSON(http.StatusOK, entity)
}

func (s *CatalogService) SearchEntities(c *gin.Context) {
	entityType := c.Query("type")
	owner := c.Query("owner")

	query := "SELECT id, type, name, owner_team, owner_email, tier, data_classification, metadata, created_at, updated_at FROM entities WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if entityType != "" {
		query += fmt.Sprintf(" AND type = $%d", argPos)
		args = append(args, entityType)
		argPos++
	}

	if owner != "" {
		query += fmt.Sprintf(" AND owner_team = $%d", argPos)
		args = append(args, owner)
		argPos++
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	entities := []Entity{}
	for rows.Next() {
		var entity Entity
		var metadataJSON []byte

		err := rows.Scan(&entity.ID, &entity.Type, &entity.Name, &entity.OwnerTeam, &entity.OwnerEmail,
			&entity.Tier, &entity.DataClassification, &metadataJSON, &entity.CreatedAt, &entity.UpdatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(metadataJSON, &entity.Metadata)
		entities = append(entities, entity)
	}

	c.JSON(http.StatusOK, gin.H{"entities": entities, "count": len(entities)})
}

func (s *CatalogService) UpdateEntity(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE entities SET updated_at = NOW(), updated_by = 'system'"
	args := []interface{}{}
	argPos := 1

	if name, ok := updates["name"]; ok {
		query += fmt.Sprintf(", name = $%d", argPos)
		args = append(args, name)
		argPos++
	}

	if argPos == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *CatalogService) DeleteEntity(c *gin.Context) {
	id := c.Param("id")
	_, err := s.db.Exec("DELETE FROM entities WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
