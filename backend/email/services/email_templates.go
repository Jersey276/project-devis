package services

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed templates/quote_sent.html
var quoteSentTmpl string

//go:embed templates/schedule_notification.html
var scheduleNotificationTmpl string

type quoteSentData struct {
	ClientName string
	QuoteName  string
	HasPDF     bool
}

type scheduleNotificationData struct {
	ClientName  string
	QuoteName   string
	StatusLabel string
	StatusIcon  string
	StatusColor string
	StatusBg    string
	Message     string
}

func RenderQuoteSent(clientName, quoteName string, hasPDF bool) (string, error) {
	return renderTmpl(quoteSentTmpl, quoteSentData{
		ClientName: clientName,
		QuoteName:  quoteName,
		HasPDF:     hasPDF,
	})
}

func RenderScheduleNotification(clientName, quoteName, status string) (string, error) {
	data := scheduleNotificationData{
		ClientName: clientName,
		QuoteName:  quoteName,
	}
	switch status {
	case "VALID":
		data.StatusLabel = "validé"
		data.StatusIcon = "✅"
		data.StatusColor = "#15803d"
		data.StatusBg = "#f0fdf4"
		data.Message = "Bonne nouvelle ! L'échéancier de paiement associé à votre devis a été validé."
	case "DENIED":
		data.StatusLabel = "refusé"
		data.StatusIcon = "❌"
		data.StatusColor = "#b91c1c"
		data.StatusBg = "#fef2f2"
		data.Message = "L'échéancier de paiement associé à votre devis a été refusé. N'hésitez pas à nous contacter pour en discuter."
	default:
		data.StatusLabel = "mis à jour"
		data.StatusIcon = "🔄"
		data.StatusColor = "#0369a1"
		data.StatusBg = "#f0f9ff"
		data.Message = "L'échéancier de paiement associé à votre devis a été mis à jour."
	}
	return renderTmpl(scheduleNotificationTmpl, data)
}

func renderTmpl(tmpl string, data any) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
