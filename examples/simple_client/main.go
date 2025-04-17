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
	// 创建客户端配置
	config := Nexus.DefaultClientConfig()
	config.Debug = true
	config.RequestTimeout = 5 * time.Second
	config.AutoReconnect = true

	// 连接到服务器
	client, err := Nexus.NewClientWithConfig("ws", "localhost:8080", "/ws", config)
	if err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer client.Close()

	// 创建一个用于接收关闭信号的通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的goroutine中执行请求
	go func() {
		// 等待2秒，确保服务器已启动
		time.Sleep(2 * time.Second)

		// 发送Hello请求
		sendHelloRequest(client)

		// 发送Echo请求
		sendEchoRequest(client)

		// 发送Time请求
		sendTimeRequest(client)

		// 发送用户信息请求
		sendGetUserInfoRequest(client, "123")

		// 发送创建用户请求
		sendCreateUserRequest(client, "张三", "zhangsan@example.com")
	}()

	// 等待关闭信号
	<-quit
	log.Println("客户端关闭...")
}

// 发送Hello请求
func sendHelloRequest(client *Nexus.Client) {
	req := client.Req(Nexus.GET, "/hello", nil)

	log.Println("发送Hello请求...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Hello请求失败: %v", err)
		return
	}

	printResponse("Hello响应", resp)
}

// 发送Echo请求
func sendEchoRequest(client *Nexus.Client) {
	req := client.Req(Nexus.POST, "/echo", Nexus.N{
		"message": "这是一条测试消息",
		"time":    time.Now().Format(time.RFC3339),
	})

	log.Println("发送Echo请求...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Echo请求失败: %v", err)
		return
	}

	printResponse("Echo响应", resp)
}

// 发送Time请求
func sendTimeRequest(client *Nexus.Client) {
	req := client.Req(Nexus.GET, "/time", nil)

	log.Println("发送Time请求...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Time请求失败: %v", err)
		return
	}

	printResponse("Time响应", resp)
}

// 发送获取用户信息请求
func sendGetUserInfoRequest(client *Nexus.Client, userID string) {
	req := client.Req(Nexus.GET, "/user/info/"+userID, nil)

	log.Printf("发送获取用户[%s]信息请求...", userID)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("获取用户信息请求失败: %v", err)
		return
	}

	printResponse("用户信息响应", resp)
}

// 发送创建用户请求
func sendCreateUserRequest(client *Nexus.Client, name, email string) {
	req := client.Req(Nexus.POST, "/user/create", Nexus.N{
		"name":  name,
		"email": email,
	})

	log.Printf("发送创建用户[%s]请求...", name)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("创建用户请求失败: %v", err)
		return
	}

	printResponse("创建用户响应", resp)
}

// 打印响应
func printResponse(title string, resp *Nexus.ResMessage) {
	fmt.Printf("\n--- %s ---\n", title)
	fmt.Printf("状态码: %d\n", resp.Status)
	fmt.Printf("ID: %s\n", resp.ID)
	fmt.Printf("时间戳: %s\n", resp.Timestamp.Format(time.RFC3339))
	fmt.Printf("响应体: %+v\n\n", resp.Body)
}
