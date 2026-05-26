package line

import (
	"database/sql"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

func Delete(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")
	lineID := c.Param("lineId")

	res, err := db.ExecContext(c.Request.Context(),
		`DELETE FROM template_lines tl
		 USING templates t
		 WHERE tl.line_id=$1 AND tl.template_id=$2 AND t.template_id=$2 AND t.user_id=$3`,
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
