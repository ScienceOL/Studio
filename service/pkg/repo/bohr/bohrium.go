package bohr

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"github.com/scienceol/studio/service/pkg/utils"
	"gorm.io/gorm"
)

type BohrUserInfo struct {
	ID              int64  `json:"id"`
	Status          int    `json:"status"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	NickName        string `json:"nickname"`
	NickNameEn      string `json:"nicknameEn"`
	Phone           string `json:"phone"`
	Kind            int    `json:"kind"`
	PhoneVerify     int    `json:"phoneVerify"`
	Oversea         int    `json:"oversea"`
	AreaCode        int    `json:"areaCode"`
	UserNo          string `json:"userNo"`
	ActivityId      int    `json:"activityId"`
	UtmCampaign     string `json:"utmCampaign"`
	CourseApply     bool   `json:"courseApply"`
	MemberRole      int    `json:"memberRole"`
	LoginSource     string `json:"loginSource"`
	UtmSource       string `json:"utmSource"`
	RegisterChannel string `json:"registerChannel"`
	ActivityUserId  int64  `json:"activityUserId"`
	IsBindWechat    bool   `json:"isBindWechat"`
	IsSetPwd        bool   `json:"isSetPwd"`
	WeChatNickname  string `json:"weChatNickname"`
	Avatar          string `json:"avatar"`
}

type BohrImpl struct {
	bohrClient *resty.Client
	repo.IDOrUUIDTranslate
}

func New() repo.Account {
	conf := webapp.Config().OAuth2
	return &BohrImpl{
		bohrClient: resty.New().
			EnableTrace().
			SetBaseURL(conf.Addr),
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}

func NewLab() repo.LabAccount {
	conf := webapp.Config().OAuth2
	return &BohrImpl{
		bohrClient: resty.New().
			EnableTrace().
			SetBaseURL(conf.Addr),
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}

func (b *BohrImpl) CreateLabUser(ctx context.Context, user *model.LabInfo) error {
	panic("not impl")
}

func (b *BohrImpl) GetLabUserInfo(ctx context.Context, req *model.LabAkSk) (*model.UserData, error) {
	// 实验室用户就是创建该实验室的人
	labData := &model.Laboratory{}
	if err := b.DBWithContext(ctx).
		Where("access_key = ? and access_secret = ?",
			req.AccessKey, req.AccessSecret).
		Select("user_id").
		Take(labData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetLabUserInfo fail err: %+v", err)
		return nil, err
	}

	return &model.UserData{
		ID: labData.UserID,
	}, nil
}

func (b *BohrImpl) DelLabUserInfo(ctx context.Context, req *model.LabAkSk) error {
	panic("not impl")
}

func (b *BohrImpl) BatchGetUserInfo(ctx context.Context, userIDs []string) ([]*model.UserData, error) {
	resData := &common.RespT[[]*BohrUserInfo]{}
	resp, err := b.bohrClient.R().
		SetContext(ctx).
		SetBody(map[string]any{
			"ids": userIDs,
		}).
		SetResult(resData).Post("/account_api/users/list")
	if err != nil {
		logger.Errorf(ctx, "BatchGetUserInfo err: %+v user ids : %+v", err, userIDs)
		return nil, code.CasDoorQueryLabUserErr.WithMsg(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Errorf(ctx, "BatchGetUserInfo http code: %d", resp.StatusCode())
		return nil, code.CasDoorQueryLabUserErr
	}

	if resData.Code != code.Success {
		logger.Errorf(ctx, "BatchGetUserInfo resp code not zero err msg:%+v", *resData.Error)
		return nil, code.BohrBatchQueryErr
	}

	return utils.FilterSlice(resData.Data, func(item *BohrUserInfo) (*model.UserData, bool) {
		return &model.UserData{
			Owner:             "",
			Name:              item.Name,
			ID:                strconv.FormatInt(item.ID, 10),
			Avatar:            item.Avatar,
			Type:              "",
			DisplayName:       item.NickName,
			SignupApplication: "",
			Phone:             item.Phone,
			Status:            item.Status,
			UserNo:            item.UserNo,
			Email:             item.Email,
		}, true
	}), nil
}

func (b *BohrImpl) GetUserInfo(ctx context.Context, userID string) (*model.UserData, error) {
	resData := &common.RespT[*BohrUserInfo]{}
	resp, err := b.bohrClient.R().
		SetContext(ctx).
		SetPathParam("id", userID).
		SetResult(resData).Get("/account_api/users/single_user/{id}")
	if err != nil {
		logger.Errorf(ctx, "GetUserInfo err: %+v user id : %+v", err, userID)
		return nil, code.CasDoorQueryLabUserErr.WithMsg(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Errorf(ctx, "GetUserInfo http code: %d", resp.StatusCode())
		return nil, code.CasDoorQueryLabUserErr
	}

	if resData.Code != code.Success {
		logger.Errorf(ctx, "GetUserInfo resp code not zero err msg:%+v", *resData.Error)
		return nil, code.BohrBatchQueryErr
	}

	return &model.UserData{
		Owner:             "",
		Name:              resData.Data.Name,
		ID:                strconv.FormatInt(resData.Data.ID, 10),
		Avatar:            resData.Data.Avatar,
		Type:              "",
		DisplayName:       resData.Data.NickName,
		SignupApplication: "",
		Phone:             resData.Data.Phone,
		Status:            resData.Data.Status,
		UserNo:            resData.Data.UserNo,
		Email:             resData.Data.Email,
	}, nil
}
