package snowflake

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

// Generator 雪花算法 ID 生成器（封装第三方库）
type Generator struct {
	node *snowflake.Node
	mu   sync.Mutex // 保护 node 访问的互斥锁
}

var (
	defaultGenerator *Generator
	once             sync.Once
)

// NewGenerator 创建新的雪花算法生成器
// nodeID: 节点ID，范围 0-1023（10位）
func NewGenerator(nodeID int64) (*Generator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	return &Generator{
		node: node,
	}, nil
}

// InitDefault 初始化默认的全局生成器（单例模式）
func InitDefault(nodeID int64) error {
	var err error
	once.Do(func() {
		defaultGenerator, err = NewGenerator(nodeID)
	})
	return err
}

// GetDefaultGenerator 获取默认的全局生成器
func GetDefaultGenerator() *Generator {
	if defaultGenerator == nil {
		// 如果没有初始化，使用默认节点 ID 1
		_ = InitDefault(1)
	}
	return defaultGenerator
}

// Generate 生成唯一的 int64 ID
func (g *Generator) Generate() int64 {
	return g.node.Generate().Int64()
}

// GenerateUint 生成 uint64 类型的 ID（适用于 GORM 的 uint64 主键）
func (g *Generator) GenerateUint() uint64 {
	return uint64(g.node.Generate().Int64())
}

// GenerateString 生成字符串形式的 ID
func (g *Generator) GenerateString() string {
	return g.node.Generate().String()
}

// GenerateBase2 生成二进制字符串形式的 ID
func (g *Generator) GenerateBase2() string {
	return g.node.Generate().Base2()
}

// GenerateBase64 生成 Base64 编码的 ID
func (g *Generator) GenerateBase64() string {
	return g.node.Generate().Base64()
}

// GenerateBatch 批量生成 ID
func (g *Generator) GenerateBatch(count int) []int64 {
	if count <= 0 {
		return nil
	}

	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		ids[i] = g.Generate()
	}
	return ids
}

// GenerateBatchUint 批量生成 uint64 类型的 ID
func (g *Generator) GenerateBatchUint(count int) []uint64 {
	if count <= 0 {
		return nil
	}

	ids := make([]uint64, count)
	for i := 0; i < count; i++ {
		ids[i] = g.GenerateUint()
	}
	return ids
}

// ParseID 解析 ID，返回详细信息
func (g *Generator) ParseID(id int64) *IDInfo {
	sfID := snowflake.ParseInt64(id)
	return &IDInfo{
		ID:        id,
		Timestamp: sfID.Time(), // sfID.Time() 返回的就是毫秒时间戳
		Node:      sfID.Node(),
		Step:      sfID.Step(),
	}
}

// IDInfo ID 详细信息
type IDInfo struct {
	ID        int64 // 原始 ID
	Timestamp int64 // 时间戳（毫秒）
	Node      int64 // 节点 ID
	Step      int64 // 序列号
}

// String 返回 ID 信息的字符串表示
func (info *IDInfo) String() string {
	return fmt.Sprintf("ID: %d, Node: %d, Step: %d, Timestamp: %d",
		info.ID, info.Node, info.Step, info.Timestamp)
}

// 全局便捷函数（使用默认生成器）

// Generate 使用默认生成器生成 ID
func Generate() int64 {
	return GetDefaultGenerator().Generate()
}

// GenerateUint 使用默认生成器生成 uint64 ID
func GenerateUint() uint64 {
	return GetDefaultGenerator().GenerateUint()
}

// GenerateString 使用默认生成器生成字符串 ID
func GenerateString() string {
	return GetDefaultGenerator().GenerateString()
}

// GenerateBatch 使用默认生成器批量生成 ID
func GenerateBatch(count int) []int64 {
	return GetDefaultGenerator().GenerateBatch(count)
}

// ParseID 使用默认生成器解析 ID
func ParseID(id int64) *IDInfo {
	return GetDefaultGenerator().ParseID(id)
}
