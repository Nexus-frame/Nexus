package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitoo.icu/Nexus/Nexus"
)

func main() {
	// 创建一个带有自定义配置的Nexus引擎
	config := Nexus.DefaultConfig()
	config.LogConfig.Debug = true
	config.LogConfig.AccessLog = true
	config.WebSocketConfig.Port = "8080"
	config.WebSocketConfig.Path = "/ws"

	// 创建引擎
	engine := Nexus.NewWithConfig(config)

	// 添加路由
	engine.GET("/hello", helloHandler)
	engine.POST("/echo", echoHandler)
	engine.GET("/time", timeHandler)

	// 注册用户组路由
	userGroup := engine.Group("/user")
	{
		userGroup.GET("/info/:id", getUserInfo)
		userGroup.POST("/create", createUser)
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

// 处理器函数

func helloHandler(c *Nexus.Context) {
	c.JSON(Nexus.StatusOK, Nexus.N{
		"message": "Hello, World!",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func echoHandler(c *Nexus.Context) {
	// 获取请求体并回显
	c.JSON(Nexus.StatusOK, Nexus.N{
		"echo":      c.Request.Body,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func timeHandler(c *Nexus.Context) {
	c.JSON(Nexus.StatusOK, Nexus.N{
		"now":        time.Now().Format(time.RFC3339),
		"unix":       time.Now().Unix(),
		"unix_nano":  time.Now().UnixNano(),
		"unix_micro": time.Now().UnixMicro(),
		"unix_milli": time.Now().UnixMilli(),
	})
}

func getUserInfo(c *Nexus.Context) {
	// 获取路径参数
	userID := c.Request.Params.ByName("id")

	// 模拟数据库查询
	c.JSON(Nexus.StatusOK, Nexus.N{
		"id":        userID,
		"name":      "User " + userID,
		"email":     "user" + userID + "@example.com",
		"createdAt": time.Now().AddDate(0, 0, -30).Format(time.RFC3339),
	})
}

func createUser(c *Nexus.Context) {
	// 解析请求体
	var userData map[string]interface{}
	var ok bool

	if userData, ok = c.Request.Body.(map[string]interface{}); !ok {
		c.JSON(Nexus.StatusBadRequest, Nexus.N{
			"error": "Invalid request body",
		})
		return
	}

	// 简单验证
	name, nameOk := userData["name"].(string)
	email, emailOk := userData["email"].(string)

	if !nameOk || !emailOk {
		c.JSON(Nexus.StatusBadRequest, Nexus.N{
			"error": "Missing required fields",
		})
		return
	}

	// 模拟用户创建
	c.JSON(Nexus.StatusCreated, Nexus.N{
		"id":        "user_" + Nexus.GenerateUniqueString()[0:8],
		"name":      name,
		"email":     email,
		"createdAt": time.Now().Format(time.RFC3339),
	})
}
