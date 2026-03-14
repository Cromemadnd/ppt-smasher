package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
)

var globalWorkflowRunner compose.Runnable[WorkflowState, WorkflowState]

// InitWorkflow 随 App 启动直接初始化 Agent 编排
func InitWorkflow() error {
	runner, err := BuildBossGraph()
	if err != nil {
		return fmt.Errorf("failed to build boss graph: %w", err)
	}
	globalWorkflowRunner = runner
	log.Println("Agent workflow initialized successfully")
	return nil
}

// RunWorkflow 触发大模型编排逻辑
func RunWorkflow(sender WsSender, theme string) {
	if globalWorkflowRunner == nil {
		sender.SendMsg("TASK_ERROR", "Workflow runner not initialized")
		return
	}

	// 1. 初始化状态
	initialState := WorkflowState{Theme: theme}
	wsCallback := &WsCallbackHandler{Sender: sender}

	// 2. 将 callback 写入 context 开启全局监控推流
	runInfo := &callbacks.RunInfo{Name: "workflow_root"}
	ctx := callbacks.InitCallbacks(context.Background(), runInfo, wsCallback)

	// 3. 执行工作流
	result, err := globalWorkflowRunner.Invoke(ctx, initialState)
	if err != nil {
		sender.SendMsg("TASK_ERROR", err.Error())
		return
	}

	// 4. 发送最终成功消息
	sender.SendMsg("TASK_SUCCESS", map[string]interface{}{
		"result": result.PPTXFiles,
	})
}
