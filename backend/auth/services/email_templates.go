package services

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed templates/password_reset.html
var passwordResetTmpl string

//go:embed templates/email_verification.html
var emailVerificationTmpl string

type passwordResetData struct {
	ResetURL string
}

type emailVerificationData struct {
	VerifyURL string
}

func renderPasswordReset(resetURL string) (string, error) {
	return renderTmpl(passwordResetTmpl, passwordResetData{ResetURL: resetURL})
}

func renderEmailVerification(verifyURL string) (string, error) {
	return renderTmpl(emailVerificationTmpl, emailVerificationData{VerifyURL: verifyURL})
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
