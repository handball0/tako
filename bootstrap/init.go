package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"github.com/handball0/tako/global/variable"
	"github.com/handball0/tako/utils"
	"github.com/handball0/tako/utils/zap_factory"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
	envPrefix  = "TAKO"

	// Cancel 全局上下文取消函数
	Cancel context.CancelFunc

	// Wg 全局等待组，用于等待所有goroutine完成
	Wg sync.WaitGroup

	// GlobalMutex 全局读写锁
	GlobalMutex sync.RWMutex

	// 退出回调相关
	exitCallbacks   []func()
	exitCallbacksMu sync.Mutex
	isExiting       bool
	isExitingMu     sync.Mutex
)

func init() {
	variable.Echo = utils.InitSugaredLogger()
	variable.ZapLog = zap_factory.CreateZapFactory(zap_factory.ZapLogHandler)
	TakoCmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "config file (default is $HOME/.tako/config.yaml)")
	// console.AppendCommand(TakoCmd, 0)
	OnInit()
}

func OnInit() {

	// 优雅的关闭资源
	defer GracefulShutdown()

	if err := initConfig(); err != nil {
		variable.Echo.Fatalw("❌ 错误: 初始化配置失败", "error", err)
	}

	//if err := log.InitILog(); err != nil {
	//	variable.Echo.Fatalw("❌ 错误: 初始化日志失败", "error", err)
	//}
}

var TakoCmd = &cobra.Command{
	Use:   "Init",
	Short: "tako框架初始化",
	Long:  `tako框架初始化`,
	Run: func(cmd *cobra.Command, args []string) {
		OnInit()
	},
}

func initConfig() error {
	if configFile == "" {
		configFile = "./config.yaml"
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置配置文件路径
	viper.SetConfigFile(configFile)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 检查是否是因为文件不存在导致的错误
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			variable.Echo.Warnf("❌ 配置文件 %s 不存在，将创建默认配置文件", configFile)
			if err := createDefaultConfig(); err != nil {
				return fmt.Errorf("❌ 创建默认配置文件失败: %w", err)
			}
			// 重新加载配置
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("❌ 读取新创建的配置文件失败: %w", err)
			}
			variable.Echo.Infof("✅ 已创建并加载默认配置文件: %s", configFile)
		}
	}

	// 验证必要配置
	if err := validateConfig(); err != nil {
		return fmt.Errorf("❌ 配置验证失败: %w", err)
	}

	// 监听配置文件变化
	viper.WatchConfig()

	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	// 设置默认配置值
	setDefaultConfigs()

	// 确保目录存在
	dir := getConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("❌ 创建配置目录失败: %w", err)
	}

	// 写入配置文件
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("❌ 写入配置文件失败: %w", err)
	}

	return nil
}

// getConfigDir 获取配置文件所在目录
func getConfigDir() string {
	dir := configFile
	lastSlash := strings.LastIndexAny(dir, "/\\")
	if lastSlash != -1 {
		dir = dir[:lastSlash]
	} else {
		dir = "."
	}
	return dir
}

// setDefaultConfigs 设置默认配置
func setDefaultConfigs() {
	// 设置应用相关默认配置
	viper.SetDefault("App.Name", "Tako应用")
	viper.SetDefault("App.Version", "1.0.0")
	viper.SetDefault("App.Mode", "development") // development, production, testing
	viper.SetDefault("App.Port", 8080)
	viper.SetDefault("App.Debug", true)

	// 设置日志相关默认配置
	viper.SetDefault("Log.Path", "logs/")
	viper.SetDefault("Log.Mode", "dev") // dev, file, console, close
	viper.SetDefault("Log.Level", "debug")
	viper.SetDefault("Log.MaxSize", 100)
	viper.SetDefault("Log.MaxBackups", 3)
	viper.SetDefault("Log.MaxAge", 30)
	viper.SetDefault("Log.Compress", false)

	// 设置数据库相关默认配置
	viper.SetDefault("Database.EnableLog", true)
	viper.SetDefault("Database.LogLevel", 3)
	viper.SetDefault("Database.EnableLogWriter", true)
	viper.SetDefault("Database.MaxIdleConns", 10)
	viper.SetDefault("Database.MaxOpenConns", 100)
	viper.SetDefault("Database.SlowThreshold", 200)
	viper.SetDefault("Database.UseSnowFlake", false)
	viper.SetDefault("Database.TablePrefix", "tk_")

	// 可以根据需要添加更多默认配置

	variable.Echo.Info("✅ 设置默认配置成功")
}

