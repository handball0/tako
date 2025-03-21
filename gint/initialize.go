package gint

import (
	"fmt"
	"os"
	"strings"

	"github.com/handball0/tako/console"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
	envPrefix  = "TAKO"
)

func init() {
	console.RootCmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "config file (default is $HOME/.tako/config.yaml)")
	console.AppendCommand(TakoCmd, 0)
}

var TakoCmd = &cobra.Command{
	Use:   "Gint",
	Short: "tako框架初始化",
	Long:  `tako框架初始化`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			console.Echo.Fatalw("❌ 错误: 初始化配置失败", "error", err)
		}

		if err := initILog(); err != nil {
			console.Echo.Fatalw("❌ 错误: 初始化日志失败", "error", err)
		}
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
		return fmt.Errorf("读取配置文件错误: %s", err)
	}

	// 验证必要配置
	if err := validateConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %s", err)
	}

	// 监听配置文件变化
	viper.WatchConfig()

	return nil
}

// 验证必要的配置文件
func validateConfig() error {
	// 验证配置文件
	if !viper.IsSet("App") {
		return fmt.Errorf("未设置App配置")
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
