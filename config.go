package Nexus

import "time"

// Config 表示Nexus引擎的配置选项
type Config struct {
	// WebSocket配置
	WebSocketConfig WebSocketConfig
	// 连接配置
	ConnectionConfig ConnectionConfig
	// 日志配置
	LogConfig LogConfig
}

// WebSocketConfig WebSocket相关配置
type WebSocketConfig struct {
	// 读缓冲区大小
	ReadBufferSize int
	// 写缓冲区大小
	WriteBufferSize int
	// 检查源函数，返回true表示允许连接
	CheckOrigin func(origin string) bool
	// WebSocket路径
	Path string
	// 端口
	Port string
}

// ConnectionConfig 连接相关配置
type ConnectionConfig struct {
	// 发送通道缓冲区大小
	SendChannelSize int
	// 连接超时时间
	ConnectionTimeout time.Duration
	// 心跳间隔
	HeartbeatInterval time.Duration
	// 心跳超时
	HeartbeatTimeout time.Duration
}

// LogConfig 日志相关配置
type LogConfig struct {
	// 是否启用调试日志
	Debug bool
	// 是否记录访问日志
	AccessLog bool
	// 日志格式
	Format string
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		WebSocketConfig: WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(origin string) bool { return true },
			Path:            "/",
			Port:            "8080",
		},
		ConnectionConfig: ConnectionConfig{
			SendChannelSize:   256,
			ConnectionTimeout: 30 * time.Second,
			HeartbeatInterval: 5 * time.Second,
			HeartbeatTimeout:  10 * time.Second,
		},
		LogConfig: LogConfig{
			Debug:     false,
			AccessLog: false,
			Format:    "text",
		},
	}
}
