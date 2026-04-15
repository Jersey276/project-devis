package actions

// Error codes returned by the auth service in gRPC responses.
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
