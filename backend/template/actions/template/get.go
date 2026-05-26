package template

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	var t Template
	var archivedAt sql.NullString
	var payloadRaw []byte

	err := db.QueryRowContext(c.Request.Context(),
		`SELECT template_id, user_id, template_type, target_resource, name,
		        archived_at, payload_version, payload, created_at, updated_at
		 FROM templates WHERE template_id=$1 AND user_id=$2`,
		templateID, userID,
	).Scan(
		&t.TemplateID, &t.UserID, &t.TemplateType, &t.TargetResource, &t.Name,
		&archivedAt, &t.PayloadVersion, &payloadRaw, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Template introuvable."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}

	if archivedAt.Valid {
		t.ArchivedAt = &archivedAt.String
	}
	t.Payload = json.RawMessage(payloadRaw)

	c.JSON(http.StatusOK, gin.H{"success": true, "template": t})
}
