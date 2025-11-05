package labstatus

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
)

const (
	ActionQueryList    = "query_list"    // 查询用户所有实验室状态
	ActionQueryDetail  = "query_detail"  // 查询单个实验室状态
	ActionStatusUpdate = "status_update" // 状态更新通知
)

type QueryListReq struct {
	Action  string    `json:"action"`
	MsgUUID uuid.UUID `json:"msg_uuid"`
}

type QueryDetailReq struct {
	Action  string    `json:"action"`
	MsgUUID uuid.UUID `json:"msg_uuid"`
	Data    struct {
		LabUUID uuid.UUID `json:"lab_uuid"`
	} `json:"data"`
}

type LabStatusData struct {
	LabUUID         uuid.UUID  `json:"lab_uuid"`
	IsOnline        bool       `json:"is_online"`
	LastConnectedAt *time.Time `json:"last_connected_at,omitempty"`
}

type Handle struct {
	wsClient     *melody.Melody
	labStore     repo.LaboratoryRepo
	userSessions sync.Map // userID -> []*melody.Session
}

func New() *Handle {
	h := &Handle{
		wsClient:     melody.New(),
		labStore:     eStore.New(),
		userSessions: sync.Map{},
	}
	h.initWebSocket()

	// 注册为全局状态变化处理器
	GetGlobalNotifier().RegisterHandler(h.NotifyStatusChange)

	return h
}

// ConnectLabStatus WebSocket 连接入口
func (h *Handle) ConnectLabStatus(ctx *gin.Context) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		common.ReplyErr(ctx, code.UnLogin)
		return
	}

	if err := h.wsClient.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		"user_id": userInfo.ID,
		"ctx":     ctx,
	}); err != nil {
		logger.Errorf(ctx, "ConnectLabStatus HandleRequestWithKeys err: %+v", err)
	}
}

func (h *Handle) initWebSocket() {
	h.wsClient.HandleConnect(func(s *melody.Session) {
		userIDI, ok := s.Get("user_id")
		if !ok {
			logger.Warnf(context.Background(), "lab status ws connect: no user_id")
			return
		}
		userID := userIDI.(string)

		// 将 session 添加到用户的会话列表
		sessions, _ := h.userSessions.LoadOrStore(userID, &sync.Map{})
		sessionsMap := sessions.(*sync.Map)
		sessionsMap.Store(s, true)

		logger.Infof(context.Background(), "lab status ws connected: user_id=%s", userID)
	})

	h.wsClient.HandleDisconnect(func(s *melody.Session) {
		userIDI, ok := s.Get("user_id")
		if !ok {
			return
		}
		userID := userIDI.(string)

		// 从用户的会话列表移除
		if sessions, ok := h.userSessions.Load(userID); ok {
			sessionsMap := sessions.(*sync.Map)
			sessionsMap.Delete(s)
		}

		logger.Infof(context.Background(), "lab status ws disconnected: user_id=%s", userID)
	})

	h.wsClient.HandleMessage(func(s *melody.Session, msg []byte) {
		ctxI, _ := s.Get("ctx")
		ctx := ctxI.(*gin.Context)

		var baseMsg common.WsMsgType
		if err := json.Unmarshal(msg, &baseMsg); err != nil {
			logger.Errorf(ctx, "lab status ws parse message err: %+v", err)
			common.ReplyWSErr(s, "", baseMsg.MsgUUID, code.ParamErr.WithErr(err))
			return
		}

		switch baseMsg.Action {
		case ActionQueryList:
			h.handleQueryList(ctx, s, baseMsg.MsgUUID)
		case ActionQueryDetail:
			h.handleQueryDetail(ctx, s, msg)
		default:
			logger.Warnf(ctx, "lab status ws unknown action: %s", baseMsg.Action)
			common.ReplyWSErr(s, baseMsg.Action, baseMsg.MsgUUID, code.ParamErr.WithMsg("unknown action"))
		}
	})
}

// handleQueryList 处理查询用户所有实验室状态
func (h *Handle) handleQueryList(ctx context.Context, s *melody.Session, msgUUID uuid.UUID) {
	userIDI, _ := s.Get("user_id")
	userID := userIDI.(string)

	// 获取用户所有实验室
	labs, err := h.labStore.GetLabByUserID(ctx, &common.PageReqT[string]{
		PageReq: common.PageReq{Page: 1, PageSize: 1000},
		Data:    userID,
	})
	if err != nil {
		logger.Errorf(ctx, "handleQueryList GetLabByUserID err: %+v", err)
		common.ReplyWSErr(s, ActionQueryList, msgUUID, err)
		return
	}

	// 获取实验室ID列表
	labIDs := make([]int64, 0, len(labs.Data))
	for _, member := range labs.Data {
		labIDs = append(labIDs, member.LabID)
	}

	if len(labIDs) == 0 {
		common.ReplyWSOk(s, ActionQueryList, msgUUID, []LabStatusData{})
		return
	}

	// 获取实验室详情
	labDatas := make([]*model.Laboratory, 0, len(labIDs))
	if err := h.labStore.FindDatas(ctx, &labDatas, map[string]any{
		"id": labIDs,
	}, "id", "uuid", "is_online", "last_connected_at"); err != nil {
		logger.Errorf(ctx, "handleQueryList FindDatas err: %+v", err)
		common.ReplyWSErr(s, ActionQueryList, msgUUID, err)
		return
	}

	// 构建响应
	statusList := make([]LabStatusData, 0, len(labDatas))
	for _, lab := range labDatas {
		statusList = append(statusList, LabStatusData{
			LabUUID:         lab.UUID,
			IsOnline:        lab.IsOnline,
			LastConnectedAt: lab.LastConnectedAt,
		})
	}

	common.ReplyWSOk(s, ActionQueryList, msgUUID, statusList)
}

