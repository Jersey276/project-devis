package controllers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	TemplateCodeNotFound            int32 = 1001
	TemplateCodeAlreadyExists       int32 = 1002
	TemplateCodeInvalidInput        int32 = 1003
	TemplateCodeInvalidTemplateType int32 = 1004
	TemplateCodeInternalError       int32 = 2001
)

var templateErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		TemplateCodeNotFound:            {http.StatusNotFound, "Template introuvable."},
		TemplateCodeAlreadyExists:       {http.StatusConflict, "Ce template existe déjà."},
		TemplateCodeInvalidInput:        {http.StatusBadRequest, "Données invalides."},
		TemplateCodeInvalidTemplateType: {http.StatusBadRequest, "Type de template invalide."},
		TemplateCodeInternalError:       {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service templates indisponible.",
}

type templateClient struct {
	base   string
	client *http.Client
}

func newTemplateClient() *templateClient {
	address := os.Getenv("TEMPLATE_SERVICE_ADDRESS")
	if address == "" {
		address = "http://localhost:8085"
	}
	return &templateClient{base: address, client: &http.Client{}}
}

func (tc *templateClient) proxy(c *gin.Context, method, path string) {
	var body io.Reader
	if c.Request.Body != nil {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": templateErrors.unavailableMessage})
			return
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), method, tc.base+path, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": templateErrors.unavailableMessage})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", userIDFromCtx(c))

	// forward query params
	req.URL.RawQuery = c.Request.URL.RawQuery

	resp, err := tc.client.Do(req)
	if err != nil {
		log.Printf("template service unreachable: %v", err)
		templateErrors.unavailable(c)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// TemplateRoutes wires the /templates API group to the template HTTP service.
func TemplateRoutes(r *gin.RouterGroup) {
	tc := newTemplateClient()

	r.GET("", func(c *gin.Context) { tc.proxy(c, http.MethodGet, "/api/templates") })
	r.POST("", func(c *gin.Context) { tc.proxy(c, http.MethodPost, "/api/templates") })
	r.GET("/:id", func(c *gin.Context) { tc.proxy(c, http.MethodGet, "/api/templates/"+c.Param("id")) })
	r.PUT("/:id", func(c *gin.Context) { tc.proxy(c, http.MethodPut, "/api/templates/"+c.Param("id")) })
	r.DELETE("/:id", func(c *gin.Context) { tc.proxy(c, http.MethodDelete, "/api/templates/"+c.Param("id")) })
	r.POST("/:id/archive", func(c *gin.Context) { tc.proxy(c, http.MethodPost, "/api/templates/"+c.Param("id")+"/archive") })
	r.POST("/:id/restore", func(c *gin.Context) { tc.proxy(c, http.MethodPost, "/api/templates/"+c.Param("id")+"/restore") })

	lines := r.Group("/:id/lines")
	lines.GET("", func(c *gin.Context) { tc.proxy(c, http.MethodGet, "/api/templates/"+c.Param("id")+"/lines") })
	lines.POST("", func(c *gin.Context) { tc.proxy(c, http.MethodPost, "/api/templates/"+c.Param("id")+"/lines") })
	lines.PUT("/:lineId", func(c *gin.Context) { tc.proxy(c, http.MethodPut, "/api/templates/"+c.Param("id")+"/lines/"+c.Param("lineId")) })
	lines.DELETE("/:lineId", func(c *gin.Context) { tc.proxy(c, http.MethodDelete, "/api/templates/"+c.Param("id")+"/lines/"+c.Param("lineId")) })
}
