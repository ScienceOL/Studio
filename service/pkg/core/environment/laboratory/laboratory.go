package laboratory

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/core/environment"
	"github.com/scienceol/studio/service/pkg/middleware/auth"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/repo/casdoor"
	eStore "github.com/scienceol/studio/service/pkg/repo/environment"
	"github.com/scienceol/studio/service/pkg/repo/invite"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/datatypes"
)

type lab struct {
	envStore      repo.LaboratoryRepo
	accountClient repo.Account
	inviteStore   repo.Invite
}

func NewLab() environment.EnvService {
	return &lab{
		envStore:      eStore.New(),
		accountClient: casdoor.NewCasClient(),

		inviteStore: invite.New(),
	}
}

func (l *lab) CreateLaboratoryEnv(ctx context.Context, req *environment.LaboratoryEnvReq) (*environment.LaboratoryEnvResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}
	var data *model.Laboratory

	err := l.envStore.ExecTx(ctx, func(txCtx context.Context) error {
		data = &model.Laboratory{
			Name:         req.Name,
			UserID:       userInfo.ID,
			Status:       model.INIT,
			AccessKey:    "tmp",
			AccessSecret: "tmp",
			Description:  req.Description,
			BaseModel: model.BaseModel{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		if err := l.envStore.CreateLaboratoryEnv(ctx, data); err != nil {
			return err
		}

		if err := l.envStore.AddLabMemeber(ctx, &model.LaboratoryMember{
			UserID: data.UserID,
			LabID:  data.ID,
			Role:   model.LaboratoryMemberAdmin,
		}); err != nil {
			return err
		}

		ak := uuid.NewV4().String()
		sk := uuid.NewV4().String()
		if config.Global().OAuth2.AuthSource == config.AuthCasdoor {
			// 将 UUID 的连字符替换为下划线，确保符合 Casdoor 用户名规则
			// 用户名规则：只能包含字母数字、下划线或连字符，不能有连续的连字符/下划线，不能以连字符/下划线开头或结尾
			safeAk := strings.ReplaceAll(ak, "-", "_")
			labUsername := fmt.Sprintf("lab_%s_%s", data.UUID, safeAk[:8]) // 使用 lab_ 前缀 + UUID + 短 AK

			err := l.accountClient.CreateLabUser(txCtx, &model.LabInfo{
				AccessKey:         ak,
				AccessSecret:      sk,
				Name:              labUsername,
				DisplayName:       req.Name,
				Avatar:            "https://stroage.sciol.ac.cn/library/default_avatar.png",
				Owner:             userInfo.Owner,
				Type:              model.LABTYPE,
				Password:          uuid.NewV4().String(), // 使用随机密码，反正不会被使用
				SignupApplication: userInfo.SignupApplication,
			})

			if err != nil {
				return err
			}
		}
		data.AccessKey = ak
		data.AccessSecret = sk

		return l.envStore.UpdateLaboratoryEnv(txCtx, data)
	})
	if err != nil {
		return nil, err
	}

	return &environment.LaboratoryEnvResp{
		UUID:         data.UUID,
		Name:         data.Name,
		AccessKey:    data.AccessKey,
		AccessSecret: data.AccessSecret,
	}, nil
}

func (l *lab) UpdateLaboratoryEnv(ctx context.Context, req *environment.UpdateEnvReq) (*environment.LaboratoryResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	data := &model.Laboratory{
		BaseModel: model.BaseModel{
			UUID:      req.UUID,
			UpdatedAt: time.Now(),
		},
		Name:        req.Name,
		UserID:      userInfo.ID,
		Description: req.Description,
	}

	err := l.envStore.UpdateLaboratoryEnv(ctx, data)
	if err != nil {
		return nil, err
	}
	return &environment.LaboratoryResp{
		UUID:        data.UUID,
		Name:        data.Name,
		Description: data.Description,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
	}, nil
}

func (l *lab) DelLab(ctx context.Context, req *environment.DelLabReq) error {
	if req.UUID.IsNil() {
		return code.ParamErr.WithMsg("lab uuid is empty")
	}
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}

	lab, err := l.envStore.GetLabByUUID(ctx, req.UUID)
	if err != nil {
		return err
	}

	if lab.UserID != userInfo.ID {
		// 退出该实验室
		if err := l.envStore.DelData(ctx, &model.LaboratoryMember{}, map[string]any{
			"lab_id":  lab.ID,
			"user_id": userInfo.ID,
		}); err != nil {
			return err
		}
	} else {
		// 自己删除实验室，清空所有成员
		if err := l.envStore.DelData(ctx, &model.LaboratoryMember{}, map[string]any{
			"lab_id": lab.ID,
		}); err != nil {
			return err
		}

		if err := l.envStore.UpdateData(ctx, &model.Laboratory{
			Status: model.DELETED,
		}, map[string]any{
			"id": lab.ID,
		}, "status"); err != nil {
			return err
		}
	}

	return nil
}

