package client

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateClientRequest) (*usersGrpc.CreateClientResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if req.UserId == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if req.FirstName == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "first_name", Message: "Champ requis."})
	}
	if req.LastName == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "last_name", Message: "Champ requis."})
	}
	if msg := sqlutil.ValidateSIRET(req.Siret, req.Siren); msg != "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "siret", Message: msg})
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.CreateClientResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	clientID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO clients (client_id, user_id, first_name, last_name, email, phone, company, siren, vat, siret, client_type)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		clientID, req.UserId, req.FirstName, req.LastName,
		sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
		sqlutil.NullableStr(req.Company), sqlutil.NullableStr(req.Siren), sqlutil.NullableStr(req.Vat),
		sqlutil.NullableStr(sqlutil.NormalizeSIRET(req.Siret)),
		sqlutil.ClientTypeToDBString(req.ClientType),
	)
	if err != nil {
		return &usersGrpc.CreateClientResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateClientResponse{Success: true, Code: codes.Success, ClientId: clientID}, nil
}
