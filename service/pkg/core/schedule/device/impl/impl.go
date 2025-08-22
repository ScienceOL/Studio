package impl

import (
	"context"

	"github.com/scienceol/studio/service/pkg/core/schedule/device"
)

type deviceManager struct {
}

func NewDeviceManager(ctx context.Context) device.Service {
	return &deviceManager{}
}

func (d *deviceManager) GetDeviceActionStatus(ctx context.Context, deviceName string, actionName string) {
	panic("not implemented") // TODO: Implement
}