func (l *lab) LabInfo(ctx context.Context, req *environment.LabInfoReq) (*environment.LabInfoResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	if req.UUID.IsNil() {
		return nil, code.ParamErr
	}

	lab := &model.Laboratory{}
	var err error

	if req.Type == environment.LABAK {
		err = l.envStore.GetData(ctx, lab, map[string]any{
			"access_key": req.UUID,
		})
	} else {
		lab, err = l.envStore.GetLabByUUID(ctx, req.UUID)
	}

	if err != nil {
		return nil, err
	}

	resp := &environment.LabInfoResp{
		UUID:         uuid.UUID{},
		Name:         "",
		UserID:       "",
		IsAdmin:      false,
		AccessKey:    "",
		AccessSecret: "",
		Status:       model.DELETED,
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
	}
	if lab.Status == model.DELETED {
		return resp, nil
	}

	count, err := l.envStore.Count(ctx, &model.LaboratoryMember{}, map[string]any{
		"lab_id":  lab.ID,
		"user_id": userInfo.ID,
	})
	if err != nil || count == 0 {
		return nil, code.NoPermission
	}

	if lab.Status != model.DELETED {
		resp.UUID = lab.UUID
		resp.Name = lab.Name
		resp.UserID = lab.UserID
		resp.IsAdmin = lab.UserID == userInfo.ID
		resp.AccessKey = lab.AccessKey
		resp.AccessSecret = lab.AccessSecret
		resp.Status = lab.Status
		resp.IsOnline = lab.IsOnline
		resp.LastConnectedAt = lab.LastConnectedAt
		resp.CreatedAt = lab.CreatedAt
		resp.UpdatedAt = lab.UpdatedAt
	}

	return resp, nil
}

func (l *lab) CreateResource(ctx context.Context, req *environment.ResourceReq) error {
	if len(req.Resources) == 0 {
		return code.ResourceIsEmptyErr
	}

	labInfo := auth.GetLabUser(ctx)
	if labInfo == nil {
		return code.UnLogin
	}
	labData, err := l.envStore.GetLabByAkSk(ctx, labInfo.AccessKey, labInfo.AccessSecret)
	if err != nil {
		return err
	}

	return db.DB().ExecTx(ctx, func(txCtx context.Context) error {
		resDatas := utils.FilterSlice(req.Resources, func(item *environment.Resource) (*model.ResourceNodeTemplate, bool) {
			data := &model.ResourceNodeTemplate{
				Name:         item.RegName,
				LabID:        labData.ID,     // 实验室的 id
				UserID:       labData.UserID, // 创建实验室的 user id
				Header:       item.RegName,
				Footer:       "",
				Version:      utils.Or(item.Version, "0.0.1"),
				Icon:         item.Icon,
				Description:  item.Description,
				Model:        item.Model,
				Module:       item.Class.Module,
				ResourceType: item.ResourceType,
				Language:     item.Class.Type,
				StatusTypes:  item.Class.StatusTypes,
				ConfigInfo:   item.ConfigInfo,
				Tags:         item.Tags,
				DataSchema: utils.SafeValue(func() datatypes.JSON {
					return item.InitParamSchema.Data.Properties
				}, datatypes.JSON{}),
				ConfigSchema: utils.SafeValue(
					func() datatypes.JSON { return item.InitParamSchema.Config.Properties },
					datatypes.JSON{}),
			}
			item.SelfDB = data
			return data, true
		})

		if err := l.envStore.UpsertResourceNodeTemplate(txCtx, resDatas); err != nil {
			return err
		}

		// if err := l.createConfigInfo(txCtx, req.Resources); err != nil {
		// 	return err
		// }

		if err := l.createHandle(txCtx, req.Resources); err != nil {
			return err
		}

		actions, err := l.createWorkflowNodeTemplate(txCtx, req.Resources)
		if err != nil {
			return err
		}

		return l.createActionHandles(ctx, actions)
	})
}

