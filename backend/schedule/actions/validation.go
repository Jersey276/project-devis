package actions

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

const (
	StatusDraft     = "DRAFT"
	StatusNegotiate = "NEGOCIATE"
	StatusDenied    = "DENIED"
	StatusValid     = "VALID"
)

var startMonthRegexp = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

func Required(field string) *scheduleGrpc.ValidationError {
	return &scheduleGrpc.ValidationError{Field: field, Message: "Champ requis."}
}

func Invalid(field, message string) *scheduleGrpc.ValidationError {
	return &scheduleGrpc.ValidationError{Field: field, Message: message}
}

func IsEditableStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case StatusDraft, StatusNegotiate:
		return true
	default:
		return false
	}
}

func ValidateCreateScheduleInput(userID, quoteID, scheduleName, startMonth string, durationMonths int32) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(quoteID) == "" {
		return errors.New("quote_id is required")
	}
	if strings.TrimSpace(scheduleName) == "" {
		return errors.New("name is required")
	}
	if !startMonthRegexp.MatchString(strings.TrimSpace(startMonth)) {
		return errors.New("start_month must match YYYY-MM")
	}
	if durationMonths <= 0 {
		return errors.New("duration_months must be greater than zero")
	}
	return nil
}

func ParseAmountEuros(input string) (int64, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, errors.New("amount is required")
	}
	if strings.HasPrefix(trimmed, "-") {
		return 0, errors.New("negative amounts are not allowed")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return 0, errors.New("invalid amount")
	}
	if len(parts) == 2 && len(parts[1]) > 2 {
		return 0, errors.New("amount must have at most 2 decimals")
	}

	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, errors.New("invalid amount")
	}
	return int64(value*100 + 0.5), nil
}
