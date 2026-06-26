package actions

// Route response codes returned in FormGenericResponse.code and GenericResponse.code.
const (
	CodeSuccess                  int32 = 0
	CodeUserAlreadyExists        int32 = 1001
	CodeUserNotFound             int32 = 1002
	CodeInvalidCredentials       int32 = 1003
	CodeInvalidRefreshToken      int32 = 1004
	CodeInvalidResetToken        int32 = 1005
	CodeExpiredResetToken        int32 = 1006
	CodeWeakPassword             int32 = 1007
	CodeSessionInvalidated       int32 = 1008
	CodeInvalidInput             int32 = 1009
	CodeInvalidVerificationToken int32 = 1010
	CodeExpiredVerificationToken int32 = 1011
	CodeAlreadyVerified          int32 = 1012
	CodeOAuthEmailNotVerified    int32 = 1013
	CodeOAuthIdentityTaken       int32 = 1014
	CodeLastLoginMethod          int32 = 1015
	CodeInvalidInvitationToken   int32 = 1016
	CodeExpiredInvitationToken   int32 = 1017
	CodeClientAlreadyLinked      int32 = 1018
	CodeInvalidEmailChangeToken  int32 = 1019
	CodeExpiredEmailChangeToken  int32 = 1020
	CodeEmailAlreadyInUse        int32 = 1021
	CodeUserServiceError         int32 = 2001
	CodeInternalError            int32 = 2002
	CodeNotImplemented           int32 = 2003
)

// Field validation codes used inside FormFieldError.error_code.
// These are independent from route response codes and describe
// why a specific field failed validation.
const (
	FieldErrRequired      int32 = 1
	FieldErrInvalidFormat int32 = 2
	FieldErrTooShort      int32 = 3
	FieldErrAlreadyInUse  int32 = 4
)
