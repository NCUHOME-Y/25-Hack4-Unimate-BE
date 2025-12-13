package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 标准错误响应结构
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    int                    `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// 标准成功响应结构
type SuccessResponse struct {
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// RespondError 统一的错误响应
func RespondError(c *gin.Context, statusCode int, message string, logFields logrus.Fields) {
	LogError(message, logFields)
	c.JSON(statusCode, gin.H{"error": message})
}

// RespondValidationError 验证错误响应（400）
func RespondValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// RespondUnauthorized 未授权响应（401）
func RespondUnauthorized(c *gin.Context, message string, logFields logrus.Fields) {
	LogWarn(message, logFields)
	c.JSON(http.StatusUnauthorized, gin.H{"error": message})
}

// RespondNotFound 未找到响应（404）
func RespondNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

// RespondConflict 资源冲突响应（409）
func RespondConflict(c *gin.Context, message string, logFields logrus.Fields) {
	LogInfo(message, logFields)
	c.JSON(http.StatusConflict, gin.H{"error": message})
}

// RespondInternalError 服务器内部错误响应（500）
func RespondInternalError(c *gin.Context, message string, logFields logrus.Fields) {
	LogError(message, logFields)
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

// RespondSuccess 成功响应（200）
func RespondSuccess(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, data)
}

// RespondCreated 资源创建成功响应（201）
func RespondCreated(c *gin.Context, data gin.H) {
	c.JSON(http.StatusCreated, data)
}

// HandleDatabaseError 处理数据库错误并返回适当的HTTP响应
func HandleDatabaseError(c *gin.Context, err error, operation string, logFields logrus.Fields) {
	if logFields == nil {
		logFields = logrus.Fields{}
	}
	logFields["operation"] = operation
	logFields["error"] = err.Error()

	RespondInternalError(c, "数据库操作失败，请稍后重试", logFields)
}
