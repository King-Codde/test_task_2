package model

type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

type ValidationError struct {
	Field   string
	Message string
}

func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
	}
}

func (vr *ValidationResult) AddError(field, message string) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

func (vr *ValidationResult) GetErrorMessage() string {
	if vr.Valid {
		return ""
	}
	msg := ""
	for i, err := range vr.Errors {
		if i > 0 {
			msg += "\n"
		}
		msg += err.Field + ": " + err.Message
	}
	return msg
}

type AppError struct {
	Type        ErrorType
	Message     string
	Cause       error
	UserMessage string
}

type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeConnection
	ErrorTypeAuthentication
	ErrorTypeDecoding
	ErrorTypeStream
	ErrorTypeUnknown
)

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func NewAppError(errType ErrorType, message string, cause error, userMessage string) *AppError {
	return &AppError{
		Type:        errType,
		Message:     message,
		Cause:       cause,
		UserMessage: userMessage,
	}
}
