package template

import (
	"database/sql"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createRequest struct {
	TemplateType   string `json:"template_type" binding:"required"`
	TargetResource string `json:"target_resource" binding:"required"`
	Name           string `json:"name" binding:"required"`
}

func Create(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "X-User-Id header manquant."})
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "Données invalides."})
		return
	}

	if !validTypes[req.TemplateType] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidTemplateType, "message": "Type de template invalide."})
		return
	}

	templateID := uuid.New().String()
	_, err := db.ExecContext(c.Request.Context(),
		`INSERT INTO templates (template_id, user_id, template_type, target_resource, name)
		 VALUES ($1, $2, $3, $4, $5)`,
		templateID, userID, req.TemplateType, req.TargetResource, req.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "code": codes.Success, "template_id": templateID})
}
