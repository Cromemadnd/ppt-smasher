package agent

// WorkflowState 工作流状态结构
type WorkflowState struct {
	Theme       string
	Researched  []string // 收集的文献等
	Outline     string   // 生成的大纲
	VisualLinks []string // URL
}
