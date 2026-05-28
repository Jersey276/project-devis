package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	template "gateway/template"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// TemplateRoutes wires the /templates API group to the template gRPC service.
func TemplateRoutes(r *gin.RouterGroup) {
	address := os.Getenv("TEMPLATE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50055"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to template gRPC server: %v", err)
	}
	client := template.NewTemplateServiceClient(conn)

	r.GET("", func(c *gin.Context) { ListTemplates(c, client) })
	r.POST("", func(c *gin.Context) { CreateTemplate(c, client) })
	r.GET("/:id", func(c *gin.Context) { GetTemplate(c, client) })
	r.PUT("/:id", func(c *gin.Context) { UpdateTemplate(c, client) })
	r.DELETE("/:id", func(c *gin.Context) { DeleteTemplate(c, client) })
	r.POST("/:id/archive", func(c *gin.Context) { ArchiveTemplate(c, client) })
	r.POST("/:id/restore", func(c *gin.Context) { RestoreTemplate(c, client) })

	lines := r.Group("/:id/lines")
	lines.GET("", func(c *gin.Context) { ListTemplateLines(c, client) })
	lines.POST("", func(c *gin.Context) { CreateTemplateLine(c, client) })
	lines.PUT("/:lineId", func(c *gin.Context) { UpdateTemplateLine(c, client) })
	lines.DELETE("/:lineId", func(c *gin.Context) { DeleteTemplateLine(c, client) })
}

// ─── Template handlers ────────────────────────────────────────────────────────

func ListTemplates(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.ListTemplates(c.Request.Context(), &template.ListTemplatesRequest{
		UserId:          userIDFromCtx(c),
		IncludeArchived: c.Query("archived") == "true",
		TemplateType:    c.Query("type"),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Templates))
	for _, t := range resp.Templates {
		out = append(out, marshalTemplate(t))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "templates": out})
}

