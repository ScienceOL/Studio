package casdoor

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"gorm.io/gorm"
)

type casClient struct {
	casDoorClient *resty.Client
	repo.IDOrUUIDTranslate
}

func NewCasClient() repo.Account {
	conf := config.Global().OAuth2
	return &casClient{
		IDOrUUIDTranslate: repo.NewBaseDB(),
		casDoorClient: resty.New().
			EnableTrace().
			SetBaseURL(conf.Addr),
	}
}

func NewLabAccess() repo.LabAccount {
	conf := config.Global().OAuth2
	return &casClient{
		IDOrUUIDTranslate: repo.NewBaseDB(),
		casDoorClient: resty.New().
			EnableTrace().
			SetBaseURL(conf.Addr),
	}
}

func (c *casClient) CreateLabUser(ctx context.Context, user *model.LabInfo) error {
	resData := &model.LabInfoResp{}
	conf := config.Global().OAuth2
	resp, err := c.casDoorClient.R().SetContext(ctx).
		SetBody(user).
		SetResult(resData).
		SetBasicAuth(conf.ClientID, conf.ClientSecret).
		SetResult(nil).Post("/api/add-user")
	if err != nil {
		logger.Errorf(ctx, "CreateLabUser err: %+v user: %+v", err, user)
		return code.CasDoorCreateLabUserErr.WithMsg(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Errorf(ctx, "CreateLabUser http code: %d", resp.StatusCode())
		return code.CasDoorCreateLabUserErr
	}

	if resData.Status != "ok" {
		logger.Errorf(ctx, "CreateLabUser res data err: %+v", resData)
		return code.CasDoorCreateLabUserErr
	}

	return nil
}

func (c *casClient) GetLabUserInfo(ctx context.Context, req *model.LabAkSk) (*model.UserData, error) {
	resData := &model.UserInfo{}
	resp, err := c.casDoorClient.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"accessKey":    req.AccessKey,
			"accessSecret": req.AccessSecret,
		}).
		SetResult(resData).Get("/api/get-account")
	if err != nil {
		logger.Errorf(ctx, "GetLabUserInfo err: %+v req: %+v", err, req)
		return nil, code.CasDoorQueryLabUserErr.WithMsg(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Errorf(ctx, "GetLabUserInfo http code: %d", resp.StatusCode())
		return nil, code.CasDoorQueryLabUserErr
	}

	// 查表获取 LabID 和 LabUUID
	labData := &model.Laboratory{}
	if err := c.DBWithContext(ctx).
		Where("access_key = ? and access_secret = ?",
			req.AccessKey, req.AccessSecret).
		Select("id", "uuid").
		Take(labData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.RecordNotFound
		}

		logger.Errorf(ctx, "GetLabUserInfo fail err: %+v", err)
		return nil, err
	}

	return &model.UserData{
		ID:      resData.Data.ID,
		LabID:   labData.ID,
		LabUUID: labData.UUID,
	}, nil
}

func (c *casClient) DelLabUserInfo(_ context.Context, _ *model.LabAkSk) error {
	panic("not impl")
}

func (c *casClient) BatchGetUserInfo(ctx context.Context, userIDs []string) ([]*model.UserData, error) {
	panic("not impl")
}

func (c *casClient) GetUserInfo(ctx context.Context, userID string) (*model.UserData, error) {
	panic("not impl")
}
