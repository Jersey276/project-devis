package template

import (
	"database/sql"
	"net/http"

	"project-devis-template/actions/codes"

	"github.com/gin-gonic/gin"
)

func Archive(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	res, err := db.ExecContext(c.Request.Context(),
		`UPDATE templates SET archived_at=now(), updated_at=now()
		 WHERE template_id=$1 AND user_id=$2 AND archived_at IS NULL`,
		templateID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Template introuvable ou déjà archivé."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "code": codes.Success})
}

func Restore(c *gin.Context, db *sql.DB) {
	userID := c.GetHeader("X-User-Id")
	templateID := c.Param("id")

	res, err := db.ExecContext(c.Request.Context(),
		`UPDATE templates SET archived_at=NULL, updated_at=now()
		 WHERE template_id=$1 AND user_id=$2 AND archived_at IS NOT NULL`,
		templateID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "code": codes.InternalError, "message": "Erreur interne."})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "code": codes.NotFound, "message": "Template introuvable ou non archivé."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "code": codes.Success})
}
