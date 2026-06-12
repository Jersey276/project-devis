# Error Codes

This document is the authoritative reference for all numeric codes used across the auth service and gateway.

There are **two independent code systems** — do not mix them:

- **Route response codes** — carried in `FormGenericResponse.code` or `GenericResponse.code`. Describe the overall outcome of an RPC call.
- **Field validation codes** — carried in `FormFieldError.error_code[]`. Describe why a specific form field failed validation. Multiple codes can appear on the same field.

---

## Route Response Codes

Defined in `backend/auth/actions/errors.go`.
Used by: auth service, forwarded as-is by the gateway.

| Code   | Constant                  | Description                                                 |
| ------ | ------------------------- | ----------------------------------------------------------- |
| `0`    | `CodeSuccess`             | Request completed successfully.                             |
| `1001` | `CodeUserAlreadyExists`   | A user with this email address is already registered.       |
| `1002` | `CodeUserNotFound`        | No user found matching the provided identifier.             |
| `1003` | `CodeInvalidCredentials`  | Email or password is incorrect.                             |
| `1004` | `CodeInvalidRefreshToken` | Refresh token is invalid or expired.                        |
| `1005` | `CodeInvalidResetToken`   | Password reset token is invalid, missing, or already used.  |
| `1006` | `CodeExpiredResetToken`   | Password reset token exists but has expired.                |
| `1007` | `CodeWeakPassword`        | Password does not satisfy security policy requirements.     |
| `1008` | `CodeSessionInvalidated`  | Access token session version is stale and has been revoked. |
| `2001` | `CodeUserServiceError`    | The user micro-service returned an error or is unreachable. |
| `2002` | `CodeInternalError`       | Unexpected internal error in the auth service.              |
| `2003` | `CodeNotImplemented`      | The requested feature is not yet implemented.               |

### Gateway HTTP mapping

The gateway maps route response codes to HTTP status codes as follows:

| Code   | HTTP Status               |
| ------ | ------------------------- |
| `1001` | 409 Conflict              |
| `1002` | 404 Not Found             |
| `1003` | 401 Unauthorized          |
| `1004` | 401 Unauthorized          |
| `1005` | 400 Bad Request           |
| `1006` | 410 Gone                  |
| `1007` | 422 Unprocessable Entity  |
| `1008` | 401 Unauthorized          |
| `2001` | 502 Bad Gateway           |
| `2002` | 500 Internal Server Error |
| `2003` | 501 Not Implemented       |

---

## Field Validation Codes

Defined in `backend/auth/actions/errors.go`.
Used inside `FormFieldError.error_code[]` on routes that return `FormGenericResponse`.
These codes are field-agnostic — the same code can appear on any field.

| Code | Constant                | Description                                                                      |
| ---- | ----------------------- | -------------------------------------------------------------------------------- |
| `1`  | `FieldErrRequired`      | Field is missing or empty.                                                       |
| `2`  | `FieldErrInvalidFormat` | Field value does not match the expected format (e.g. invalid email address).     |
| `3`  | `FieldErrTooShort`      | Field value is shorter than the minimum allowed length.                          |
| `4`  | `FieldErrAlreadyInUse`  | Field value is already taken by another account (e.g. email already registered). |

### Response shape when field errors are present

```json
{
  "success": false,
  "code": 1001,
  "field_errors": [
    {
      "field": "email",
      "error_code": [4]
    }
  ]
}
```

The gateway returns HTTP **422 Unprocessable Entity** when `field_errors` is non-empty.

---

## Routes that use FormGenericResponse

- auth: `Register` / `POST /auth/register` (field validation by auth service)
