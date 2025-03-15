package app

import "errors"

var (
	errInitFailed         = errors.New("init failed")
	errPlanFailed         = errors.New("plan failed")
	errApplyFailed        = errors.New("apply failed")
	errNotFoundPlanFile   = errors.New("plan file is not found")
	errApprovalsRequired  = errors.New("approvals are required")
	errForceUnlockFailed  = errors.New("force unlock failed")
	errImportFailed       = errors.New("import failed")
	errAlreadyLocked      = errors.New("already locked")
	errPanicOccurred      = errors.New("panic occurred")
	errMultipleLockLabels = errors.New("multiple lock labels")
	errInvalidForceUnlock = errors.New("invalid force unlock")
)
