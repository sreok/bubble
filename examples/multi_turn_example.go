package main

import (
	"bubble/internal/config"
	"bubble/pkg/aicli"
	"bufio"
	"context"
	"fmt"
	"os"
)

func main() {
	// 加载配置文件
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	provider, model, err := aicli.SelectAvailableProviderAndModel()
	if err != nil {
		fmt.Printf("选择供应商和模型失败: %v\n", err)
		return
	}

	client, err := aicli.NewClient(
		provider.APIKey,
		aicli.WithModel(model.ID),
		aicli.WithBaseURL(provider.BaseURL),
		aicli.WithEnableContext(),
		aicli.WithMaxContextSize(10),
		aicli.WithInitialRole("你是一个 helpful assistant，用中文回答问题。"),
	)

	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}

	fmt.Println("=== 多轮对话示例 ===")
	fmt.Printf("使用提供商: %s\n", provider.Name)
	fmt.Printf("使用模型: %s\n", model.ID)
	fmt.Println("输入 'exit' 退出对话")

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
			fmt.Println("对话结束")
			break
		}

		// 发送请求
		fmt.Print("助手: ")
		stream := client.SendStream(context.Background(), prompt)
		for chunk := range stream.Content {
			fmt.Print(chunk)
		}
		fmt.Println()
		if stream.Err != nil {
			fmt.Printf("流式请求失败: %v\n", stream.Err)
			continue
		}

		// 显示上下文状态
		fmt.Println("\n--- 上下文状态 ---")
		messages := client.ListContextMessages()
		fmt.Printf("上下文消息数量: %d\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("[%d] %s: %s\n", i+1, msg.Role, msg.Content)
		}
		fmt.Println()
	}
}
