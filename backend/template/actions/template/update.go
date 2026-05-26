package template

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

type updateRequest struct {
	Name           string          `json:"name"`
	TargetResource string          `json:"target_resource"`
	Payload        json.RawMessage `json:"payload"`
	PayloadVersion int             `json:"payload_version"`
}

func Update(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "Données invalides."})
		return
	}

	res, err := db.ExecContext(c.Request.Context(),
		`UPDATE templates SET name=COALESCE(NULLIF($1,''), name),
		        target_resource=COALESCE(NULLIF($2,''), target_resource),
		        payload=CASE WHEN $3::text IS NOT NULL AND $3::text != '' THEN $3::jsonb ELSE payload END,
		        payload_version=CASE WHEN $4 > 0 THEN $4 ELSE payload_version END,
		        updated_at=now()
		 WHERE template_id=$5 AND user_id=$6 AND archived_at IS NULL`,
		req.Name, req.TargetResource, nullableJSON(req.Payload), req.PayloadVersion,
		templateID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Template introuvable."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "code": codes.Success})
}

func nullableJSON(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return string(raw)
}
