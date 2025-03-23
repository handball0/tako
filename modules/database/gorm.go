package database

import (
	"fmt"
	"github.com/handball0/tako/global/variable"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	SlowThreshold   int    // 慢查询阈值（毫秒）
	LogLevel        int    // 日志级别
	EnableLog       bool   // 是否开启日志
	MaxIdleConns    int    // 最大空闲连接数
	MaxOpenConns    int    // 最大连接数
	ConnMaxLifetime int    // 连接最大生命周期（秒）
	TablePrefix     string // 表前缀
	EnableLogWriter bool   // 是否开启日志写入
	UseSnowFlake    bool   // 是否使用雪花ID
}

// GormInstance GORM实例
type GormInstance struct {
	DB     *gorm.DB
	Config DatabaseConfig
}

func InitGormInstance() *GormInstance {
	variable.Echo.Info("✅ 初始化GORM实例")
	g := &GormInstance{}
	g.Config = DatabaseConfig{
		SlowThreshold:   viper.GetInt("Database.SlowThreshold"),
		LogLevel:        viper.GetInt("Database.LogLevel"),
		EnableLog:       viper.GetBool("Database.EnableLog"),
		MaxIdleConns:    viper.GetInt("Database.MaxIdleConns"),
		MaxOpenConns:    viper.GetInt("Database.MaxOpenConns"),
		ConnMaxLifetime: viper.GetInt("Database.ConnMaxLifetime"),
		TablePrefix:     viper.GetString("Database.TablePrefix"),
		EnableLogWriter: viper.GetBool("Database.EnableLogWriter"),
		UseSnowFlake:    viper.GetBool("Database.UseSnowFlake"),
	}
	return g
}

// CreateDbConnection 创建数据库连接
func (g *GormInstance) CreateDbConnection(dialector gorm.Dialector) (*gorm.DB, error) {
	if g.Config.EnableLog {
		db, err := gorm.Open(dialector, &gorm.Config{
			Logger: g.getLogger(),
		})
		if err != nil {
			variable.Echo.Error("❌ 创建数据库连接失败: %v", err)
			return nil, err
		}
		g.DB = db
		return db, nil
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		variable.Echo.Error("❌ 创建数据库连接失败: %v", err)
		return nil, err
	}

	if err := g.configureConnectionPool(db); err != nil {
		variable.Echo.Error("❌ 配置数据库连接池失败: %v", err)
		return nil, err
	}

	// 如果启用了雪花ID，添加插件
	if viper.GetBool("Database.UseSnowFlake") {
		if err := db.Use(&SnowflakePlugin{}); err != nil {
			return nil, fmt.Errorf("应用雪花ID失败: %w", err)
		}
		variable.Echo.Info("✅ 提示: 加载雪花ID插件成功...")
	}
	return db, nil
}

func (g *GormInstance) configureConnectionPool(db *gorm.DB) error {
	sqlDb, err := db.DB()
	if err != nil {
		variable.Echo.Error("❌ 获取数据库连接失败: %v", err)
		return err
	}

	sqlDb.SetMaxIdleConns(g.Config.MaxIdleConns)
	sqlDb.SetMaxOpenConns(g.Config.MaxOpenConns)

	if err := sqlDb.Ping(); err != nil {
		variable.Echo.Error("❌ 数据库连接失败: %v", err)
		return err
	}

	variable.Echo.Info("✅ 数据库连接成功")

	return nil
}

// getLogger 获取日志记录器
func (m *GormInstance) getLogger() logger.Interface {
	logLevel := viper.GetInt("PostgreSQL.LogLevel")
	logMode := m.getLogMode(logLevel)

	return logger.New(m.getLogWriter(), logger.Config{
		SlowThreshold:             time.Duration(viper.GetInt("PostgreSQL.SlowThreshold")) * time.Millisecond,
		LogLevel:                  logMode,
		IgnoreRecordNotFoundError: true,
		Colorful:                  !viper.GetBool("PostgreSQL.EnableLogWriter"),
	})
}

// getLogMode 获取日志级别
func (m *GormInstance) getLogMode(level int) logger.LogLevel {
	switch level {
	case 1:
		return logger.Error
	case 2:
		return logger.Warn
	default:
		return logger.Info
	}
}

// getLogWriter 获取日志写入器
func (m *GormInstance) getLogWriter() logger.Writer {
	logPath := viper.GetString("Log.Path")
	var writer io.Writer

	if viper.GetBool("PostgreSQL.EnableLogWriter") {
		writer = m.createFileWriter(logPath)
	} else {
		writer = os.Stdout
	}

	return log.New(writer, "\r\n", log.LstdFlags)
}

// createFileWriter 创建文件写入器
func (m *GormInstance) createFileWriter(logPath string) io.Writer {
	if !strings.HasSuffix(logPath, "/") {
		logPath += "/"
	}
	fileName := fmt.Sprintf("%s%s/pgsqlmodule.log", logPath, time.Now().Format("2006-01-02"))
	return &lumberjack.Logger{
		Filename:   path.Join(logPath, fileName),
		MaxSize:    viper.GetInt("Log.MaxSize"),
		MaxBackups: viper.GetInt("Log.MaxBackups"),
		MaxAge:     viper.GetInt("Log.MaxAge"),
		Compress:   viper.GetBool("Log.Compress"),
		LocalTime:  true,
	}
}
