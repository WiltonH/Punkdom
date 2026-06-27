package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"punkdom/config"
)

// handleSettingsGet GET /api/settings — 返回三层配置快照。
func (h *Handlers) HandleSettingsGet(ctx context.Context, c *app.RequestContext) {
	layered, err := h.app.Settings()
	if err != nil {
		writeError(c, consts.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, layered)
}

// handleSettingsUserUpdate PUT /api/settings/user — 持久化用户级配置。
func (h *Handlers) HandleSettingsUserUpdate(ctx context.Context, c *app.RequestContext) {
	var body config.Settings
	if err := c.BindJSON(&body); err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequestWithDetail", "detail", err.Error())
		return
	}
	layered, err := h.app.UpdateUserSettings(body)
	if err != nil {
		writeError(c, consts.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, layered)
}

// handleSettingsWorkspaceUpdate PUT /api/settings/workspace — 持久化工作区级配置。
func (h *Handlers) HandleSettingsWorkspaceUpdate(ctx context.Context, c *app.RequestContext) {
	var body config.Settings
	if err := c.BindJSON(&body); err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequestWithDetail", "detail", err.Error())
		return
	}
	layered, err := h.app.UpdateWorkspaceSettings(body)
	if err != nil {
		if err.Error() == "当前没有打开的工作区" {
			writeErrorKey(c, consts.StatusBadRequest, "api.settings.workspaceMissing")
			return
		}
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, layered)
}

// TestModelRequest is the payload for model testing.
type TestModelRequest struct {
	OpenAIAPIKey  string `json:"openai_api_key"`
	OpenAIBaseURL string `json:"openai_base_url"`
	OpenAIModel   string `json:"openai_model"`
}

// HandleSettingsTestModel POST /api/settings/test-model — 测试模型连接性。
func (h *Handlers) HandleSettingsTestModel(ctx context.Context, c *app.RequestContext) {
	var body TestModelRequest
	if err := c.BindJSON(&body); err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequestWithDetail", "detail", err.Error())
		return
	}

	testCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	maxTokens := 5
	modelCfg := openai.ChatModelConfig{
		APIKey:    body.OpenAIAPIKey,
		BaseURL:   body.OpenAIBaseURL,
		Model:     body.OpenAIModel,
		MaxTokens: &maxTokens,
	}

	cm, err := openai.NewChatModel(testCtx, &modelCfg)
	if err != nil {
		writeJSON(c, consts.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("创建模型实例失败: %v", err),
		})
		return
	}

	// 发送简单的 ping 进行连通性测试
	resp, err := cm.Generate(testCtx, []*schema.Message{
		schema.UserMessage("ping"),
	})
	if err != nil {
		writeJSON(c, consts.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("模型调用失败: %v", err),
		})
		return
	}

	if resp == nil || resp.Content == "" {
		writeJSON(c, consts.StatusOK, map[string]any{
			"success": false,
			"error":   "模型返回内容为空",
		})
		return
	}

	writeJSON(c, consts.StatusOK, map[string]any{
		"success": true,
		"message": resp.Content,
	})
}
