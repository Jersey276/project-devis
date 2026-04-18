package actions

// Route response codes returned in FormGenericResponse.code and GenericResponse.code.
const (
	CodeSuccess             int32 = 0
	CodeUserAlreadyExists   int32 = 1001
	CodeUserNotFound        int32 = 1002
	CodeInvalidCredentials  int32 = 1003
	CodeInvalidRefreshToken int32 = 1004
	CodeUserServiceError    int32 = 2001
	CodeInternalError       int32 = 2002
	CodeNotImplemented      int32 = 2003
)

// Field validation codes used inside FormFieldError.error_code.
// These are independent from route response codes and describe
// why a specific field failed validation.
const (
	FieldErrRequired       int32 = 1
	FieldErrInvalidFormat  int32 = 2
	FieldErrTooShort       int32 = 3
	FieldErrAlreadyInUse   int32 = 4
)
