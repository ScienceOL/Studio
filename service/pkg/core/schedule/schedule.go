package schedule

import (
	"context"
)



type Service interface {
	StartJob(ctx context.Context, jobID int64)
	StopJob(ctx context.Context)
	GetJobStatus(ctx context.Context)
}