// handleQueryDetail 处理查询单个实验室状态
func (h *Handle) handleQueryDetail(ctx context.Context, s *melody.Session, msg []byte) {
	var req QueryDetailReq
	if err := json.Unmarshal(msg, &req); err != nil {
		logger.Errorf(ctx, "handleQueryDetail unmarshal err: %+v", err)
		common.ReplyWSErr(s, ActionQueryDetail, req.MsgUUID, code.ParamErr.WithErr(err))
		return
	}

	if req.Data.LabUUID.IsNil() {
		common.ReplyWSErr(s, ActionQueryDetail, req.MsgUUID, code.ParamErr.WithMsg("lab_uuid is required"))
		return
	}

	// 获取实验室信息
	lab, err := h.labStore.GetLabByUUID(ctx, req.Data.LabUUID, "uuid", "is_online", "last_connected_at")
	if err != nil {
		logger.Errorf(ctx, "handleQueryDetail GetLabByUUID err: %+v", err)
		common.ReplyWSErr(s, ActionQueryDetail, req.MsgUUID, err)
		return
	}

	status := LabStatusData{
		LabUUID:         lab.UUID,
		IsOnline:        lab.IsOnline,
		LastConnectedAt: lab.LastConnectedAt,
	}

	common.ReplyWSOk(s, ActionQueryDetail, req.MsgUUID, status)
}

// NotifyStatusChange 通知状态变化（由外部调用）
func (h *Handle) NotifyStatusChange(ctx context.Context, labUUID uuid.UUID, isOnline bool, lastConnectedAt *time.Time) {
	logger.Infof(ctx, "🔔 [LabStatus] NotifyStatusChange called: lab=%s, online=%v, time=%v", labUUID, isOnline, lastConnectedAt)

	// 获取实验室的所有成员
	lab, err := h.labStore.GetLabByUUID(ctx, labUUID, "id")
	if err != nil {
		logger.Errorf(ctx, "NotifyStatusChange GetLabByUUID err: %+v", err)
		return
	}

	logger.Infof(ctx, "📊 [LabStatus] Lab ID: %d, UUID: %s", lab.ID, lab.UUID)

	members, err := h.labStore.GetLabByLabID(ctx, &common.PageReqT[int64]{
		PageReq: common.PageReq{Page: 1, PageSize: 1000},
		Data:    lab.ID,
	})
	if err != nil {
		logger.Errorf(ctx, "NotifyStatusChange GetLabByLabID err: %+v", err)
		return
	}

	logger.Infof(ctx, "👥 [LabStatus] Found %d member(s) for lab %s", len(members.Data), labUUID)

	// 构建状态更新数据
	statusData := []LabStatusData{
		{
			LabUUID:         labUUID,
			IsOnline:        isOnline,
			LastConnectedAt: lastConnectedAt,
		},
	}

	msgUUID := uuid.NewV4()

	// 向所有成员发送通知
	sentCount := 0
	for _, member := range members.Data {
		logger.Infof(ctx, "🔍 [LabStatus] Checking user %s for active sessions...", member.UserID)

		if sessions, ok := h.userSessions.Load(member.UserID); ok {
			sessionsMap := sessions.(*sync.Map)
			sessionCount := 0
			sessionsMap.Range(func(key, value interface{}) bool {
				sessionCount++
				if session, ok := key.(*melody.Session); ok {
					// 使用标准的 WebSocket 响应格式
					if err := common.ReplyWSOk(session, ActionStatusUpdate, msgUUID, statusData); err != nil {
						logger.Errorf(ctx, "❌ [LabStatus] Failed to send to user %s session %d: %+v", member.UserID, sessionCount, err)
					} else {
						sentCount++
						logger.Infof(ctx, "✅ [LabStatus] Sent to user %s session %d", member.UserID, sessionCount)
					}
				}
				return true
			})
			logger.Infof(ctx, "📱 [LabStatus] User %s has %d active session(s)", member.UserID, sessionCount)
		} else {
			logger.Infof(ctx, "⚠️ [LabStatus] User %s has no active WebSocket sessions", member.UserID)
		}
	}

	logger.Infof(ctx, "✨ [LabStatus] NotifyStatusChange completed: sent to %d session(s)", sentCount)
}

func (h *Handle) Close() {
	if h.wsClient != nil {
		_ = h.wsClient.Close()
	}
}
