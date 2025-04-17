package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitoo.icu/Nexus/Nexus"
)

func main() {
	// 创建一个Nexus引擎
	config := Nexus.DefaultConfig()
	config.LogConfig.Debug = true
	engine := Nexus.NewWithConfig(config)

	// 添加全局中间件
	engine.Use(LoggerMiddleware, RecoveryMiddleware)

	// 创建API路由组
	api := engine.Group("/api")
	api.Use(AuthMiddleware) // 应用于api组的所有路由
	{
		// 公开API（有认证中间件但没有权限中间件）
		api.GET("/public", publicHandler)

		// 管理员API组（有认证和权限中间件）
		admin := api.Group("/admin")
		admin.Use(PermissionMiddleware)
		{
			admin.GET("/stats", statsHandler)
			admin.POST("/config", updateConfigHandler)
		}
	}

	// 创建一个用于接收关闭信号的通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的goroutine中启动服务器
	go func() {
		log.Println("Starting Nexus server on :8080...")
		if err := engine.Run(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待关闭信号
	<-quit
	log.Println("Server is shutting down...")
}

// 中间件

// LoggerMiddleware 是一个记录请求日志的中间件
func LoggerMiddleware(c *Nexus.Context) {
	// 请求开始时间
	start := time.Now()

	// 记录请求信息
	requestID := c.Request.ID
	method := c.Request.Method
	path := c.Request.Path

	// 调用下一个处理器
	c.Next()

	// 请求结束，计算耗时
	duration := time.Since(start)

	// 获取响应状态
	status := c.Response.Status

	// 记录请求日志
	log.Printf("[%s] %s %s %d %v", requestID, method, path, status, duration)
}

// RecoveryMiddleware 是一个恢复中间件，用于捕获和处理panic
func RecoveryMiddleware(c *Nexus.Context) {
	defer func() {
		if err := recover(); err != nil {
			// 记录错误
			log.Printf("[PANIC] %v", err)

			// 返回500错误
			c.JSON(Nexus.StatusInternalServerError, Nexus.N{
				"error":   "Internal Server Error",
				"message": fmt.Sprintf("%v", err),
			})
		}
	}()

	// 调用下一个处理器
	c.Next()
}

// AuthMiddleware 是一个身份验证中间件
func AuthMiddleware(c *Nexus.Context) {
	// 从请求头中获取认证信息
	auth, exists := c.Request.Header["Authorization"]
	if !exists {
		c.JSON(Nexus.StatusUnauthorized, Nexus.N{
			"error":   "Unauthorized",
			"message": "Missing Authorization header",
		})
		c.Exit()
		return
	}

	// 验证token（这里简化处理，只检查token是否存在）
	token, ok := auth.(string)
	if !ok || token == "" {
		c.JSON(Nexus.StatusUnauthorized, Nexus.N{
			"error":   "Unauthorized",
			"message": "Invalid Authorization token",
		})
		c.Exit()
		return
	}

	// 设置用户信息到上下文
	c.Set("user", Nexus.N{
		"id":    "user123",
		"roles": []string{"user"},
	})

	// 继续处理请求
	c.Next()
}

// PermissionMiddleware 是一个权限检查中间件
func PermissionMiddleware(c *Nexus.Context) {
	// 获取用户信息
	user, exists := c.Get("user").(Nexus.N)
	if !exists {
		c.JSON(Nexus.StatusUnauthorized, Nexus.N{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		c.Exit()
		return
	}

	// 检查用户角色
	roles, ok := user["roles"].([]string)
	if !ok || !containsRole(roles, "admin") {
		c.JSON(Nexus.StatusForbidden, Nexus.N{
			"error":   "Forbidden",
			"message": "Insufficient permissions",
		})
		c.Exit()
		return
	}

	// 继续处理请求
	c.Next()
}

// 辅助函数

// containsRole 检查角色切片是否包含特定角色
func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// 处理器函数

// publicHandler 处理公开API请求
func publicHandler(c *Nexus.Context) {
	user := c.Get("user").(Nexus.N)

	c.JSON(Nexus.StatusOK, Nexus.N{
		"message": "This is a public API",
		"user":    user["id"],
		"time":    time.Now().Format(time.RFC3339),
	})
}

// statsHandler 处理统计信息请求
func statsHandler(c *Nexus.Context) {
	// 模拟统计数据
	c.JSON(Nexus.StatusOK, Nexus.N{
		"users":    10000,
		"requests": 500000,
		"cpu":      45.2,
		"memory":   "1.2GB",
		"uptime":   "5d 12h 30m",
	})
}

// updateConfigHandler 处理配置更新请求
func updateConfigHandler(c *Nexus.Context) {
	// 解析请求体
	var configData map[string]interface{}
	var ok bool

	if configData, ok = c.Request.Body.(map[string]interface{}); !ok {
		c.JSON(Nexus.StatusBadRequest, Nexus.N{
			"error": "Invalid request body",
		})
		return
	}

	// 模拟配置更新
	c.JSON(Nexus.StatusOK, Nexus.N{
		"message":   "Configuration updated successfully",
		"config":    configData,
		"updatedAt": time.Now().Format(time.RFC3339),
	})
}
