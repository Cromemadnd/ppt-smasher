package agent

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
)

// 状态定义
type WorkflowState struct {
	Theme       string
	Researched  []string
	Outline     string
	VisualLinks []string
}

// RunWorkflow 触发大模型编排逻辑
func RunWorkflow(sender WsSender, theme string) {
	// 1. 初始化 Graph 构建器
	g := compose.NewGraph[WorkflowState, WorkflowState]()

	// 2. 注册节点
	_ = g.AddLambdaNode("boss_node", compose.InvokableLambda(bossAction))
	_ = g.AddLambdaNode("researcher_node", compose.InvokableLambda(researcherAction))
	_ = g.AddLambdaNode("content_node", compose.InvokableLambda(contentAction))
	_ = g.AddLambdaNode("visual_node", compose.InvokableLambda(visualAction))

	// 3. 编排边
	_ = g.AddEdge(compose.START, "boss_node")
	_ = g.AddEdge("boss_node", "researcher_node")
	_ = g.AddEdge("researcher_node", "content_node")
	_ = g.AddEdge("content_node", "visual_node")
	_ = g.AddEdge("visual_node", compose.END)

	// 4. 编译 Graph
	runner, err := g.Compile(context.Background())
	if err != nil {
		sender.SendMsg("TASK_ERROR", err.Error())
		return
	}

	// 5. 注入 Callback
	initialState := WorkflowState{Theme: theme}
	wsCallback := &WsCallbackHandler{Sender: sender}

	// 初始化 callback context
	ctx := callbacks.InitCallbacks(context.Background(), wsCallback)

	// 6. 运行
	result, err := runner.Invoke(ctx, initialState)
	if err != nil {
		sender.SendMsg("TASK_ERROR", err.Error())
		return
	}

	// 7. 发送最终成功消息
	sender.SendMsg("TASK_SUCCESS", map[string]interface{}{
		"result": result.VisualLinks,
	})
}

// nodes 模拟

func bossAction(ctx context.Context, state WorkflowState) (WorkflowState, error) {
	return state, nil
}
func researcherAction(ctx context.Context, state WorkflowState) (WorkflowState, error) {
	state.Researched = append(state.Researched, "发现一些文献...")
	return state, nil
}
func contentAction(ctx context.Context, state WorkflowState) (WorkflowState, error) {
	state.Outline = "第一部分：前言..."
	return state, nil
}
func visualAction(ctx context.Context, state WorkflowState) (WorkflowState, error) {
	state.VisualLinks = append(state.VisualLinks, "https://mock.com/slide1.png")
	return state, nil
}
