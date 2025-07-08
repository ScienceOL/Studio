package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"golang.org/x/oauth2"
)

// 用于认证的错误
var (
	ErrAuthHeaderMissing = errors.New("authorization header is required")
	ErrInvalidAuthFormat = errors.New("invalid authorization header format")
	ErrInvalidToken      = errors.New("invalid or expired token")
)

// ValidateToken 检查令牌是否有效
func ValidateToken(token string, c *gin.Context) (bool, map[string]any, error) {
	// 获取OAuth2配置
	oauthConfig := GetOAuthConfig()
	if oauthConfig == nil {
		logger.Errorf(c.Request.Context(), "OAuth2 configuration not available")
		return false, nil, errors.New("configuration not available")
	}

	// 创建一个包含传入token的oauth2.Token对象
	oauthToken := &oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	}

	// 创建context
	ctx := context.Background()

	// 使用token构建OAuth2客户端
	client := oauthConfig.Client(ctx, oauthToken)

	// 获取配置中的用户信息URL
	config := webapp.Config()

	// 获取用户信息 - 如果token有效，这个请求将成功
	resp, err := client.Get(config.OAuth2.UserInfoURL)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get user info: %v", err)
		return false, nil, ErrInvalidToken
	}
	logger.Infof(c.Request.Context(), "Response status: %d", resp.StatusCode)
	defer resp.Body.Close()

	// 如果状态码不是2xx，则认为token无效
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Errorf(c.Request.Context(), "Invalid token, status code: %d", resp.StatusCode)
		return false, nil, ErrInvalidToken
	}

	// 解析用户信息
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to parse user info: %v", err)
		return false, nil, err
	}

	// 检查API调用是否成功
	if status, ok := result["status"].(string); ok && status == "ok" {
		// 提取用户数据
		userData, ok := result["data"].(map[string]interface{})
		if !ok {
			logger.Errorf(c.Request.Context(), "Invalid user data format, result: %+v", result)
			return false, nil, errors.New("invalid user data format")
		}

		// 返回验证成功和用户数据
		return true, map[string]any{
			"valid": true,
			"user":  userData,
		}, nil
	} else {
		// 如果没有status字段或者status不是ok，直接将result作为用户数据
		// 验证返回的数据包含必要的用户信息
		if result["id"] == nil {
			logger.Errorf(c.Request.Context(), "Invalid user data format: missing required fields")
			return false, nil, errors.New("invalid user data format")
		}

		// 返回验证成功和用户数据
		return true, map[string]any{
			"valid": true,
			"user":  result, // 直接使用返回的用户对象
		}, nil
	}
}

// RequireAuth 中间件函数验证用户是否已登录
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrAuthHeaderMissing.Error()})
			c.Abort()
			return
		}

		// 检查格式是否为 "Bearer {token}"
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidAuthFormat.Error()})
			c.Abort()
			return
		}

		// 验证令牌
		token := bearerToken[1]
		valid, userInfo, err := ValidateToken(token, c)
		if err != nil || !valid {
			logger.Errorf(c.Request.Context(), "Token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("user", userInfo["user"])

		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(c *gin.Context) (map[string]any, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	return user.(map[string]any), true
}
