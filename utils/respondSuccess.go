package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RespondSuccess(c *gin.Context, data interface{}, pagination *Pagination) {
	// 构建基础的响应数据
	response := Response{
		Status:  http.StatusOK,
		Message: "Request succeeded",
	}

	// 如果传递了 data，则直接将其添加到响应中，而不做额外处理
	if data != nil {
		response.Data = data
	}

	// 如果传递了 pagination 参数，则添加分页信息
	if pagination != nil {
		if response.Data == nil {
			response.Data = gin.H{} // 如果没有 data，就初始化为空的对象
		}
		dataMap := response.Data.(gin.H)
		dataMap["page"] = pagination.Page
		dataMap["page_size"] = pagination.PageSize
		dataMap["total"] = pagination.Total
	}

	// 返回响应
	c.JSON(http.StatusOK, response)
}
