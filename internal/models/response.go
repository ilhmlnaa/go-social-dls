package models

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, err error) ErrorResponse {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return ErrorResponse{
		Success: false,
		Message: message,
		Error:   errMsg,
	}
}