func (l *lab) createWorkflowNodeTemplate(ctx context.Context, res []*environment.Resource) ([]*model.WorkflowNodeTemplate, error) {
	resDeviceAction, err := utils.FilterSliceWithErr(res, func(item *environment.Resource) ([]*model.WorkflowNodeTemplate, bool, error) {
		actions := make([]*model.WorkflowNodeTemplate, 0, len(item.Class.ActionValueMappings))
		for actionName, action := range item.Class.ActionValueMappings {
			if actionName == "" {
				return nil, false, code.RegActionNameEmptyErr
			}

			actions = append(actions, &model.WorkflowNodeTemplate{
				LabID:          item.SelfDB.LabID,
				ResourceNodeID: item.SelfDB.ID,
				Name:           actionName,
				Goal:           action.Goal,
				GoalDefault:    action.GoalDefault,
				Feedback:       action.Feedback,
				Result:         action.Result,
				Schema:         action.Schema,
				Type:           action.Type,
				Handles:        action.Handles,
				Header:         actionName,
				Footer:         item.SelfDB.Name,
			})
		}
		return actions, true, nil
	})
	if err != nil {
		return nil, err
	}
	return resDeviceAction, l.envStore.UpsertWorkflowNodeTemplate(ctx, resDeviceAction)
}

func (l *lab) createHandle(ctx context.Context, res []*environment.Resource) error {
	resDeviceHandles, err := utils.FilterSliceWithErr(res, func(item *environment.Resource) ([]*model.ResourceHandleTemplate, bool, error) {
		handles := make([]*model.ResourceHandleTemplate, 0, len(item.Handles))
		for _, handle := range item.Handles {
			handles = append(handles, &model.ResourceHandleTemplate{
				ResourceNodeID: item.SelfDB.ID,
				Name:           handle.HandlerKey,
				DisplayName:    handle.Label,
				Type:           handle.DataType,
				IOType:         handle.IoType,
				Source:         handle.DataSource,
				Key:            handle.DataKey,
				Side:           handle.Side,
			})
		}
		return handles, true, nil
	})
	if err != nil {
		return err
	}

	return l.envStore.UpsertResourceHandleTemplate(ctx, resDeviceHandles)
}

func (l *lab) createActionHandles(ctx context.Context, actions []*model.WorkflowNodeTemplate) error {
	resHandles, _ := utils.FilterSliceWithErr(actions, func(item *model.WorkflowNodeTemplate) ([]*model.WorkflowHandleTemplate, bool, error) {
		resHi, _ := utils.FilterSliceWithErr(item.Handles.Data().Input, func(h *model.Handle) ([]*model.WorkflowHandleTemplate, bool, error) {
			return []*model.WorkflowHandleTemplate{{
				WorkflowNodeID: item.ID,
				HandleKey:      h.HandlerKey,
				IoType:         "source",
				DisplayName:    h.Label,
				Type:           h.DataType,
				DataSource:     h.DataSource,
				DataKey:        h.DataKey,
			}}, true, nil
		})
		resHo, _ := utils.FilterSliceWithErr(item.Handles.Data().Output, func(h *model.Handle) ([]*model.WorkflowHandleTemplate, bool, error) {
			return []*model.WorkflowHandleTemplate{{
				WorkflowNodeID: item.ID,
				HandleKey:      h.HandlerKey,
				IoType:         "source",
				DisplayName:    h.Label,
				Type:           h.DataType,
				DataSource:     h.DataSource,
				DataKey:        h.DataKey,
			}}, true, nil
		})

		resH := make([]*model.WorkflowHandleTemplate, 0, len(resHi)+len(resHo)+2)

		resH = append(resH, &model.WorkflowHandleTemplate{
			WorkflowNodeID: item.ID,
			HandleKey:      "ready",
			IoType:         "target",
		})
		resH = append(resH, &model.WorkflowHandleTemplate{
			WorkflowNodeID: item.ID,
			HandleKey:      "ready",
			IoType:         "source",
		})
		resH = append(resH, resHi...)
		resH = append(resH, resHo...)

		return resH, true, nil
	})

	return l.envStore.UpsertActionHandleTemplate(ctx, resHandles)
}

