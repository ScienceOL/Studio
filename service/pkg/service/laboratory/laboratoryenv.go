package laboratory

import "context"

type Laboratory struct{}

func NewLaboratory() *Laboratory {
	return &Laboratory{}
}

func (l *Laboratory) GetEnvs(ctx context.Context) (*LaboratoryEnv, error) {
	// 函数
	return &LaboratoryEnv{}, nil
}
