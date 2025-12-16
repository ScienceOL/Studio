package edge

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/olahol/melody"
	"github.com/scienceol/studio/service/pkg/core/schedule/engine"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

func (e *EdgeImpl) sendAction(ctx context.Context, s *melody.Session, data any) {
	bData, _ := json.Marshal(data)
	if err := s.Write(bData); err != nil {
		logger.Errorf(ctx, "EdgeImpl.sendAction err: %+v", err)
	}
}

func (e *EdgeImpl) isTaskNil(_ context.Context, t engine.Task) bool {
	if t == nil {
		return true
	}

	v := reflect.ValueOf(t)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return true
	}

	return false
}
