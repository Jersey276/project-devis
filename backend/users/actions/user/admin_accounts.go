package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/lib/pq"
)

func scanAccessInfo(row *sql.Row) (*usersGrpc.GetUserAccessInfoResponse, error) {
	var info usersGrpc.GetUserAccessInfoResponse
	if err := row.Scan(&info.UserId, &info.Email, &info.Role, &info.Suspended); err != nil {
		return nil, err
	}
	return &info, nil
}

func getAccessInfo(ctx context.Context, db *sql.DB, query string, arg any) (*usersGrpc.GetUserAccessInfoResponse, error) {
	row := db.QueryRowContext(ctx, query, arg)
	info, err := scanAccessInfo(row)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetUserAccessInfoResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetUserAccessInfoResponse{Success: false, Code: codes.InternalError}, err
	}
	info.Success = true
	info.Code = codes.Success
	return info, nil
}

func GetUserAccessInfo(ctx context.Context, db *sql.DB, req *usersGrpc.GetUserAccessInfoRequest) (*usersGrpc.GetUserAccessInfoResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.GetUserAccessInfoResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	return getAccessInfo(ctx, db, `SELECT user_id, email, role, suspended FROM users WHERE user_id = $1`, req.UserId)
}

func GetUserAccessInfoByEmail(ctx context.Context, db *sql.DB, req *usersGrpc.GetUserAccessInfoByEmailRequest) (*usersGrpc.GetUserAccessInfoResponse, error) {
	if strings.TrimSpace(req.Email) == "" {
		return &usersGrpc.GetUserAccessInfoResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	return getAccessInfo(ctx, db, `SELECT user_id, email, role, suspended FROM users WHERE email = $1`, strings.TrimSpace(req.Email))
}

func ListAdminAccounts(ctx context.Context, db *sql.DB, req *usersGrpc.ListAdminAccountsRequest) (*usersGrpc.ListAdminAccountsResponse, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if userID := strings.TrimSpace(req.UserId); userID != "" {
		conditions = append(conditions, `user_id = $`+fmt.Sprintf("%d", argIdx))
		args = append(args, userID)
		argIdx++
	}

	if search := strings.TrimSpace(req.Search); search != "" {
		conditions = append(conditions, `(
			LOWER(COALESCE(first_name, '')) LIKE $`+fmt.Sprintf("%d", argIdx)+`
			OR LOWER(COALESCE(last_name, '')) LIKE $`+fmt.Sprintf("%d", argIdx)+`
			OR LOWER(email) LIKE $`+fmt.Sprintf("%d", argIdx)+`
		)`)
		args = append(args, "%"+strings.ToLower(search)+"%")
		argIdx++
	}

	if len(req.Roles) > 0 {
		conditions = append(conditions, `role = ANY($`+fmt.Sprintf("%d", argIdx)+`)`)
		args = append(args, pq.Array(req.Roles))
		argIdx++
	}

	if len(req.Statuses) > 0 {
		var statusConds []string
		for _, s := range req.Statuses {
			switch s {
			case "active":
				statusConds = append(statusConds, "suspended = FALSE")
			case "suspended":
				statusConds = append(statusConds, "suspended = TRUE")
			}
		}
		if len(statusConds) > 0 {
			conditions = append(conditions, "("+strings.Join(statusConds, " OR ")+")")
		}
	}

	query := `
		SELECT
			user_id,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			email,
			role,
			COALESCE(plan, ''),
			last_login_at,
			suspended,
			COALESCE(phone, ''),
			COALESCE(company, ''),
			COALESCE(siren, ''),
			COALESCE(vat, ''),
			created_at
		FROM users`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += `
		ORDER BY last_name ASC, first_name ASC, email ASC,
			CASE WHEN role = 'admin' THEN 0 ELSE 1 END ASC,
			user_id ASC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return &usersGrpc.ListAdminAccountsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	accounts := make([]*usersGrpc.AdminAccount, 0)
	for rows.Next() {
		var account usersGrpc.AdminAccount
		var lastLoginAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(
			&account.UserId,
			&account.FirstName,
			&account.LastName,
			&account.Email,
			&account.Role,
			&account.Plan,
			&lastLoginAt,
			&account.Suspended,
			&account.Phone,
			&account.Company,
			&account.Siren,
			&account.Vat,
			&createdAt,
		); err != nil {
			return &usersGrpc.ListAdminAccountsResponse{Success: false, Code: codes.InternalError}, err
		}
		if lastLoginAt.Valid {
			account.LastLoginAt = lastLoginAt.Time.UTC().Format(time.RFC3339)
		}
		account.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		accounts = append(accounts, &account)
	}
	if err := rows.Err(); err != nil {
		return &usersGrpc.ListAdminAccountsResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.ListAdminAccountsResponse{
		Success:  true,
		Code:     codes.Success,
		Accounts: accounts,
	}, nil
}

func UpdateAdminAccount(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateAdminAccountRequest) (*usersGrpc.GenericResponse, error) {
	if req.UserId == "" || strings.TrimSpace(req.Email) == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	role := strings.TrimSpace(req.Role)
	if role != "user" && role != "admin" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx, `
		UPDATE users
		SET
			first_name = $1,
			last_name = $2,
			email = $3,
			role = $4,
			plan = $5,
			phone = $6,
			company = $7,
			siren = $8,
			vat = $9,
			updated_at = NOW()
		WHERE user_id = $10
	`,
		sqlutil.NullableStr(strings.TrimSpace(req.FirstName)),
		sqlutil.NullableStr(strings.TrimSpace(req.LastName)),
		strings.TrimSpace(req.Email),
		role,
		sqlutil.NullableStr(strings.TrimSpace(req.Plan)),
		sqlutil.NullableStr(strings.TrimSpace(req.Phone)),
		sqlutil.NullableStr(strings.TrimSpace(req.Company)),
		sqlutil.NullableStr(strings.TrimSpace(req.Siren)),
		sqlutil.NullableStr(strings.TrimSpace(req.Vat)),
		req.UserId,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return &usersGrpc.GenericResponse{Success: false, Code: codes.AlreadyExists}, nil
			}
			if pqErr.Code == "23514" {
				return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
			}
		}
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}

func SuspendAdminAccount(ctx context.Context, db *sql.DB, req *usersGrpc.SuspendAdminAccountRequest) (*usersGrpc.GenericResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx, `
		UPDATE users
		SET suspended = TRUE, updated_at = NOW()
		WHERE user_id = $1
	`, req.UserId)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}

func TouchUserLastLogin(ctx context.Context, db *sql.DB, req *usersGrpc.TouchUserLastLoginRequest) (*usersGrpc.GenericResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx, `
		UPDATE users
		SET last_login_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
	`, req.UserId)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
