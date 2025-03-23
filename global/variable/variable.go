package variable

import (
	"context"
	"go.uber.org/zap"
)

// LogConfig 日志配置
type LogConfig struct {
	Path       string
	Mode       string
	Recover    bool
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Level      string
}

// ILog 日志接口
type ILog struct {
	*zap.Logger
	config *LogConfig
}

// Global 全局变量
var (
	// Db 全局数据库连接
	//Db *gorm.DB

	// Log 全局日志实例
	Log *ILog

	// ZapLog 全局日志实例
	ZapLog *zap.Logger
	// Ctx 全局上下文
	Ctx context.Context

	Echo *zap.SugaredLogger
)
