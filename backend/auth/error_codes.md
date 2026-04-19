# Auth Service — Error Codes

Error codes returned by the auth gRPC service in the `code` field (`int32`) of responses.

The gateway is responsible for mapping these codes to HTTP status codes and user-facing messages.

## Codes

| Code   | Name                    | HTTP | Description                                                              | Used by                      |
|--------|-------------------------|------|--------------------------------------------------------------------------|------------------------------|
| `0`    | SUCCESS                 | 200  | Operation completed successfully.                                        | All RPCs                     |
| `1001` | USER_ALREADY_EXISTS     | 409  | A user with this email already exists.                                   | Register                     |
| `1002` | USER_NOT_FOUND          | 404  | No user found for the given email or user ID.                            | Login, RefreshToken          |
| `1003` | INVALID_CREDENTIALS     | 401  | The email/password combination is incorrect.                             | Login                        |
| `1004` | INVALID_REFRESH_TOKEN   | 401  | The refresh token is invalid, expired, or has already been revoked.      | RefreshToken                 |
| `2001` | USER_SERVICE_ERROR      | 502  | The upstream user service returned an error or failed to create the user.| Register                     |
| `2002` | INTERNAL_ERROR          | 500  | An unexpected internal error occurred.                                   | All RPCs                     |
| `2003` | NOT_IMPLEMENTED         | 501  | This endpoint is not yet implemented.                                    | ResetPassword, UpdatePassword, VerifyEmail |

## Ranges

- `0` — Success
- `1xxx` — Client/business errors
- `2xxx` — Server/infrastructure errors
