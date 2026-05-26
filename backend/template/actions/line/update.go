package line

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"project-devis-template/actions/codes"
	"project-devis-template/actions/sqlutil"

	"github.com/gin-gonic/gin"
)

type updateLineRequest struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Quantity  string          `json:"quantity"`
	Unit      string          `json:"unit"`
	UnitPrice int64           `json:"unit_price"`
	Data      json.RawMessage `json:"data"`
	Position  int32           `json:"position"`
	TaxID     int32           `json:"tax_id"`
}

func Update(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")
	lineID := c.Param("lineId")

	var req updateLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "Données invalides."})
		return
	}
	if req.Quantity != "" {
		if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": codes.InvalidInput, "message": "Quantité invalide."})
			return
		}
	}

	dataStr := interface{}(nil)
	if len(req.Data) > 0 {
		dataStr = string(req.Data)
	}

	res, err := db.ExecContext(c.Request.Context(),
		`UPDATE template_lines tl
		 SET type=COALESCE(NULLIF($1,''), tl.type),
		     name=COALESCE(NULLIF($2,''), tl.name),
		     quantity=COALESCE(NULLIF($3,'')::DECIMAL, tl.quantity),
		     unit=$4,
		     unit_price=$5,
		     data=CASE WHEN $6::text IS NOT NULL THEN $6::jsonb ELSE tl.data END,
		     position=$7,
		     tax_id=$8,
		     updated_at=now()
		 FROM templates t
		 WHERE tl.line_id=$9 AND tl.template_id=$10 AND t.template_id=$10 AND t.user_id=$11`,
		req.Type, req.Name, sqlutil.NullableStr(req.Quantity),
		sqlutil.NullableStr(req.Unit), req.UnitPrice,
		dataStr, req.Position, sqlutil.NullableInt32(req.TaxID),
		lineID, templateID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Ligne introuvable."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "code": codes.Success})
}
