package database

import (
	"fmt"
	"github.com/handball0/tako/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// SnowflakePlugin 是一个 GORM 插件，用于自动生成雪花 ID
type SnowflakePlugin struct{}

// Name 返回插件名称
func (sp *SnowflakePlugin) Name() string {
	return "SnowflakePlugin"
}

// Initialize 初始化插件
func (sp *SnowflakePlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gorm:create").Register("snowflake:before_create", func(db *gorm.DB) {
		if err := sp.BeforeCreate(db); err != nil {
			db.AddError(err)
		}
	}); err != nil {
		return fmt.Errorf("注册雪花ID插件失败: %w", err)
	}
	return nil
}

// BeforeCreate 在创建记录前生成雪花 ID
func (sp *SnowflakePlugin) BeforeCreate(db *gorm.DB) error {
	if db.Statement.Schema == nil {
		return nil
	}

	for _, field := range db.Statement.Schema.Fields {
		// 检查字段是否满足雪花ID生成条件：
		// 1. 字段名为 "id" 且是主键
		// 2. 字段标签中包含 "gorm:SnowFlake"
		if !isSnowflakeField(field) {
			continue
		}

		// 如果字段为空（零值），则生成雪花 ID
		if shouldGenerateSnowflakeID(field, db) {
			if err := setSnowflakeID(field, db); err != nil {
				return fmt.Errorf("设置雪花ID失败: %w", err)
			}
		}
	}

	return nil
}

// isSnowflakeField 检查字段是否应该使用雪花ID
func isSnowflakeField(field *schema.Field) bool {
	return (field.DBName == "id" && field.PrimaryKey) || field.TagSettings["GORM"] == "SnowFlake"
}

// shouldGenerateSnowflakeID 检查是否需要生成雪花ID
func shouldGenerateSnowflakeID(field *schema.Field, db *gorm.DB) bool {
	v, ok := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
	return ok && v == int64(0)
}

// setSnowflakeID 为字段设置雪花ID
func setSnowflakeID(field *schema.Field, db *gorm.DB) error {
	return field.Set(db.Statement.Context, db.Statement.ReflectValue, utils.GenerateSnowflakeID())
}