func (l *lab) LabList(ctx context.Context, req *common.PageReq) (*common.PageMoreResp[[]*environment.LaboratoryResp], error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	labs, err := l.envStore.GetLabByUserID(ctx, &common.PageReqT[string]{
		PageReq: *req,
		Data:    userInfo.ID,
	})
	if err != nil {
		return nil, err
	}

	labIDs := utils.FilterSlice(labs.Data, func(labMemeber *model.LaboratoryMember) (int64, bool) {
		return labMemeber.LabID, true
	})

	labDatas := make([]*model.Laboratory, 0, len(labIDs))
	if err := l.envStore.FindDatas(ctx, &labDatas, map[string]any{
		"id": labIDs,
	}); err != nil {
		return nil, err
	}

	labMap := utils.Slice2Map(labDatas, func(l *model.Laboratory) (int64, *model.Laboratory) {
		return l.ID, l
	})

	labMemberMap := l.envStore.GetLabMemberCount(ctx, labIDs...)

	labResp := utils.FilterSlice(labs.Data, func(item *model.LaboratoryMember) (*environment.LaboratoryResp, bool) {
		lab, ok := labMap[item.LabID]
		if !ok {
			logger.Infof(ctx, "can not found lab id: %+d", item.LabID)
			return nil, false
		}

		return &environment.LaboratoryResp{
			UUID:            lab.UUID,
			Name:            lab.Name,
			UserID:          lab.UserID,
			Description:     lab.Description,
			MemberCount:     labMemberMap[lab.ID],
			IsAdmin:         lab.UserID == userInfo.ID,
			IsOnline:        lab.IsOnline,
			LastConnectedAt: lab.LastConnectedAt,
			CreatedAt:       lab.CreatedAt,
			UpdatedAt:       lab.UpdatedAt,
		}, true
	})

	return &common.PageMoreResp[[]*environment.LaboratoryResp]{
		Data:     labResp,
		Page:     labs.Page,
		HasMore:  labs.Total > int64(labs.Page)*int64(labs.PageSize),
		PageSize: labs.PageSize,
	}, nil
}

func (l *lab) LabMemberList(ctx context.Context, req *environment.LabMemberReq) (*common.PageResp[[]*environment.LabMemberResp], error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	labID := l.envStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	if labID == 0 {
		return nil, code.CanNotGetLabIDErr
	}

	c, err := l.envStore.Count(ctx, &model.LaboratoryMember{}, map[string]any{
		"user_id": userInfo.ID,
		"lab_id":  labID,
	})
	if err != nil {
		return nil, err
	}
	if c == 0 {
		return nil, code.NoPermission
	}

	labMembers, err := l.envStore.GetLabByLabID(ctx, &common.PageReqT[int64]{
		PageReq: req.PageReq,
		Data:    labID,
	})
	if err != nil {
		return nil, err
	}

	// 获取实验室所有者信息用于判断是否为管理员
	labData, _ := l.envStore.GetLabByUUID(ctx, req.LabUUID)

	resp := utils.FilterSlice(labMembers.Data, func(member *model.LaboratoryMember) (*environment.LabMemberResp, bool) {
		isAdmin := labData != nil && labData.UserID == member.UserID

		return &environment.LabMemberResp{
			UUID:    member.UUID,
			UserID:  member.UserID,
			LabID:   member.LabID,
			Role:    member.Role,
			IsAdmin: isAdmin,
		}, true
	})

	return &common.PageResp[[]*environment.LabMemberResp]{
		Total:    labMembers.Total,
		Page:     labMembers.Page,
		PageSize: labMembers.PageSize,
		Data:     resp,
	}, nil
}

