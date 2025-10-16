package migrate

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/repo/model"
	"gorm.io/gorm"
)

// MigrationRecord 迁移记录
type MigrationRecord struct {
	Version     string    `gorm:"column:version;primaryKey"`
	Applied     bool      `gorm:"column:applied;default:true"`
	AppliedTime time.Time `gorm:"column:applied_time;default:CURRENT_TIMESTAMP"`
}

func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// HandleAutoMigration 处理自动迁移逻辑
func HandleAutoMigration(ctx context.Context, config *webapp.WebGlobalConfig) error {
	if !config.Database.AutoMigrate {
		logger.Infof(ctx, "Database auto-migration is disabled")
		return nil
	}

	logger.Infof(ctx, "🔄 Starting database migration check...")

	// 获取当前迁移版本
	currentVersion, err := GetCurrentVersion(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get current migration version: %v", err)
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// 获取最新迁移版本
	latestVersion, err := GetLatestVersion(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get latest migration version: %v", err)
		return fmt.Errorf("failed to get latest migration version: %w", err)
	}

	logger.Infof(ctx, "📊 Current migration version: %s", currentVersion)
	logger.Infof(ctx, "📊 Latest migration version: %s", latestVersion)

	// 检查是否需要迁移
	if currentVersion == latestVersion {
		logger.Infof(ctx, "✅ Database is up to date")
		return nil
	}

	logger.Infof(ctx, "🚀 Starting database migrations from version %s to %s...", currentVersion, latestVersion)

	// 执行迁移
	if err := RunMigrations(ctx); err != nil {
		logger.Errorf(ctx, "❌ Migration failed: %v", err)
		return fmt.Errorf("migration failed: %w", err)
	}

	// 获取迁移后的版本确认
	finalVersion, err := GetCurrentVersion(ctx)
	if err != nil {
		logger.Warnf(ctx, "Failed to verify final migration version: %v", err)
	} else {
		logger.Infof(ctx, "✅ Migration completed successfully - Current version: %s", finalVersion)
	}

	return nil
}

// GetCurrentVersion 获取当前数据库迁移版本
func GetCurrentVersion(ctx context.Context) (string, error) {
	database := db.DB().DBIns()
	if database == nil {
		return "", fmt.Errorf("database connection not available")
	}

	// 确保迁移表存在
	if err := database.AutoMigrate(&MigrationRecord{}); err != nil {
		return "", fmt.Errorf("failed to create migration table: %w", err)
	}

	var latestRecord MigrationRecord
	err := database.Where("applied = ?", true).
		Order("version DESC").
		First(&latestRecord).Error

	if err != nil {
		// 使用GORM的错误检查方式
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "none", nil // 没有迁移记录
		}
		return "", fmt.Errorf("failed to get current version: %w", err)
	}

	return latestRecord.Version, nil
}

// GetLatestVersion 获取最新可用的迁移版本
func GetLatestVersion(ctx context.Context) (string, error) {
	// 基于当前模型结构生成版本号
	version := generateSchemaVersion()
	return version, nil
}

// generateSchemaVersion 基于模型结构生成版本号
func generateSchemaVersion() string {
	// 获取所有模型的类型信息
	models := []interface{}{
		&model.Laboratory{},
		&model.ResourceNodeTemplate{},
		&model.ResourceHandleTemplate{},
		&model.WorkflowNodeTemplate{},
		&model.MaterialNode{},
		&model.MaterialEdge{},
		&model.Workflow{},
		&model.WorkflowNode{},
		&model.WorkflowEdge{},
		&model.WorkflowConsole{},
		&model.WorkflowHandleTemplate{},
		&model.WorkflowNodeJob{},
		&model.WorkflowTask{},
		&model.Tags{},
		&model.LaboratoryMember{},
		&model.LaboratoryInvitation{},
	}

	// 生成所有模型的结构哈希
	var modelHashes []string
	for _, model := range models {
		modelType := reflect.TypeOf(model).Elem()
		hash := generateModelHash(modelType)
		modelHashes = append(modelHashes, fmt.Sprintf("%s:%s", modelType.Name(), hash))
	}

	// 排序确保一致性
	sort.Strings(modelHashes)

	// 生成总的哈希值
	allModels := ""
	for _, hash := range modelHashes {
		allModels += hash + ";"
	}

	hasher := md5.New()
	hasher.Write([]byte(allModels))
	schemaHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// 返回基于当前日期和哈希的版本号
	return fmt.Sprintf("schema_%s", schemaHash[:12])
}

// generateModelHash 生成单个模型的哈希
func generateModelHash(modelType reflect.Type) string {
	var fieldInfo []string

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		// 包含字段名、类型和gorm标签
		fieldStr := fmt.Sprintf("%s:%s", field.Name, field.Type.String())
		if gormTag := field.Tag.Get("gorm"); gormTag != "" {
			fieldStr += ":" + gormTag
		}
		fieldInfo = append(fieldInfo, fieldStr)
	}

	// 排序确保一致性
	sort.Strings(fieldInfo)

	allFields := ""
	for _, field := range fieldInfo {
		allFields += field + ";"
	}

	hasher := md5.New()
	hasher.Write([]byte(allFields))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:8]
}

// CheckMigrationStatus 检查迁移状态（向后兼容）
func CheckMigrationStatus(ctx context.Context) (bool, error) {
	currentVersion, err := GetCurrentVersion(ctx)
	if err != nil {
		return false, err
	}

	latestVersion, err := GetLatestVersion(ctx)
	if err != nil {
		return false, err
	}

	return currentVersion != latestVersion, nil
}

// RunMigrations 执行数据库迁移
func RunMigrations(ctx context.Context) error {
	logger.Infof(ctx, "🔄 Executing database schema migrations...")

	// 调用现有的迁移逻辑
	if err := Table(ctx); err != nil {
		return fmt.Errorf("table migration failed: %w", err)
	}

	// 记录迁移完成
	latestVersion, err := GetLatestVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest version after migration: %w", err)
	}

	// 更新迁移记录
	migrationRecord := MigrationRecord{
		Version:     latestVersion,
		Applied:     true,
		AppliedTime: time.Now(),
	}

	database := db.DB().DBIns()
	if database != nil {
		if err := database.Save(&migrationRecord).Error; err != nil {
			logger.Warnf(ctx, "Failed to record migration: %v", err)
		}
	}

	logger.Infof(ctx, "✅ Database schema migrations completed")
	return nil
}
