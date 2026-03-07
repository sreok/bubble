package api

import (
	"bubble/internal/config"
	"bubble/internal/i18n"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Model 模型的结构体
// type Model struct {
// 	BaseURL string   `json:"base_url"`
// 	ID      string   `json:"id"`
// 	Input   []string `json:"input"`
// 	Output  []string `json:"output"`
// }

// GetModels 获取所有模型的接口
// @Summary 获取所有模型
// @Description 返回所有配置的模型信息
// @Tags models
// @Produce json
// @Success 200 {object} map[string]interface{}{"models": []config.ModelConfig}
// @Router /models [get]
func GetModels(c *gin.Context) {
	// 检查配置文件是否变化并重新加载
	err, _ := config.CheckAndReload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T("config_modified_reload"),
		})
		return
	}

	var models []config.ModelConfig
	for _, provider := range config.AppConfig.Models.Providers {
		for _, m := range provider.Models {
			models = append(models, config.ModelConfig{
				ID:     m.ID,
				Input:  m.Input,
				Output: m.Output,
			})
		}
	}

	// 返回模型列表
	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}
