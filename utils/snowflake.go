package utils

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// InitSnowflake 初始化雪花ID生成器
func InitSnowflake(nodeID int64) error {
	var err error
	once.Do(func() {
		node, err = snowflake.NewNode(nodeID)
	})
	return err
}

// GenerateSnowflakeID 生成雪花ID
func GenerateSnowflakeID() int64 {
	if node == nil {
		if err := InitSnowflake(1); err != nil {
			panic(fmt.Sprintf("初始化雪花ID生成器失败: %v", err))
		}
	}
	return node.Generate().Int64()
}

// GenerateSnowflakeString 生成雪花ID字符串
func GenerateSnowflakeString() string {
	if node == nil {
		if err := InitSnowflake(1); err != nil {
			panic(fmt.Sprintf("初始化雪花ID生成器失败: %v", err))
		}
	}
	return node.Generate().String()
}

// ParseSnowflakeID 解析雪花ID
func ParseSnowflakeID(id int64) snowflake.ID {
	return snowflake.ParseInt64(id)
}

// ParseSnowflakeString 解析雪花ID字符串
func ParseSnowflakeString(id string) (snowflake.ID, error) {
	return snowflake.ParseString(id)
}

// GetSnowflakeTime 获取雪花ID的时间戳
func GetSnowflakeTime(id int64) int64 {
	return ParseSnowflakeID(id).Time()
}

// GetSnowflakeNodeID 获取雪花ID的节点ID
func GetSnowflakeNodeID(id int64) int64 {
	return ParseSnowflakeID(id).Node()
}

// GetSnowflakeSequence 获取雪花ID的序列号
func GetSnowflakeSequence(id int64) int64 {
	return ParseSnowflakeID(id).Step()
}

// IsValidSnowflakeID 检查是否是有效的雪花ID
func IsValidSnowflakeID(id int64) bool {
	return ParseSnowflakeID(id).String() != ""
}

// GenerateSnowflakeIDs 批量生成雪花ID
func GenerateSnowflakeIDs(count int) []int64 {
	if count <= 0 {
		return nil
	}

	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		ids[i] = GenerateSnowflakeID()
	}
	return ids
}

// GenerateSnowflakeStrings 批量生成雪花ID字符串
func GenerateSnowflakeStrings(count int) []string {
	if count <= 0 {
		return nil
	}

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = GenerateSnowflakeString()
	}
	return ids
}