func (l *lab) DelLabMember(ctx context.Context, req *environment.DelLabMemberReq) error {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}

	labID := l.envStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	if labID <= 0 {
		return code.LabNotFound
	}

	// 只有管理员可以删除
	if count, err := l.envStore.Count(ctx, &model.LaboratoryMember{}, map[string]any{
		"lab_id":  labID,
		"user_id": userInfo.ID,
		"role":    model.LaboratoryMemberAdmin,
	}); err != nil {
		return err
	} else if count == 0 {
		return code.NoPermission
	}

	datas := []*model.LaboratoryMember{}

	if err := l.envStore.FindDatas(ctx, &datas, map[string]any{
		"uuid": req.MemberUUID,
	}); err != nil {
		return err
	}

	if len(datas) != 1 {
		return nil
	}

	if datas[0].UserID == userInfo.ID {
		return code.NoPermission
	}

	return l.envStore.DelData(ctx, &model.LaboratoryMember{}, map[string]any{
		"uuid": req.MemberUUID,
	})
}

func (l *lab) CreateInvite(ctx context.Context, req *environment.InviteReq) (*environment.InviteResp, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	labID := l.envStore.UUID2ID(ctx, &model.Laboratory{}, req.LabUUID)[req.LabUUID]
	if labID <= 0 {
		return nil, code.LabNotFound
	}

	// 只有管理员可以创建
	if count, err := l.envStore.Count(ctx, &model.LaboratoryMember{}, map[string]any{
		"lab_id":  labID,
		"user_id": userInfo.ID,
		"role":    model.LaboratoryMemberAdmin,
	}); err != nil {
		return nil, err
	} else if count == 0 {
		return nil, code.NoPermission
	}

	data := &model.LaboratoryInvitation{
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Type:      model.InvitationTypeLab,
		ThirdID:   strconv.FormatInt(labID, 10),
		UserID:    userInfo.ID,
	}
	if err := l.inviteStore.CreateData(ctx, data); err != nil {
		return nil, err
	}

	return &environment.InviteResp{
		Path: fmt.Sprintf("/api/v1/lab/invite/%s", data.UUID),
	}, nil
}

func (l *lab) AcceptInvite(ctx context.Context, req *environment.AcceptInviteReq) error {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return code.UnLogin
	}

	datas := make([]*model.LaboratoryInvitation, 0, 1)
	if err := l.inviteStore.FindDatas(ctx, &datas, map[string]any{
		"uuid": req.UUID,
	}); err != nil {
		return err
	}

	if len(datas) != 1 {
		return code.LabInviteNotFoundErr
	}

	data := datas[0]
	if data.ExpiresAt.Unix() < time.Now().Unix() {
		return code.InviteExpiredErr
	}

	if data.UserID == userInfo.ID {
		return nil
	}

	switch data.Type {
	case model.InvitationTypeLab:
		return l.addLabMemeber(ctx, data)

	default:
		logger.Warnf(ctx, "can not found this invite type: %+s", data.Type)
	}

	return nil
}

func (l *lab) addLabMemeber(ctx context.Context, data *model.LaboratoryInvitation) error {
	userInfo := auth.GetCurrentUser(ctx)
	labID, err := strconv.ParseInt(data.ThirdID, 10, 64)
	if err != nil {
		return code.InvalidateThirdID.WithErr(err)
	}

	return l.envStore.AddLabMemeber(ctx, &model.LaboratoryMember{
		UserID: userInfo.ID,
		LabID:  labID,
		Role:   model.LaboratoryMemberNormal,
	})
}

func (l *lab) UserInfo(ctx context.Context) (*model.UserData, error) {
	userInfo := auth.GetCurrentUser(ctx)
	if userInfo == nil {
		return nil, code.UnLogin
	}

	return l.accountClient.GetUserInfo(ctx, userInfo.ID)
}
