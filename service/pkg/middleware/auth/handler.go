package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"golang.org/x/oauth2"
)

// HandleLogin 处理登录请求
func HandleLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取OAuth2配置
		oauthConfig := GetOAuthConfig()
		if oauthConfig == nil {
			logger.Errorf(c.Request.Context(), "OAuth2 configuration not available")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Configuration not available"})
			return
		}

		// 生成随机state用于防止CSRF攻击
		state := fmt.Sprintf("%d", time.Now().UnixNano())

		// 将state保存到Redis中，设置5分钟过期时间
		stateKey := fmt.Sprintf("oauth_state:%s", state)
		if err := redis.GetClient().Set(c.Request.Context(), stateKey, "valid", 5*time.Minute).Err(); err != nil {
			logger.Errorf(c.Request.Context(), "Failed to save state to Redis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 构建授权URL并重定向用户到OAuth2提供商登录页面
		authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusFound, authURL)
	}
}

// HandleCallback 处理OAuth2回调
func HandleCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取OAuth2配置
		oauthConfig := GetOAuthConfig()
		if oauthConfig == nil {
			logger.Errorf(c.Request.Context(), "OAuth2 configuration not available")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Configuration not available"})
			return
		}

		// 获取授权码和state
		code := c.Query("code")
		state := c.Query("state")

		// 验证state是否存在于Redis中
		stateKey := fmt.Sprintf("oauth_state:%s", state)
		redisResult := redis.GetClient().Get(c.Request.Context(), stateKey)
		if redisResult.Err() != nil {
			logger.Errorf(c.Request.Context(), "State validation failed, state not found or expired: %s, error: %v", state, redisResult.Err())
			c.JSON(http.StatusBadRequest, gin.H{"error": "State validation failed"})
			return
		}

		// 删除使用过的state
		redis.GetClient().Del(c.Request.Context(), stateKey)

		ctx := context.Background()

		// 用授权码交换token
		token, err := oauthConfig.Exchange(ctx, code, oauth2.AccessTypeOffline)
		if err != nil {
			logger.Errorf(c.Request.Context(), "Token exchange failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to exchange token"})
			return
		}

		// 检查是否收到刷新令牌
		if token.RefreshToken == "" {
			logger.Warnf(c.Request.Context(), "No refresh token received from Casdoor")
		} else {
			logger.Infof(c.Request.Context(), "Successfully received refresh token from Casdoor")
		}

		// 使用token构建OAuth2客户端
		client := oauthConfig.Client(ctx, token)

		// 获取用户信息
		config := webapp.Config()
		resp, err := client.Get(config.OAuth2.UserInfoURL)
		if err != nil {
			logger.Errorf(c.Request.Context(), "Failed to get user info: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
			return
		}
		defer resp.Body.Close()

		// 解析用户信息
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			logger.Errorf(c.Request.Context(), "Failed to parse user info: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
			return
		}

		// 检查API调用是否成功
		if status, ok := result["status"].(string); !ok || status != "ok" {
			logger.Errorf(c.Request.Context(), "Failed to get valid user info, result: %+v", result)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get valid user info"})
			return
		}

		// 提取用户数据
		userData, ok := result["data"].(map[string]interface{})
		if !ok {
			logger.Errorf(c.Request.Context(), "Invalid user data format, result: %+v", result)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data format"})
			return
		}

		// 将用户信息和token返回给前端
		c.JSON(http.StatusOK, gin.H{
			"user":          userData,
			"token":         token.AccessToken,
			"refresh_token": token.RefreshToken,
			"expires_in":    token.Expiry.Unix() - time.Now().Unix(),
			"token_type":    token.TokenType,
		})
	}
}

// HandleRefresh 处理令牌刷新请求
func HandleRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取OAuth2配置
		oauthConfig := GetOAuthConfig()
		if oauthConfig == nil {
			logger.Errorf(c.Request.Context(), "OAuth2 configuration not available")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Configuration not available"})
			return
		}

		// 从请求中获取刷新令牌
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Errorf(c.Request.Context(), "Invalid request format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// 创建一个已过期的令牌对象，但包含有效的刷新令牌
		expiredToken := &oauth2.Token{
			RefreshToken: req.RefreshToken,
			Expiry:       time.Now().Add(-1 * time.Hour), // 确保令牌已过期
		}

		ctx := context.Background()

		// 使用TokenSource刷新令牌
		tokenSource := oauthConfig.TokenSource(ctx, expiredToken)
		newToken, err := tokenSource.Token()
		if err != nil {
			logger.Errorf(c.Request.Context(), "Failed to refresh token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token"})
			return
		}

		// 返回新的令牌
		c.JSON(http.StatusOK, gin.H{
			"access_token":  newToken.AccessToken,
			"refresh_token": newToken.RefreshToken,
			"expires_in":    newToken.Expiry.Unix() - time.Now().Unix(),
			"token_type":    newToken.TokenType,
		})
	}
}

// RefreshTokenRequest 定义刷新令牌请求的结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
