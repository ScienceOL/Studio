package device

import "context"

/*
	设备管理模块，管理该实验室下所有设备实时状态
*/

type Service interface {
	GetDeviceActionStatus(ctx context.Context, deviceName string, actionName string)
}
