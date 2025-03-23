package log

//
//import (
//	"bufio"
//	"context"
//	"fmt"
//
//	"github.com/handball0/tako/global/variable"
//
//	"io"
//	"os"
//	"path/filepath"
//	"strings"
//	"sync"
//	"time"
//
//	"github.com/spf13/viper"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//	"gopkg.in/natefinch/lumberjack.v2"
//)
//
//// InitILog 初始化日志
//func InitILog() error {
//	// 设置默认配置
//	viper.SetDefault("Log.Path", "./logs")
//	viper.SetDefault("Log.Mode", "both")
//	viper.SetDefault("Log.Recover", false)
//	viper.SetDefault("Log.MaxSize", 100)
//	viper.SetDefault("Log.MaxBackups", 3)
//	viper.SetDefault("Log.MaxAge", 7)
//	viper.SetDefault("Log.Compress", true)
//	viper.SetDefault("Log.Level", "info")
//
//	// 创建日志配置
//	config := &variable.LogConfig{
//		Path:       viper.GetString("Log.Path"),
//		Mode:       viper.GetString("Log.Mode"),
//		Recover:    viper.GetBool("Log.Recover"),
//		MaxSize:    viper.GetInt("Log.MaxSize"),
//		MaxBackups: viper.GetInt("Log.MaxBackups"),
//		MaxAge:     viper.GetInt("Log.MaxAge"),
//		Compress:   viper.GetBool("Log.Compress"),
//		Level:      viper.GetString("Log.Level"),
//	}
//
//	// 创建日志目录
//	if err := os.MkdirAll(config.Path, 0755); err != nil {
//		return fmt.Errorf("创建日志目录失败: %w", err)
//	}
//
//	// 创建日志实例InitILog
//	variable.Log = &variable.ILog{
//		Logger: zap.New(
//			getCore(config),
//			zap.AddCaller(),
//			zap.AddCallerSkip(0),
//			zap.AddStacktrace(zap.ErrorLevel),
//		),
//		config: config,
//	}
//
//	return nil
//}
//
//// With 添加字段
//func (l *variable.ILog) With(fields ...zap.Field) *variable.ILog {
//	return &variable.ILog{
//		Logger: l.Logger.With(fields...),
//		config: l.config,
//	}
//}
//
//// WithCtx 添加上下文
//func (l *ILog) WithCtx(ctx context.Context) *ILog {
//	var traceIdStr, sourceStr string
//	traceId := ctx.Value("trace_id")
//	if traceId != nil {
//		traceIdStr, _ = traceId.(string)
//	}
//
//	source := ctx.Value("source")
//	if source != nil {
//		sourceStr, _ = source.(string)
//	}
//
//	if traceIdStr != "" {
//		l.With(zap.String("trace_id", traceIdStr))
//	}
//	if sourceStr != "" {
//		l.With(zap.String("source", sourceStr))
//	}
//
//	return l
//}
//
//// getCore 获取日志核心
//func getCore(config *LogConfig) zapcore.Core {
//	encoder := zapcore.NewJSONEncoder(getEncoderConfig())
//	debugWrite := getLogWriter(config, zapcore.DebugLevel)
//	infoWrite := getLogWriter(config, zapcore.InfoLevel)
//	warnWrite := getLogWriter(config, zapcore.WarnLevel)
//	errorWrite := getLogWriter(config, zapcore.ErrorLevel)
//
//	debugLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
//		return level == zapcore.DebugLevel
//	})
//	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
//		return level == zapcore.InfoLevel
//	})
//	warnLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
//		return level == zapcore.WarnLevel
//	})
//	errorLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
//		return level == zapcore.ErrorLevel
//	})
//
//	return zapcore.NewTee(
//		zapcore.NewCore(encoder, zapcore.AddSync(debugWrite), debugLevel),
//		zapcore.NewCore(encoder, zapcore.AddSync(infoWrite), infoLevel),
//		zapcore.NewCore(encoder, zapcore.AddSync(warnWrite), warnLevel),
//		zapcore.NewCore(encoder, zapcore.AddSync(errorWrite), errorLevel),
//	)
//}
//
//// getLogWriter 获取日志写入器
//func getLogWriter(config *LogConfig, level zapcore.Level) zapcore.WriteSyncer {
//	if !strings.HasSuffix(config.Path, "/") {
//		config.Path += "/"
//	}
//
//	// 创建日期目录
//	dateDir := time.Now().Format("2006-01-02")
//	dir := filepath.Join(config.Path, dateDir)
//	if err := os.MkdirAll(dir, 0755); err != nil {
//		variable.Echo.Errorf("❌ 错误: 创建日志目录失败: %v", err)
//		return zapcore.AddSync(io.Discard)
//	}
//
//	// 创建日志文件
//	fileName := filepath.Join(dir, fmt.Sprintf("%s.log", level))
//	var fileWriter io.Writer
//
//	if config.Recover {
//		fileWriter = NewCustomWrite(fileName, config.MaxSize, config.MaxBackups, config.MaxAge, config.Compress)
//	} else {
//		fileWriter = &lumberjack.Logger{
//			Filename:   fileName,
//			MaxSize:    config.MaxSize,    // 单文件最大容量, 单位是MB
//			MaxBackups: config.MaxBackups, // 最大保留过期文件个数
//			MaxAge:     config.MaxAge,     // 保留过期文件的最大时间间隔, 单位是天
//			Compress:   config.Compress,   // 是否需要压缩滚动日志, 使用的gzip压缩
//			LocalTime:  true,              // 是否使用计算机的本地时间, 默认UTC
//		}
//	}
//
//	var writer zapcore.WriteSyncer
//	switch config.Mode {
//	case "file":
//		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter))
//	case "console":
//		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
//	case "close":
//		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(io.Discard))
//	default:
//		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter))
//	}
//
//	return writer
//}
//
//// getEncoderConfig 获取编码器配置
//func getEncoderConfig() zapcore.EncoderConfig {
//	return zapcore.EncoderConfig{
//		TimeKey:        "timestamp",
//		LevelKey:       "level",
//		NameKey:        "log",
//		CallerKey:      "file_line",
//		FunctionKey:    zapcore.OmitKey,
//		MessageKey:     "msg",
//		StacktraceKey:  "stack",
//		LineEnding:     zapcore.DefaultLineEnding,
//		EncodeLevel:    zapcore.LowercaseLevelEncoder,
//		EncodeTime:     customTimeEncoder,
//		EncodeDuration: zapcore.SecondsDurationEncoder,
//		EncodeCaller:   zapcore.ShortCallerEncoder,
//		EncodeName:     zapcore.FullNameEncoder,
//	}
//}
//
//// customTimeEncoder 自定义时间编码器
//func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
//	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
//}
//
//// CustomWrite 自定义写入器
//type CustomWrite struct {
//	mu       sync.Mutex
//	filepath string
//	logger   *lumberjack.Logger
//	inner    zapcore.WriteSyncer
//	buffer   *bufio.Writer
//	done     chan struct{}
//}
//
//// NewCustomWrite 创建自定义写入器
//func NewCustomWrite(filepath string, maxSize, maxBackups, maxAge int, compress bool) *CustomWrite {
//	cw := &CustomWrite{
//		filepath: filepath,
//		done:     make(chan struct{}),
//	}
//	cw.initLogger(filepath, maxSize, maxBackups, maxAge, compress)
//
//	// 启动文件状态监控
//	go cw.monitorFile()
//
//	return cw
//}
//
//// initLogger 初始化日志文件和写入器
//func (cw *CustomWrite) initLogger(filepath string, maxSize, maxBackups, maxAge int, compress bool) {
//	cw.logger = &lumberjack.Logger{
//		Filename:   filepath,
//		MaxSize:    maxSize,
//		MaxBackups: maxBackups,
//		MaxAge:     maxAge,
//		Compress:   compress,
//	}
//	cw.inner = zapcore.AddSync(cw.logger)
//	cw.buffer = bufio.NewWriterSize(cw.inner, 4096)
//}
//
//// Write 写入日志
//func (cw *CustomWrite) Write(p []byte) (n int, err error) {
//	cw.mu.Lock()
//	defer cw.mu.Unlock()
//
//	// 写入缓冲区
//	n, err = cw.buffer.Write(p)
//	if err != nil {
//		cw.recreateLogger()
//		return n, err
//	}
//
//	// 刷新缓冲区
//	err = cw.flushBuffer()
//	if err != nil {
//		cw.recreateLogger()
//	}
//	return n, err
//}
//
//// recreateLogger 重新创建日志文件和写入器
//func (cw *CustomWrite) recreateLogger() {
//	cw.logger.Close()
//	cw.initLogger(cw.filepath, cw.logger.MaxSize, cw.logger.MaxBackups, cw.logger.MaxAge, cw.logger.Compress)
//}
//
//// flushBuffer 刷新缓冲区
//func (cw *CustomWrite) flushBuffer() error {
//	return cw.buffer.Flush()
//}
//
//// monitorFile 异步监控日志文件状态
//func (cw *CustomWrite) monitorFile() {
//	ticker := time.NewTicker(5 * time.Second)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			cw.checkFile()
//		case <-cw.done:
//			return
//		}
//	}
//}
//
//// checkFile 检查日志文件是否存在
//func (cw *CustomWrite) checkFile() {
//	if _, err := os.Stat(cw.filepath); os.IsNotExist(err) {
//		cw.recreateLogger()
//	}
//}
//
//// Sync 同步日志
//func (cw *CustomWrite) Sync() error {
//	cw.mu.Lock()
//	defer cw.mu.Unlock()
//
//	return cw.flushBuffer()
//}
//
//// Close 关闭日志
//func (cw *CustomWrite) Close() error {
//	close(cw.done)
//	cw.mu.Lock()
//	defer cw.mu.Unlock()
//
//	if err := cw.flushBuffer(); err != nil {
//		return err
//	}
//
//	return cw.logger.Close()
//}
