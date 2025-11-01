package foo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

// HandleHelloWorld godoc
// @Summary      Anonymous Test Endpoint
// @Description  A simple test endpoint that does not require authentication
// @Tags         Test
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /foo [get]
func HandleTestAnomy(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "This is an anonymous test endpoint",
	})
}

// @Summary      Authenticated Test Endpoint
// @Description  A simple hello world endpoint that requires Bearer token authentication
// @Tags         Test
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /foo/auth [get]
func HandleTestAuth(ctx *gin.Context) {
	// 使用GetCurrentUser获取当前用户
	user := auth.GetCurrentUser(ctx)

	logger.Infof(ctx.Request.Context(), "User %v is authenticated", user)

	// 用户已登录，返回欢迎信息
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
		"user":    user,
	})
}
