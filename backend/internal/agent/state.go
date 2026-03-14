package agent

// WorkflowState 工作流状态结构 (Boss 统筹全局状态)
type WorkflowState struct {
	Theme          string   // 主题
	GivenDocuments []string // 提供的参考资料
	ReferencePPT   string   // 参考PPT的模板文件
	
	// 中间产品和结果
	KnowledgeReady bool     // VDB知识是否已准备好
	LayoutSchemas  []string // 从参考PPT中拆解出的版式 Schema 和 HTML view 结构
	Outline        string   // 大纲及每页对应的模板 Schema 映射
	ContentDrafts  []string // 填充好文案的模板结构
	PPTXFiles      []string // 生成文件路径
	ErrorLog       []string
}

// 下位 Agent 自己的编排状态声明
type TeamResearchState struct {
	Theme          string
	GivenDocuments []string
	VDBStatus      bool
}

type TeamTemplateState struct {
	ReferencePPT string
	Schemas      []string
}

type TeamContentState struct {
	Outline            string
	VDBStatus          bool
	AvailableLayouts   []string
	FilledContentDraft []string
}

type TeamRenderState struct {
	ContentDrafts []string
	RenderResults []string
}
