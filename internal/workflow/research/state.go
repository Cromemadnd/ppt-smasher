package research

type TeamResearchState struct {
	Theme          string
	GivenDocuments []string
	VDBStatus      bool
	Documents      []string // 检索到的文档内容
	Images         []string // 检索到的图片链接
	Analytics      []string // 提取的统计数据/图表定义

	// LLM 生成的搜索关键词
	DocQueries   []string
	ImageQueries []string
	DataQueries  []string
}
