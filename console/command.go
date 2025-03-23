package console

//
//import (
//	"fmt"
//	"github.com/handball0/tako/global/variable"
//	"sort"
//	"sync"
//	"time"
//
//	"github.com/spf13/cobra"
//)
//
//var (
//	//Echo       *zap.SugaredLogger
//	mu         sync.RWMutex
//	mapCommand = make(map[string]*CommandInfo)
//)
//
//type runFunc func(cmd *cobra.Command, args []string)
//
//// CommandStatus 命令状态
//type CommandStatus int
//
//const (
//	CommandStatusPending CommandStatus = iota
//	CommandStatusRunning
//	CommandStatusCompleted
//	CommandStatusFailed
//	CommandStatusTimeout
//)
//
//// CommandInfo 命令信息
//type CommandInfo struct {
//	Command   *cobra.Command
//	Run       runFunc
//	Status    CommandStatus
//	Priority  int
//	Error     error
//	StartTime time.Time
//	EndTime   time.Time
//}
//
//func init() {
//	// 初始化日志
//	//Echo = zap.New(zap.NewNop().Core(), zap.AddCaller()).Sugar()
//	//Echo = utils.InitSugaredLogger()
//}
//
//var RootCmd = &cobra.Command{
//	Use:   "Root",
//	Short: "go gin frame",
//	Long:  `Web project scaffolding based on go+gin framework`,
//	Run: func(cmd *cobra.Command, args []string) {
//		if err := executeCommands(cmd, args); err != nil {
//			variable.Echo.Fatalw("❌ 错误: 命令执行失败", "error", err)
//		}
//
//	},
//}
//
//// Append 添加命令 不指定优先级 默认优先级为50
//func Append(cmdList ...*cobra.Command) {
//	mu.Lock()
//	defer mu.Unlock()
//	for _, cmd := range cmdList {
//		AppendCommand(cmd, 50)
//	}
//}
//
//// AppendCommand 添加命令
//func AppendCommand(cmd *cobra.Command, priority int) {
//	mu.Lock()
//	defer mu.Unlock()
//
//	RootCmd.AddCommand(cmd)
//	mapCommand[cmd.Name()] = &CommandInfo{
//		Command:  cmd,
//		Run:      cmd.Run,
//		Priority: priority,
//	}
//}
//
//func executeCommand(cmdInfo *CommandInfo, rootCmd *cobra.Command, args []string) error {
//
//	errChan := make(chan error, 1)
//	cmdInfo.Status = CommandStatusRunning
//	cmdInfo.StartTime = time.Now()
//	cmdInfo.Error = nil
//	go func() {
//		defer func() {
//			if r := recover(); r != nil {
//				errChan <- fmt.Errorf("命令执行: %s 错误: %v", cmdInfo.Command.Name(), r)
//			}
//		}()
//		cmdInfo.Run(rootCmd, args)
//		errChan <- nil
//	}()
//
//	err := <-errChan
//	cmdInfo.EndTime = time.Now()
//	if err != nil {
//		cmdInfo.Status = CommandStatusFailed
//		cmdInfo.Error = err
//		variable.Echo.Error("命令执行失败", "command", cmdInfo.Command.Use, "error", err, "duration", cmdInfo.EndTime.Sub(cmdInfo.StartTime))
//	} else {
//		cmdInfo.Status = CommandStatusCompleted
//		variable.Echo.Info("命令执行成功", "command", cmdInfo.Command.Use, "duration", cmdInfo.EndTime.Sub(cmdInfo.StartTime))
//	}
//	close(errChan)
//	return nil
//}
//
//func executeCommands(cmd *cobra.Command, args []string) error {
//	mu.RLock()
//	defer mu.RUnlock()
//
//	if mapCommand["Init"] == nil {
//		return fmt.Errorf("请务必在入口函数 `main()` 中通过 `_ github.com/handball0/tako/gint` 加载tako模块")
//	}
//
//	commands := sortCommandsByPriority()
//	for _, cmdInfo := range commands {
//		fmt.Printf("执行命令: %s\n", cmdInfo.Command.Use)
//		if err := executeCommand(cmdInfo, cmd, args); err != nil {
//			return fmt.Errorf("执行命令失败: %s, 错误: %v", cmdInfo.Command.Use, err)
//		}
//	}
//
//	return nil
//}
//
//// sortCommandsByPriority 按优先级排序
//func sortCommandsByPriority() []*CommandInfo {
//	// 创建一个CommandInfo切片用于排序
//	commands := make([]*CommandInfo, 0, len(mapCommand))
//	for _, cmd := range mapCommand {
//		commands = append(commands, cmd)
//	}
//
//	// 按优先级排序
//	sort.Slice(commands, func(i, j int) bool {
//		return commands[i].Priority < commands[j].Priority
//	})
//
//	return commands
//}
