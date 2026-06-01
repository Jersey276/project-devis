package actions

import "project-devis-schedule/actions/codes"

const (
	CodeSuccess            = codes.Success
	CodeNotFound           = codes.NotFound
	CodeAlreadyExists      = codes.AlreadyExists
	CodeInvalidInput       = codes.InvalidInput
	CodeScheduleFinalized  = codes.ScheduleFinalized
	CodeScheduleUnbalanced = codes.ScheduleUnbalanced
	CodeScheduleValidated  = codes.ScheduleValidated
	CodeInternalError      = codes.InternalError
)