// GracefulShutdown 优雅关闭资源
func GracefulShutdown() {
	// 触发所有退出回调
	triggerExitCallbacks()

	// 通知所有使用全局上下文的goroutine退出
	if Cancel != nil {
		Cancel()
	}

	// 等待所有goroutine完成
	Wg.Wait()

	// 关闭数据库连接
	//closeDatabase()

	// 同步日志缓冲区
	flushLogs()
}

// RegisterExitCallback 注册一个在程序退出时要执行的回调函数
func RegisterExitCallback(callback func()) {
	exitCallbacksMu.Lock()
	defer exitCallbacksMu.Unlock()
	exitCallbacks = append(exitCallbacks, callback)
}

// triggerExitCallbacks 触发所有退出回调函数
func triggerExitCallbacks() {
	// 检查是否已经在退出过程中
	isExitingMu.Lock()
	if isExiting {
		isExitingMu.Unlock()
		return
	}
	isExiting = true
	isExitingMu.Unlock()

	// 执行所有退出回调
	exitCallbacksMu.Lock()
	defer exitCallbacksMu.Unlock()

	for _, callback := range exitCallbacks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					variable.Echo.Errorf("退出回调执行失败: %v", r)
				}
			}()
			callback()
		}()
	}

	// 清空回调列表
	exitCallbacks = nil
}

// closeDatabase 关闭数据库连接
//func closeDatabase() {
//	if variable.Db == nil {
//		return
//	}
//
//	sqlDB, err := variable.Db.DB()
//	if err != nil {
//		variable.Echo.Errorf("获取SQL.DB失败: %v", err)
//		return
//	}
//
//	if err := sqlDB.Close(); err != nil {
//		variable.Echo.Errorf("关闭数据库连接失败: %v", err)
//	} else {
//		variable.Echo.Info("数据库连接已关闭")
//	}
//}

// flushLogs 刷新日志
func flushLogs() {
	if variable.Log == nil {
		return
	}

	if err := variable.Log.Sync(); err != nil {
		fmt.Printf("同步日志失败: %v\n", err)
	} else {
		variable.Echo.Info("日志已同步")
	}
}

// 验证必要的配置文件
func validateConfig() error {
	// 验证配置文件
	if !viper.IsSet("App") {
		return fmt.Errorf("❌ 未设置App配置")
	}

	return nil
}

// GetConfig 获取配置
func GetConfig(key string) interface{} {
	return viper.Get(key)
}

// GetConfigString 获取字符串配置
func GetConfigString(key string) string {
	return viper.GetString(key)
}

// GetConfigInt 获取整数配置
func GetConfigInt(key string) int {
	return viper.GetInt(key)
}

// GetConfigBool 获取布尔配置
func GetConfigBool(key string) bool {
	return viper.GetBool(key)
}

// GetConfigStringSlice 获取字符串切片配置
func GetConfigStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetConfigMap 获取Map配置
func GetConfigMap(key string) map[string]interface{} {
	return viper.GetStringMap(key)
}

// SetConfig 设置配置
func SetConfig(key string, value interface{}) {
	viper.Set(key, value)
}

// IsSet 检查配置是否存在
func IsSet(key string) bool {
	return viper.IsSet(key)
}

// GetEnv 获取环境变量
func GetEnv(key string) string {
	return os.Getenv(envPrefix + "_" + key)
}

// SetEnv 设置环境变量
func SetEnv(key, value string) error {
	return os.Setenv(envPrefix+"_"+key, value)
}
