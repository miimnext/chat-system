package models

// Response 是一个通用的响应结构
type Response struct {
	Code    int         `json:"code"`    // 响应的状态码
	Message string      `json:"message"` // 响应的信息
	Data    interface{} `json:"data"`    // 响应的数据
}

// NewSuccessResponse 创建一个成功的响应
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "OK",
		Data:    data,
	}
}

// NewErrorResponse 创建一个错误的响应
func NewErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
