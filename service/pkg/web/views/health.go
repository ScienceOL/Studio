package views

import "github.com/gin-gonic/gin"

// @Summary 健康检查
// @Description 检查服务运行状态
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "返回成功状态"
// @Router /api/health [get]
func Health(g *gin.Context) {
	g.JSON(200, map[string]any{
		"success": "ok",
	})
}
