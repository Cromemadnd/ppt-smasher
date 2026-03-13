package agent

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
)

// WsSender 是推送接口，避免与 ws 包产生循环依赖
type WsSender interface {
	SendMsg(msgType string, payload interface{})
}

// WsCallbackHandler 实现 eino 的 Callback 接口，用于拦截工作流状态并推流
type WsCallbackHandler struct {
	Sender WsSender
}

// OnNodeStart 节点开始执行，触发 NODE_ACTIVE 动画
func (h *WsCallbackHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	h.Sender.SendMsg("NODE_ACTIVE", map[string]string{
		"nodeId": info.Name,
	})
	return ctx
}

// OnNodeEnd 节点执行完成，更新 Timeline 弹窗
func (h *WsCallbackHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	h.Sender.SendMsg("STEP_COMPLETED", map[string]string{
		"stepName": info.Name,
		"status":   "success",
	})
	return ctx
}

// OnError 节点内部发生错误
func (h *WsCallbackHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	h.Sender.SendMsg("TASK_ERROR", err.Error())
	return ctx
}

// OnChatModelGenerateBody 大模型流式输出片段，触发 AGENT_THOUGHT_STREAM
func (h *WsCallbackHandler) OnChatModelGenerateBody(ctx context.Context, info *callbacks.RunInfo, chunk *schema.Message) context.Context {
	h.Sender.SendMsg("AGENT_THOUGHT_STREAM", map[string]string{
		"nodeId": info.ComponentID,
		"chunk":  chunk.Content,
		"status": "pending",
	})
	return ctx
}
