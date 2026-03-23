package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

func (s *CatalogService) RegisterEntity(c *gin.Context) {
	var entity Entity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID if not provided
	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(entity.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata format"})
		return
	}

	// Insert into database
	query := `
		INSERT INTO entities (id, type, name, owner_team, owner_email, tier, data_classification, metadata, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`
	
	err = s.db.QueryRow(
		query,
		entity.ID, entity.Type, entity.Name, entity.OwnerTeam, entity.OwnerEmail,
		entity.Tier, entity.DataClassification, metadataJSON,
		c.GetString("user"), c.GetString("user"),
	).Scan(&entity.CreatedAt, &entity.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Publish event to NATS
	eventData, _ := json.Marshal(map[string]interface{}{
		"event_type": "entity.registered",
		"entity_id":  entity.ID,
		"entity":     entity,
		"timestamp":  time.Now(),
	})
	s.nats.Publish("platform.events", eventData)

	c.JSON(http.StatusCreated, entity)
}

func (s *CatalogService) GetEntity(c *gin.Context) {
	id := c.Param("id")

	var entity Entity
	var metadataJSON []byte

	query := `
		SELECT id, type, name, owner_team, owner_email, tier, data_classification, 
		       metadata, created_at, updated_at, created_by, updated_by
		FROM entities WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&entity.ID, &entity.Type, &entity.Name, &entity.OwnerTeam, &entity.OwnerEmail,
		&entity.Tier, &entity.DataClassification, &metadataJSON,
		&entity.CreatedAt, &entity.UpdatedAt, &entity.CreatedBy, &entity.UpdatedBy,
	)

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
	argCount := 1

	if entityType != "" {
		query += " AND type = $" + string(rune(argCount+'0'))
		args = append(args, entityType)
		argCount++
	}

	if owner != "" {
		query += " AND owner_team = $" + string(rune(argCount+'0'))
		args = append(args, owner)
		argCount++
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

		err := rows.Scan(
			&entity.ID, &entity.Type, &entity.Name, &entity.OwnerTeam, &entity.OwnerEmail,
			&entity.Tier, &entity.DataClassification, &metadataJSON,
			&entity.CreatedAt, &entity.UpdatedAt,
		)
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

	// Build dynamic update query
	query := "UPDATE entities SET updated_at = NOW(), updated_by = $1"
	args := []interface{}{c.GetString("user")}
	argCount := 2

	if name, ok := updates["name"]; ok {
		query += ", name = $" + string(rune(argCount+'0'))
		args = append(args, name)
		argCount++
	}

	query += " WHERE id = $" + string(rune(argCount+'0'))
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
