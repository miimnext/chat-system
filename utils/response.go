package utils

type Response struct {
	Code       int         `json:"code"`                 // 状态码
	Message    string      `json:"message"`              // 提示信息
	Data       interface{} `json:"data"`                 // 数据内容，可以是任何类型
	Pagination *Pagination `json:"pagination,omitempty"` // 分页信息（如果需要）
}
