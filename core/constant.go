package core

import "sync"

// 数据包常量
const (
	Uint8Size     = 1
	Uint16Size    = 2
	Uint32Size    = 4
	Uint64Size    = 8
	MaxPacketSize = 4 * 1 << 20 // 最大数据包大小为最大数据包大小为 4M
)

// 消息类型常量
const (
	TypeAuth                = 0x01 // 验证消息以检查访问密钥是否正确
	typeNoAvailablePort     = 0x02 // 访问密钥没有可用端口
	TypeConnectNatok        = 0xa1 //连接到NATOK服务
	TypeConnectIntra        = 0xa2 //连接到内部服务
	TypeDisconnect          = 0x04 //  断开
	TypeTransfer            = 0x05 //  数据传输
	TypeIsInuseKey          = 0x06 // 访问秘钥已在其他客户端使用
	TypeHeartbeat           = 0x07 // 心跳
	TypeDisabledAccessKey   = 0x08 // 禁用的访问密钥
	TypeDisabledTrialClient = 0x09 // 禁用的试用客户端
	TypeInvalidKey          = 0x10 // 无效的访问密钥
	HeartbeatInterval       = 10   //心跳间隔时长10秒
)

// Counter 计数器
type Counter struct {
	mu    sync.Mutex
	count int
}

func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *Counter) Decrement() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count--
}

func (c *Counter) GetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
