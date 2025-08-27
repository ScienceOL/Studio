package impl

import (
	"context"

	"github.com/scienceol/studio/service/pkg/core/schedule/device"
)

type deviceManager struct{}

func NewDeviceManager(_ context.Context) device.Service {
	return &deviceManager{}
}

func (d *deviceManager) GetDeviceActionStatus(_ context.Context, _ string, _ string) {
	panic("not implemented") // TODO: Implement
}
