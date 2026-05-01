package utils

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	sfNode *snowflake.Node
	sfOnce sync.Once
)

func InitSnowflake(nodeID int64) error {
	var initErr error
	sfOnce.Do(func() {
		node, err := snowflake.NewNode(nodeID)
		if err != nil {
			initErr = fmt.Errorf("初始化雪花算法节点失败: %w", err)
			return
		}
		sfNode = node
	})
	return initErr
}

func GenerateSnowflakeID() int64 {
	if sfNode == nil {
		if err := InitSnowflake(1); err != nil {
			return 0
		}
	}
	return sfNode.Generate().Int64()
}

func GenerateSnowflakeString() string {
	id := GenerateSnowflakeID()
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("%d", id)
}
