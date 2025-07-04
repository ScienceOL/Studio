package db

import (
	"context"
	"fmt"
	"time"

	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/plugin/opentelemetry/tracing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const (
	timeoutIdel   = time.Hour
	maxIdleConn   = 10
	maxOpenConn   = 100
	slowThreshold = time.Second * 10
)

type LogConf struct {
	SlowThreshold time.Duration
	Level         string
}

var Level gormLogger.LogLevel

type Config struct {
	Host   string
	Port   int
	User   string
	PW     string
	DBName string
	LogConf

	Conns
}

type Conns struct {
	TimeoutIdle time.Duration
	MaxIdleConn int
	MaxOpenConn int
}

func initPG(ctx context.Context, conf *Config) *gorm.DB {
	if err := conf.parseConf(); err != nil {
		logger.Fatalf(ctx, "config err: %+v", err)
		return nil
	}

	dbIns, err := newInstances(ctx, conf)
	if err != nil {
		logger.Fatalf(ctx, "config init err: %+v", err)
		return nil
	}
	if err = dbIns.Use(tracing.NewPlugin(tracing.WithAttributes(
		attribute.String("db.connection_string", fmt.Sprintf("Server=%s;Port=%d,Database=%s;Uid=%s",
			conf.Host, conf.Port, conf.DBName, conf.User)),
	), tracing.WithoutMetrics())); err != nil {
		logger.Fatalf(ctx, "db tracing err: %+v", err)
		return nil
	}

	return dbIns
}

func (conf *Config) parseConf() error {
	if conf.Host == "" {
		return fmt.Errorf("postgres host empty")
	}
	if conf.Port == 0 {
		return fmt.Errorf("postgres port is zero")
	}
	if conf.User == "" {
		return fmt.Errorf("postgres user is empty")
	}
	if conf.PW == "" {
		return fmt.Errorf("postgres password is empty")
	}
	if conf.DBName == "" {
		return fmt.Errorf("postgres database name is empty")
	}

	if conf.Conns.MaxOpenConn == 0 {
		conf.Conns.MaxOpenConn = maxOpenConn
	}

	if conf.Conns.MaxIdleConn == 0 {
		conf.Conns.MaxIdleConn = maxIdleConn
	}

	if conf.Conns.TimeoutIdle == time.Duration(0) {
		conf.Conns.TimeoutIdle = timeoutIdel
	}

	return nil
}

func (conf *Config) LogLevel() gormLogger.LogLevel {
	switch conf.Level {
	case "debug", "info":
		return gormLogger.Info
	case "warn":
		return gormLogger.Warn
	case "error", "dpanic", "panic", "fatal":
		return gormLogger.Error
	default:
		return gormLogger.Error
	}
}

func (conf *Config) SDN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		conf.Host, conf.Port, conf.User, conf.PW, conf.DBName)
}

func newInstances(ctx context.Context, conf *Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(conf.SDN()), &gorm.Config{
		Logger: NewGormLogger(logger.BaseLogger(), conf.LogLevel(), conf.SlowThreshold),
	})
	if err != nil {
		logger.Fatalf(ctx, "sql: can't establish connection with mdsn: %s, err: %+v", conf.SDN(), err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalf(ctx, "sql: can't select db err: %+v", err)
		return nil, err
	}

	sqlDB.SetMaxIdleConns(conf.Conns.MaxIdleConn)
	sqlDB.SetMaxOpenConns(conf.Conns.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(conf.Conns.TimeoutIdle)

	return db, nil
}
