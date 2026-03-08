package web_search

import (
	"bubble/internal/config"
	"context"
	"fmt"
	"strings"

	"github.com/jzf21/tavily-client/tavily"
)

// TavilyClient Tavily 客户端
type TavilyClient struct {
	client *tavily.Client
}

// NewTavilyClient 创建一个新的 Tavily 客户端
func NewTavilyClient(apiKey string) *TavilyClient {
	client := tavily.NewClient(apiKey)
	return &TavilyClient{
		client: client,
	}
}

// Search 使用 Tavily 搜索
func (t *TavilyClient) Search(ctx context.Context, query string) (string, error) {
	result, err := t.client.Search(query)
	if err != nil {
		return "", err
	}

	// 格式化搜索结果
	var formattedResult string
	formattedResult = "搜索结果：\n"
	for i, item := range result.Results {
		formattedResult += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		formattedResult += fmt.Sprintf("   链接: %s\n", item.URL)
		formattedResult += fmt.Sprintf("   摘要: %s\n\n", item.Content)
	}

	return formattedResult, nil
}

// IsTavilyEnabled 检查 Tavily 是否启用
func IsTavilyEnabled() bool {
	return config.AppConfig.Tools.WebSearch.Tavily.Enable
}

// ContainsSearchKeywords 检查 prompt 是否包含搜索关键词
func ContainsSearchKeywords(prompt string) bool {
	// 简单的关键词检查
	keywords := []string{"什么", "如何", "为什么", "怎样", "是否", "哪里", "何时", "最新", "现在", "当前", "今天", "最近", "新闻", "信息", "查询", "搜索", "了解", "天气", "股票", "价格", "时间", "日期", "事件", "比赛", "结果", "更新", "版本", "发布", "推出", "上市", "政策", "法规", "规定", "法律", "条款", "条件", "优惠", "活动", "促销", "折扣", "价格", "费用", "成本", "收费", "报价", "价格表", "菜单", "目录", "列表", "清单", "排名", "排行榜", "推荐", "建议", "意见", "评价", "评论", "口碑", "体验", "感受", "使用", "操作", "步骤", "方法", "教程", "指南", "手册", "说明", "文档", "资料", "信息", "数据", "统计", "报告", "分析", "研究", "调查", "结果", "结论", "发现", "趋势", "预测", "未来", "趋势", "发展", "变化", "更新", "升级", "改进", "优化", "修复", "解决", "处理", "应对", "预防", "避免", "防止", "减少", "增加", "提高", "降低", "改善", "恶化", "影响", "作用", "效果", "后果", "结果", "结局", "最终", "结论", "总结", "概括", "归纳", "整理", "组织", "安排", "计划", "规划", "设计", "方案", "策略", "方法", "技巧", "技能", "能力", "知识", "经验", "教训", "启示", "启发", "灵感", "创意", "创新", "创造", "发明", "发现", "探索", "研究", "开发", "设计", "制作", "生产", "制造", "加工", "处理", "操作", "使用", "应用", "实践", "实施", "执行", "落实", "完成", "实现", "达成", "达到", "实现", "完成", "结束", "终止", "停止", "暂停", "继续", "持续", "保持", "维持", "维护", "保护", "保存", "存储", "备份", "恢复", "还原", "重置", "重启", "启动", "开始", "初始化", "设置", "配置", "调整", "修改", "编辑", "更新", "升级", "降级", "安装", "卸载", "删除", "添加", "插入", "导入", "导出", "上传", "下载", "复制", "粘贴", "剪切", "移动", "重命名", "删除", "清空", "清理", "整理", "组织", "分类", "归档", "备份", "恢复", "还原", "重置", "重启", "启动", "开始", "初始化", "设置", "配置", "调整", "修改", "编辑", "更新", "升级", "降级", "安装", "卸载", "删除", "添加", "插入", "导入", "导出", "上传", "下载", "复制", "粘贴", "剪切", "移动", "重命名", "删除", "清空", "清理", "整理", "组织", "分类", "归档"}
	for _, keyword := range keywords {
		if strings.Contains(prompt, keyword) {
			return true
		}
	}
	return false
}
