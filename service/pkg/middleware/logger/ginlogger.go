package logger

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxLength = 40 * 1024
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogWithWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为不适合缓存响应体的请求类型
		if shouldSkipResponseBodyLogging(c.Request) {
			logConnectionEstablished(c)
			c.Next()
			return
		}

		// 开始时间
		startTime := time.Now()
		if c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			// 请求体
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = c.GetRawData()
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				safeBody := make([]byte, 0)
				if len(bodyBytes) > 0 {
					// 限制记录大小
					if len(bodyBytes) > maxLength {
						safeBody = bodyBytes[:maxLength]
					} else {
						safeBody = bodyBytes
					}

					// 敏感信息过滤
					safeBody = filterSensitiveData(safeBody)
				}

				Infof(c, "request url: %s request body: %s", c.Request.RequestURI, string(safeBody))
			}
		}
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// http版本
		httpVersion := c.Request.Proto

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.Request.Host

		Infof(c, "response-info: %3d | %13v | %15s | %s | %s | %s| %s",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
			blw.body.String(),
			httpVersion,
		)
	}
}

func shouldSkipResponseBodyLogging(req *http.Request) bool {
	// WebSocket 连接
	if isWebSocketRequest(req) {
		return true
	}

	// SSE 连接
	if isSSERequest(req) {
		return true
	}

	// HTTP CONNECT 方法
	if req.Method == http.MethodConnect {
		return true
	}

	// gRPC 请求
	if strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
		return true
	}

	// 文件上传/下载（基于 Content-Type）
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") ||
		strings.Contains(contentType, "application/octet-stream") {
		return true
	}

	// 排除 swagger 路由
	if strings.HasPrefix(req.URL.Path, "/api/swagger/") {
		return true
	}

	return false
}

// 检查是否为 SSE 请求
func isSSERequest(req *http.Request) bool {
	accept := req.Header.Get("Accept")
	return strings.Contains(accept, "text/event-stream")
}

// 检查是否为 WebSocket 请求
func isWebSocketRequest(req *http.Request) bool {
	return strings.ToLower(req.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(req.Header.Get("Upgrade")) == "websocket"
}

func filterSensitiveData(body []byte) []byte {
	// 示例：过滤密码字段
	strBody := string(body)
	if strings.Contains(strBody, `"password"`) {
		// 使用正则表达式进行更健壮的过滤
		re := regexp.MustCompile(`"password":\s*"[^"]*"`)
		strBody = re.ReplaceAllString(strBody, `password": "***FILTERED***`)
	}

	// 添加其他敏感字段过滤（如信用卡号、token等）
	return []byte(strBody)
}

func logConnectionEstablished(c *gin.Context) {
	var connType string
	if isWebSocketRequest(c.Request) {
		connType = "WebSocket"
	} else if isSSERequest(c.Request) {
		connType = "SSE"
	} else {
		connType = "Special"
	}

	Infof(c, "%s connection established: %s %s from %s",
		connType, c.Request.Method, c.Request.RequestURI, c.Request.RemoteAddr)
}
