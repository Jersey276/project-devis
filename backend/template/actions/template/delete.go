package template

import (
	"database/sql"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

func Delete(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	res, err := db.ExecContext(c.Request.Context(),
		`DELETE FROM templates WHERE template_id=$1 AND user_id=$2`,
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
