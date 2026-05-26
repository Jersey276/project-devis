package template

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "X-User-Id header manquant."})
		return
	}

	includeArchived := c.Query("archived") == "true"
	templateType := c.Query("type")

	query := `SELECT template_id, user_id, template_type, target_resource, name,
	                 archived_at, payload_version, payload, created_at, updated_at
	          FROM templates WHERE user_id=$1`
	args := []interface{}{userID}

	if !includeArchived {
		query += " AND archived_at IS NULL"
	}
	if templateType != "" {
		args = append(args, templateType)
		query += " AND template_type=$2"
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}
	defer rows.Close()

	templates := make([]Template, 0)
	for rows.Next() {
		var t Template
		var archivedAt sql.NullString
		var payloadRaw []byte
		if err := rows.Scan(
			&t.TemplateID, &t.UserID, &t.TemplateType, &t.TargetResource, &t.Name,
			&archivedAt, &t.PayloadVersion, &payloadRaw, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
			return
		}
		if archivedAt.Valid {
			t.ArchivedAt = &archivedAt.String
		}
		t.Payload = json.RawMessage(payloadRaw)
		templates = append(templates, t)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "templates": templates})
}
