package sandbox

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/scienceol/studio/service/internal/config"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo"
)

// {"code":0,"message":"success","data":{"error":"","stdout":"hello\n"}}

type Data struct {
	Error  string `json:"error"`
	Stdout string `json:"stdout"`
}

type SandboxRet struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type SandboxImpl struct {
	base   *BaseTemplateTransformer
	client *resty.Client
}

func NewSandbox() repo.Sandbox {
	sandboxConf := config.Global().RPC.Sandbox

	return &SandboxImpl{
		base: NewBaseTemplateTransformer(),
		client: resty.New().
			EnableTrace().
			SetHeaders(map[string]string{
				"X-Api-Key":    sandboxConf.ApiKey,
				"Content-Type": "application/json",
			}).
			SetBaseURL(sandboxConf.Addr),
	}
}

func (s *SandboxImpl) ExecCode(ctx context.Context, pyCode string, inputs map[string]any) (map[string]any, string, error) {
	runnerScript, preloadScript, err := s.base.TransformCaller(pyCode, inputs, NewPython3TemplateTransformer())
	if err != nil {
		return nil, "", err
	}

	ret := &SandboxRet{}
	res, err := s.client.R().SetContext(ctx).
		SetBody(map[string]any{
			"language":       "python3",
			"code":           runnerScript,
			"preload":        preloadScript,
			"enable_network": true,
		}).
		SetResult(ret).Post("/api/v1/sandbox/run")
	if err != nil {
		logger.Errorf(ctx, "ExecCode post run code err: %+v", err)
		return nil, "", code.RPCHttpErr.WithErr(err)
	}

	if res.StatusCode() != http.StatusOK {
		logger.Errorf(ctx, "ExecCode fail py code: %s, http code: %+v", pyCode, res.StatusCode())
		return nil, "", code.RPCHttpCodeErr.WithMsgf("http code: %d", res.StatusCode())
	}

	if ret.Code != 0 {
		logger.Errorf(ctx, "ExecCode code not success py code: %s, code: %+d", pyCode, ret.Code)
		return nil, "", code.RPCHttpCodeErr.WithMsgf("code: %d", ret.Code)
	}

	if ret.Data.Error != "" {
		return nil, ret.Data.Error, code.ExecWorkflowNodeScriptErr.WithMsg(ret.Data.Error)
	}

	codeRet, err := s.base.TransformResponse(ret.Data.Stdout)
	if err != nil {
		return nil, err.Error(), code.ExecWorkflowNodeScriptErr.WithErr(err)
	}

	return codeRet, "", err
}
