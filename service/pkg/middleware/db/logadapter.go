package db

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

type gormLoggerAdapter struct {
	zapLogger     *otelzap.Logger
	logLevel      gormLogger.LogLevel
	slowThreshold time.Duration
}

func NewGormLogger(logger *otelzap.Logger, level gormLogger.LogLevel, slowThreshold time.Duration) gormLogger.Interface {
	return &gormLoggerAdapter{
		zapLogger:     logger,
		slowThreshold: slowThreshold,
		logLevel:      gormLogger.Warn, // 默认日志级别
	}
}

func (l *gormLoggerAdapter) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func (l *gormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormLogger.Info || l.zapLogger == nil {
		return
	}
	l.zapLogger.Ctx(ctx).Info(fmt.Sprintf(msg, data...), zap.Any("caller", "gorm"))
}

func (l *gormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormLogger.Warn || l.zapLogger == nil {
		return
	}
	l.zapLogger.Ctx(ctx).Warn(fmt.Sprintf(msg, data...), zap.Any("caller", "gorm"))
}

func (l *gormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormLogger.Error || l.zapLogger == nil {
		return
	}
	l.zapLogger.Ctx(ctx).Error(fmt.Sprintf(msg, data...), zap.Any("caller", "gorm"))
}

func (l *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormLogger.Silent || l.zapLogger == nil {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	logger := otelzap.Ctx(ctx).WithOptions(zap.AddCallerSkip(3))

	switch {
	case err != nil && l.logLevel >= gormLogger.Error:
		logger.Error("SQL执行错误", zap.Error(err), zap.Duration("elapsed", elapsed), zap.String("sql", sql), zap.Int64("rows", rows))
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.logLevel >= gormLogger.Warn:
		logger.Warn(fmt.Sprintf("慢查询 (耗时 %s)", elapsed), zap.Duration("threshold", l.slowThreshold), zap.String("sql", sql), zap.Int64("rows", rows))
	case l.logLevel == gormLogger.Info:
		logger.Debug("SQL查询", zap.Duration("elapsed", elapsed), zap.String("sql", sql), zap.Int64("rows", rows))
	}
}
