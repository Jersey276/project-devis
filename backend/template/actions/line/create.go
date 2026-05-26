package line

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"project-devis-template/actions/codes"
	"project-devis-template/actions/sqlutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createLineRequest struct {
	Type      string          `json:"type" binding:"required"`
	Name      string          `json:"name"`
	Quantity  string          `json:"quantity" binding:"required"`
	Unit      string          `json:"unit"`
	UnitPrice int64           `json:"unit_price"`
	Data      json.RawMessage `json:"data"`
	Position  int32           `json:"position"`
	TaxID     int32           `json:"tax_id"`
}

func Create(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	var req createLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "Données invalides."})
		return
	}
	if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "Quantité invalide."})
		return
	}

	// verify template belongs to user
	var count int
	err := db.QueryRowContext(c.Request.Context(),
		`SELECT COUNT(1) FROM templates WHERE template_id=$1 AND user_id=$2 AND archived_at IS NULL`,
		templateID, userID,
	).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Template introuvable."})
		return
	}

	dataStr := ""
	if len(req.Data) > 0 {
		dataStr = string(req.Data)
	} else {
		dataStr = "{}"
	}

	lineID := uuid.New().String()
	_, err = db.ExecContext(c.Request.Context(),
		`INSERT INTO template_lines (line_id, template_id, type, name, quantity, unit, unit_price, data, position, tax_id)
		 VALUES ($1, $2, $3, $4, $5::DECIMAL, $6, $7, $8::jsonb, $9, $10)`,
		lineID, templateID, req.Type, req.Name, req.Quantity,
		sqlutil.NullableStr(req.Unit), req.UnitPrice, dataStr, req.Position,
		sqlutil.NullableInt32(req.TaxID),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "code": codes.Success, "line_id": lineID})
}
