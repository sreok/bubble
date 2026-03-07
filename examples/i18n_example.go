package main

import (
	"bubble/internal/config"
	"bubble/internal/i18n"
	"log"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Printf("Failed to load config: %v\n", err)
		return
	}

	// 设置语言（从配置文件读取）
	i18n.SetLanguage(i18n.Language(config.AppConfig.Language))

	// 测试国际化输出
	log.Println(i18n.T("Response"))
	log.Println(i18n.T("Error"))
	log.Println(i18n.TF("Config file %s has been modified", config.GetConfigPath()))

	// 切换语言为中文
	log.Println("\nSwitching to Chinese...")
	i18n.SetLanguage(i18n.Chinese)

	// 测试中文输出
	log.Println(i18n.T("Response"))
	log.Println(i18n.T("Error"))
	log.Println(i18n.TF("Config file %s has been modified", config.GetConfigPath()))

	// 切换语言为英文
	log.Println("\nSwitching to English...")
	i18n.SetLanguage(i18n.English)

	// 测试英文输出
	log.Println(i18n.T("Response"))
	log.Println(i18n.T("Error"))
	log.Println(i18n.TF("Config file %s has been modified", config.GetConfigPath()))
}
