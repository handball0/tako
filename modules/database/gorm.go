package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/handball0/tako/console"
	"github.com/handball0/tako/gint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Config 数据库配置
type Config struct {
	Type            string // 数据库类型: mysql, postgres, sqlite
	Host            string // 主机地址
	Port            int    // 端口
	Username        string // 用户名
	Password        string // 密码
	Database        string // 数据库名
	Charset         string // 字符集
	MaxIdleConn     int    // 最大空闲连接数
	MaxConn         int    // 最大连接数
	LogLevel        int    // 日志级别
	SlowThreshold   int    // 慢查询阈值（毫秒）
	EnableLogWriter bool   // 启用日志记录器
	UseSnowFlake    bool   // 使用雪花算法生成ID
	TablePrefix     string // 表前缀
}

func setDefaultConfig() {
	viper.SetDefault("Database.LogLevel", 3)
	viper.SetDefault("Database.EnableLogWriter", true)
	viper.SetDefault("Database.MaxIdleConn", 10)
	viper.SetDefault("Database.MaxConn", 200)
	viper.SetDefault("Database.SlowThreshold", 200)
	viper.SetDefault("Database.UseSnowFlake", false)
	viper.SetDefault("Database.Type", "mysql")
	viper.SetDefault("Database.Charset", "utf8mb4")
	viper.SetDefault("Database.TablePrefix", "")
}

var initDbCmd = &cobra.Command{
	Use:   "init-db",
	Short: "初始化数据库连接",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取logger
		logger := zap.NewExample().Sugar().Named("database")

		// 初始化数据库连接
		if err := InitDB(logger.Desugar()); err != nil {
			logger.Fatalw("数据库初始化失败", "error", err)
			return
		}

		logger.Infow("数据库连接成功",
			"type", viper.GetString("Database.Type"),
			"host", viper.GetString("Database.Host"),
			"database", viper.GetString("Database.Database"),
		)

		// 检查表结构
		if viper.GetBool("Database.AutoMigrate") {
			logger.Info("开始自动迁移数据库表结构...")
			// 这里可以添加自动迁移的代码
			// 如: gint.Db.AutoMigrate(&User{}, &Product{})
			logger.Info("数据库表结构迁移完成")
		}
	},
}

func init() {
	// 设置默认自动迁移选项
	viper.SetDefault("Database.AutoMigrate", false)
	// 设置默认自动初始化选项
	viper.SetDefault("Database.AutoInit", false)

	// 注册init-db命令（如果有需要，可以在应用程序中手动添加这个命令）
	console.AppendCommand(initDbCmd, 1)
}

// // AutoInit 在应用启动时自动初始化数据库
// // 可以在应用程序的main函数中调用此方法
// func AutoInit() error {
// 	if !viper.GetBool("Database.AutoInit") {
// 		return nil
// 	}

// 	logger := zap.NewExample().Sugar().Named("database")
// 	if err := InitDB(logger.Desugar()); err != nil {
// 		logger.Warnw("数据库自动初始化失败", "error", err)
// 		return err
// 	}

// 	logger.Info("数据库自动初始化成功")
// 	return nil
// }

// // GetInitDbCmd 获取数据库初始化命令
// // 可以在应用程序中将此命令添加到根命令中
// func GetInitDbCmd() *cobra.Command {
// 	return initDbCmd
// }

// InitDB 初始化数据库连接
func InitDB(zapLogger *zap.Logger) error {
	setDefaultConfig()

	config := Config{
		Type:            viper.GetString("Database.Type"),
		Host:            viper.GetString("Database.Host"),
		Port:            viper.GetInt("Database.Port"),
		Username:        viper.GetString("Database.Username"),
		Password:        viper.GetString("Database.Password"),
		Database:        viper.GetString("Database.Database"),
		Charset:         viper.GetString("Database.Charset"),
		MaxIdleConn:     viper.GetInt("Database.MaxIdleConn"),
		MaxConn:         viper.GetInt("Database.MaxConn"),
		LogLevel:        viper.GetInt("Database.LogLevel"),
		SlowThreshold:   viper.GetInt("Database.SlowThreshold"),
		EnableLogWriter: viper.GetBool("Database.EnableLogWriter"),
		UseSnowFlake:    viper.GetBool("Database.UseSnowFlake"),
		TablePrefix:     viper.GetString("Database.TablePrefix"),
	}

	db, err := Connect(config, zapLogger)
	if err != nil {
		return err
	}
	// 设置全局 DB 变量
	gint.Db = db
	return nil
}

// Connect 连接数据库
func Connect(config Config, zapLogger *zap.Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector
	var dsn string

	switch config.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host, config.Username, config.Password, config.Database, config.Port)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dsn = fmt.Sprintf("%s.db", config.Database)
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// 配置 GORM 日志
	var gormLogger logger.Interface
	if config.EnableLogWriter {
		gormLogger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Duration(config.SlowThreshold) * time.Millisecond,
				LogLevel:      logger.LogLevel(config.LogLevel),
				Colorful:      true,
			},
		)
	} else if zapLogger != nil {
		gormLogger = NewGormZapLogger(zapLogger)
	}

	// 创建 GORM 配置
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   config.TablePrefix,
			SingularTable: true, // 使用单数表名
		},
	}

	// 连接数据库
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 *sql.DB 失败: %w", err)
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConn)
	sqlDB.SetMaxOpenConns(config.MaxConn)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Close 关闭数据库连接
func Close() error {
	if gint.Db == nil {
		return nil
	}

	sqlDB, err := gint.Db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GormZapLogger Zap 日志适配器
type GormZapLogger struct {
	ZapLogger *zap.Logger
	LogLevel  logger.LogLevel
}

// NewGormZapLogger 创建 Zap 日志适配器
func NewGormZapLogger(zapLogger *zap.Logger) logger.Interface {
	return &GormZapLogger{
		ZapLogger: zapLogger,
		LogLevel:  logger.LogLevel(viper.GetInt("Database.LogLevel")),
	}
}

// LogMode 设置日志级别
func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 记录信息日志
func (l GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn 记录警告日志
func (l GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error 记录错误日志
func (l GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace 记录 SQL 跟踪日志
func (l GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error:
		sql, rows := fc()
		l.ZapLogger.Sugar().Errorw("SQL 执行错误",
			"error", err,
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	case elapsed > time.Duration(viper.GetInt("Database.SlowThreshold"))*time.Millisecond && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		l.ZapLogger.Sugar().Warnw("慢查询 SQL",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	case l.LogLevel >= logger.Info:
		sql, rows := fc()
		l.ZapLogger.Sugar().Infow("SQL 执行",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	}
}

// 事务相关辅助函数

// Transaction 执行事务
func Transaction(fc func(tx *gorm.DB) error) error {
	return gint.Db.Transaction(fc)
}

// WithTransaction 在事务中执行操作
func WithTransaction(tx *gorm.DB, fc func(tx *gorm.DB) error) error {
	if tx == nil {
		return Transaction(fc)
	}
	return fc(tx)
}
