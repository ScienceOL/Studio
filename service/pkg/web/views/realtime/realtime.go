package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/scienceol/studio/service/pkg/core/realtime"
)

// Handle holds no state for prototype; uses global realtime.Manager.
type Handle struct{}

func NewHandle() *Handle { return &Handle{} }

// ClientSignal upgrades to websocket for client signaling.
func (h *Handle) ClientSignal(ctx *gin.Context) {
	clientID := ctx.Query("clientId")
	if clientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "clientId required"})
		return
	}
	conn, err := realtime.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	realtime.Manager.RegisterClient(clientID, conn)
	realtime.Manager.HandleClientLoop(clientID, conn)
}

// HostSignal upgrades to websocket for host signaling.
func (h *Handle) HostSignal(ctx *gin.Context) {
	hostID := ctx.Param("hostId")
	if hostID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hostId required"})
		return
	}
	conn, err := realtime.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	realtime.Manager.RegisterHost(hostID, conn)
	realtime.Manager.HandleHostLoop(hostID, conn)
}

// StartCamera triggers stream start on host.
func (h *Handle) StartCamera(ctx *gin.Context) {
	var req struct {
		HostID   string `json:"hostId" binding:"required"`
		CameraID string `json:"cameraId" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := realtime.Manager.StartCamera(req.HostID, req.CameraID); err != nil {
		if err == websocket.ErrBadHandshake {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "send start failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "start sent"})
}

// StopCamera triggers stream stop on host.
func (h *Handle) StopCamera(ctx *gin.Context) {
	var req struct {
		HostID   string `json:"hostId" binding:"required"`
		CameraID string `json:"cameraId" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := realtime.Manager.StopCamera(req.HostID, req.CameraID); err != nil {
		if err == websocket.ErrBadHandshake {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "send stop failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "stop sent"})
}
