package nacos

import (
	"context"
	"fmt"
	"os"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type Kms struct {
	Endpoint         string
	Password         string
	ClientKeyContent string
}

type nacosClient struct {
	Ctx    context.Context
	client config_client.IConfigClient
	*NacoConf
}

type NacoConf struct {
	Endpoint    string
	ContextPath string
	NamespaceID string
	AccessKey   string
	SecretKey   string
	User        string
	Password    string
	Port        uint64
	RegionID    string
	DataID      string // naco key
	Group       string // naco group
	NeedWatch   bool   // 是否需要动态监控变化
	Kms         *Kms   // kms 加密
}

func (conf *NacoConf) checkParam() error {
	if conf.Endpoint == "" {
		return fmt.Errorf("endpoint is empty")
	}

	if conf.ContextPath == "" {
		conf.ContextPath = "/nacos"
	}
	if conf.Port == 0 {
		conf.Port = 8848
	}
	if conf.RegionID == "" {
		conf.RegionID = "cn-zhangjiakou"
	}
	if conf.DataID == "" {
		return fmt.Errorf("data id is empty")
	}
	if conf.Group == "" {
		return fmt.Errorf("group is empty")
	}

	if conf.Kms != nil {
		if conf.Kms.Endpoint == "" {
			return fmt.Errorf("kms endpoint is empty")
		}

		if conf.Kms.Password == "" {
			return fmt.Errorf("kms password is empty")
		}

		if conf.Kms.ClientKeyContent == "" {
			return fmt.Errorf("kms client key content is empty")
		}
	}
	return nil
}

// 如果初始化化失败，会直接退出进程
func MustInit(ctx context.Context, conf *NacoConf, onContent func(content []byte) error) {
	n := &nacosClient{Ctx: ctx, NacoConf: conf}

	if err := conf.checkParam(); err != nil {
		logger.Fatalf(ctx, "参数校验失败 err: %+v", err)
		os.Exit(1)
	}

	if err := n.initClient(ctx); err != nil {
		logger.Fatalf(ctx, "初始化 naco 失败 err: %v\n", err)
		os.Exit(1)
	}

	if err := n.getConf(ctx, onContent); err != nil {
		logger.Fatalf(ctx, "初始化 naco 失败 err: %v\n", err)
		os.Exit(1)
	}
}

func (nc *nacosClient) getConf(ctx context.Context, onContent func(content []byte) error) error {
	content, err := nc.get()
	if err != nil {
		return fmt.Errorf("get dynamic config failed, err: %v", err)
	}

	if err := onContent([]byte(content)); err != nil {
		return fmt.Errorf("nacos content onchange err: %v", err)
	}

	if nc.NeedWatch {
		err = nc.watch(onContent)
		if err != nil {
			logger.Fatalf(ctx, "nacos watch config fail: %+v", err)
		}
	}
	return nil
}

func (nc *nacosClient) initClient(_ context.Context) error {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nc.Endpoint, nc.Port, constant.WithContextPath(nc.ContextPath)),
	}

	opts := make([]constant.ClientOption, 0, 20)
	opts = append(opts,
		constant.WithNamespaceId(nc.NamespaceID),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("warn"),
		constant.WithAccessKey(nc.AccessKey),
		constant.WithSecretKey(nc.SecretKey),
		constant.WithUsername(nc.User),
		constant.WithPassword(nc.Password),
	)

	if nc.Kms != nil {
		opts = append(opts,
			constant.WithRegionId(nc.RegionID),
			constant.WithOpenKMS(true),
			constant.WithKMSv3Config(
				&constant.KMSv3Config{
					Endpoint:         nc.Kms.Endpoint,
					Password:         nc.Kms.Password,
					ClientKeyContent: nc.Kms.ClientKeyContent,
				}))
	}

	// create ClientConfig
	cc := *constant.NewClientConfig(opts...)

	// create config client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return fmt.Errorf("init nacos err: %v", err)
	}
	nc.client = client

	return err
}

func (nc *nacosClient) get() (string, error) {
	return nc.client.GetConfig(vo.ConfigParam{
		DataId: nc.DataID,
		Group:  nc.Group,
	})
}

func (nc *nacosClient) watch(onContent func(content []byte) error) error {
	return nc.client.ListenConfig(vo.ConfigParam{
		DataId: nc.DataID,
		Group:  nc.Group,
		OnChange: func(namespace, group, dataID, data string) {
			if err := onContent([]byte(data)); err != nil {
				logger.Errorf(nc.Ctx, "namespace: %s, group: %s, dataID: %s watch naco decode content err:%s", namespace, group, dataID, err.Error())
			}
		},
	})
}
