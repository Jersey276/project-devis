package sqlutil

import (
	"fmt"
	"strconv"

	usersGrpc "project-devis-users/services/grpc"
)

func Required(field string) *usersGrpc.ValidationError {
	return &usersGrpc.ValidationError{Field: field, Message: "Champ requis."}
}

func Invalid(field, message string) *usersGrpc.ValidationError {
	return &usersGrpc.ValidationError{Field: field, Message: message}
}

func ValidateRate(rate string) error {
	v, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return fmt.Errorf("invalid rate format")
	}
	if v < 0 || v > 999.99 {
		return fmt.Errorf("rate out of range")
	}
	return nil
}
