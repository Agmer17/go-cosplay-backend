package shared

type SuccessResponse struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Code  int `json:"-"`
	Error any `json:"error"`
}

func NewErrorResponse(code int, err any) *ErrorResponse {
	return &ErrorResponse{
		Code:  code,
		Error: err,
	}
}

func NewSuccessResponse(code int, message string, data any) SuccessResponse {

	return SuccessResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
