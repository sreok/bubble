package main

import (
	"bubble/pkg/aicli"
	"fmt"
	"bufio"
	"os"
)

func main() {
	fmt.Println("=== 会话管理示例 ===")

	// 获取会话管理器
	manager := aicli.GetSessionManager()

	// 创建一个新会话
	sessionID, err := manager.CreateSession(
		"sk-b35a79f8392e8038be4d8c9fb607fa71", // iflow API key
		aicli.WithModel("qwen3-max"),
		aicli.WithBaseURL("https://apis.iflow.cn/v1"),
		aicli.WithEnableContext(),
		aicli.WithMaxContextSize(10),
		aicli.WithInitialRole("你是一个 helpful assistant，用中文回答问题。"),
	)

	if err != nil {
		fmt.Printf("创建会话失败: %v\n", err)
		return
	}

	fmt.Printf("创建会话成功，会话 ID: %s\n", sessionID)
	fmt.Println("输入 'exit' 退出对话\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		// 读取用户输入
		fmt.Print("用户: ")
		prompt, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("读取输入失败: %v\n", err)
			continue
		}

		// 去除换行符
		prompt = prompt[:len(prompt)-1]

		// 检查是否退出
		if prompt == "exit" {
			// 删除会话
			if err := manager.DeleteSession(sessionID); err != nil {
				fmt.Printf("删除会话失败: %v\n", err)
			}
			fmt.Println("对话结束，会话已删除")
			break
		}

		// 发送消息
		fmt.Print("助手: ")
		response, err := manager.SendMessage(sessionID, prompt)
		if err != nil {
			fmt.Printf("发送消息失败: %v\n", err)
			continue
		}

		// 显示回复
		fmt.Println(response)

		// 显示上下文状态
		fmt.Println("\n--- 上下文状态 ---")
		messages, err := manager.ListContextMessages(sessionID)
		if err != nil {
			fmt.Printf("获取上下文消息失败: %v\n", err)
			continue
		}
		fmt.Printf("上下文消息数量: %d\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("[%d] %s: %s\n", i+1, msg.Role, msg.Content)
		}
		fmt.Println("\n")
	}
}
