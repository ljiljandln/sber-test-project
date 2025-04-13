package dto

// @Summary Response
// @Description Стандартная модель ответа сервера на запрос
// @Tags models
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty" swaggertype:"object"`
}

func SuccessResponse(msg string, data any) *Response {
	return &Response{
		Status:  "success",
		Message: msg,
		Data:    data,
	}
}

func ErrorResponse(message string) *Response {
	return &Response{
		Status:  "error",
		Message: message,
	}
}