func CreateTemplate(c *gin.Context, client template.TemplateServiceClient) {
	var input struct {
		TemplateType   string `json:"template_type" binding:"required"`
		TargetResource string `json:"target_resource" binding:"required"`
		Name           string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateTemplate(c.Request.Context(), &template.CreateTemplateRequest{
		UserId:         userIDFromCtx(c),
		TemplateType:   input.TemplateType,
		TargetResource: input.TargetResource,
		Name:           input.Name,
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "template_id": resp.TemplateId})
}

func GetTemplate(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.GetTemplate(c.Request.Context(), &template.GetTemplateRequest{
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "template": marshalTemplate(resp.Template)})
}

func UpdateTemplate(c *gin.Context, client template.TemplateServiceClient) {
	var input struct {
		Name           string          `json:"name"`
		TargetResource string          `json:"target_resource"`
		Payload        json.RawMessage `json:"payload"`
		PayloadVersion int32           `json:"payload_version"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	payloadStr := ""
	if len(input.Payload) > 0 {
		payloadStr = string(input.Payload)
	}
	resp, err := client.UpdateTemplate(c.Request.Context(), &template.UpdateTemplateRequest{
		TemplateId:     c.Param("id"),
		UserId:         userIDFromCtx(c),
		Name:           input.Name,
		TargetResource: input.TargetResource,
		Payload:        payloadStr,
		PayloadVersion: input.PayloadVersion,
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteTemplate(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.DeleteTemplate(c.Request.Context(), &template.DeleteTemplateRequest{
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ArchiveTemplate(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.ArchiveTemplate(c.Request.Context(), &template.ArchiveTemplateRequest{
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func RestoreTemplate(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.RestoreTemplate(c.Request.Context(), &template.RestoreTemplateRequest{
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Line handlers ────────────────────────────────────────────────────────────

func ListTemplateLines(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.ListTemplateLines(c.Request.Context(), &template.ListTemplateLinesRequest{
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Lines))
	for _, l := range resp.Lines {
		out = append(out, marshalTemplateLine(l))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "lines": out})
}

type templateLineInput struct {
	Type      string          `json:"type" binding:"required"`
	Name      string          `json:"name"`
	Quantity  string          `json:"quantity" binding:"required"`
	Unit      string          `json:"unit"`
	UnitPrice int64           `json:"unit_price"`
	Data      json.RawMessage `json:"data"`
	Position  int32           `json:"position"`
	TaxID     int32           `json:"tax_id"`
}

func (in templateLineInput) dataString() string {
	if len(in.Data) == 0 {
		return ""
	}
	return string(in.Data)
}

func CreateTemplateLine(c *gin.Context, client template.TemplateServiceClient) {
	var input templateLineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateTemplateLine(c.Request.Context(), &template.CreateTemplateLineRequest{
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
		Type:       input.Type,
		Name:       input.Name,
		Quantity:   input.Quantity,
		Unit:       input.Unit,
		UnitPrice:  input.UnitPrice,
		Data:       input.dataString(),
		Position:   input.Position,
		TaxId:      input.TaxID,
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "line_id": resp.LineId})
}

func UpdateTemplateLine(c *gin.Context, client template.TemplateServiceClient) {
	var input struct {
		Type      string          `json:"type" binding:"required"`
		Name      string          `json:"name"`
		Quantity  string          `json:"quantity" binding:"required"`
		Unit      string          `json:"unit"`
		UnitPrice int64           `json:"unit_price" binding:"required"`
		Data      json.RawMessage `json:"data"`
		Position  int32           `json:"position" binding:"required"`
		TaxID     int32           `json:"tax_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	dataStr := ""
	if len(input.Data) > 0 {
		dataStr = string(input.Data)
	}
	resp, err := client.UpdateTemplateLine(c.Request.Context(), &template.UpdateTemplateLineRequest{
		LineId:     c.Param("lineId"),
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
		Type:       input.Type,
		Name:       input.Name,
		Quantity:   input.Quantity,
		Unit:       input.Unit,
		UnitPrice:  input.UnitPrice,
		Data:       dataStr,
		Position:   input.Position,
		TaxId:      input.TaxID,
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteTemplateLine(c *gin.Context, client template.TemplateServiceClient) {
	resp, err := client.DeleteTemplateLine(c.Request.Context(), &template.DeleteTemplateLineRequest{
		LineId:     c.Param("lineId"),
		TemplateId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		templateErrors.unavailable(c)
		return
	}
	if !resp.Success {
		templateErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Marshal helpers ──────────────────────────────────────────────────────────

func marshalTemplate(t *template.Template) gin.H {
	if t == nil {
		return nil
	}
	var archivedAt *string
	if t.ArchivedAt != "" {
		archivedAt = &t.ArchivedAt
	}
	var payload json.RawMessage
	if t.Payload != "" {
		payload = json.RawMessage(t.Payload)
	} else {
		payload = json.RawMessage("{}")
	}
	return gin.H{
		"template_id":     t.TemplateId,
		"user_id":         t.UserId,
		"template_type":   t.TemplateType,
		"target_resource": t.TargetResource,
		"name":            t.Name,
		"archived_at":     archivedAt,
		"payload_version": t.PayloadVersion,
		"payload":         payload,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}
}

func marshalTemplateLine(l *template.TemplateLine) gin.H {
	if l == nil {
		return nil
	}
	var unit *string
	if l.Unit != "" {
		unit = &l.Unit
	}
	out := gin.H{
		"line_id":     l.LineId,
		"template_id": l.TemplateId,
		"type":        l.Type,
		"name":        l.Name,
		"quantity":    l.Quantity,
		"unit":        unit,
		"unit_price":  l.UnitPrice,
		"position":    l.Position,
		"tax_id":      nullableInt(l.TaxId),
	}
	if l.Data == "" {
		out["data"] = json.RawMessage("{}")
	} else {
		out["data"] = json.RawMessage(l.Data)
	}
	return out
}
