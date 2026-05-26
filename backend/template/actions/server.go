package actions

import (
	"database/sql"

	"project-devis-template/actions/line"
	templateAction "project-devis-template/actions/template"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	templates := api.Group("/templates")
	templates.GET("", func(c *gin.Context) { templateAction.List(c, s.db) })
	templates.POST("", func(c *gin.Context) { templateAction.Create(c, s.db) })
	templates.GET("/:id", func(c *gin.Context) { templateAction.Get(c, s.db) })
	templates.PUT("/:id", func(c *gin.Context) { templateAction.Update(c, s.db) })
	templates.DELETE("/:id", func(c *gin.Context) { templateAction.Delete(c, s.db) })
	templates.POST("/:id/archive", func(c *gin.Context) { templateAction.Archive(c, s.db) })
	templates.POST("/:id/restore", func(c *gin.Context) { templateAction.Restore(c, s.db) })

	lines := templates.Group("/:id/lines")
	lines.GET("", func(c *gin.Context) { line.List(c, s.db) })
	lines.POST("", func(c *gin.Context) { line.Create(c, s.db) })
	lines.PUT("/:lineId", func(c *gin.Context) { line.Update(c, s.db) })
	lines.DELETE("/:lineId", func(c *gin.Context) { line.Delete(c, s.db) })
}
