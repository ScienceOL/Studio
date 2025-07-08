package foo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

// HandleHelloWorld 是一个简单的路由处理函数，检查用户是否已登录
func HandleHelloWorld() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用GetCurrentUser获取当前用户
		user, exists := auth.GetCurrentUser(c)
		if !exists {
			// 由于已经使用了RequireAuth中间件，这种情况正常不会发生
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Not authenticated",
			})
			return
		}

		logger.Infof(c.Request.Context(), "User %v is authenticated", user)

		// 用户已登录，返回欢迎信息
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
			"user":    user,
		})
	}
}
