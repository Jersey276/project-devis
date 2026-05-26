package line

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	// verify template belongs to user
	var count int
	err := db.QueryRowContext(c.Request.Context(),
		`SELECT COUNT(1) FROM templates WHERE template_id=$1 AND user_id=$2`,
		templateID, userID,
	).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Template introuvable."})
		return
	}

	rows, err := db.QueryContext(c.Request.Context(),
		`SELECT line_id, template_id, type, name, quantity::text, unit, unit_price, data, position, tax_id, created_at, updated_at
		 FROM template_lines WHERE template_id=$1 ORDER BY position ASC`,
		templateID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}
	defer rows.Close()

	lines := make([]TemplateLine, 0)
	for rows.Next() {
		var l TemplateLine
		var unit sql.NullString
		var taxID sql.NullInt32
		var dataRaw []byte
		if err := rows.Scan(
			&l.LineID, &l.TemplateID, &l.Type, &l.Name, &l.Quantity,
			&unit, &l.UnitPrice, &dataRaw, &l.Position, &taxID,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
			return
		}
		if unit.Valid {
			l.Unit = &unit.String
		}
		if taxID.Valid {
			l.TaxID = &taxID.Int32
		}
		l.Data = json.RawMessage(dataRaw)
		lines = append(lines, l)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "lines": lines})
}
