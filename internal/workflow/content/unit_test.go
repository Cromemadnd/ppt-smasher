package content

import (
	"context"
	"fmt"
	"log"
	"os"
	"ppt-smasher/internal/config"
	"ppt-smasher/internal/db"
	"ppt-smasher/internal/llm"
	"testing"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Test_RAG(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化配置 (需确保环境变量或 config.yaml 存在)
	// 如果是本地测试，可能需要手动设置一些基础配置
	configPath := "../../../"
	if _, err := os.Stat(configPath); err == nil {
		config.InitConfig([]string{configPath})
	} else {
		t.Skip("config.yaml not found, skipping RAG test")
		return
	}

	// 2. 初始化 LLM 和 向量数据库
	llm.InitChatModels(ctx)
	db.InitVectorDB(ctx)

	theme := "DeepSeek-V3 技术架构"

	// 3. 模拟数据入库
	testDocs := []string{
		"DeepSeek-V3 采用了 Multi-Head Latent Attention (MLA) 架构，显著降低了推理时的 KV Cache 内存占用。",
		"DeepSeek-V3 引入了 DeepSeek-MoE 架构，通过专家并行和共享专家策略提高了训练和推理效率。",
		"在训练过程中，DeepSeek-V3 使用了 FP8 混合精度训练，大幅提升了计算吞吐量并平衡了精度。",
		"DeepSeek-V3 支持 128K 的上下文窗口，在长文本场景下表现出色。",
	}

	log.Println("正在准备测试数据入库...")
	for i, doc := range testDocs {
		err := db.AddDocumentChunk(ctx, theme, string(rune(i)), doc)
		if err != nil {
			t.Fatalf("Failed to add document chunk: %v", err)
		}
	}

	// 4. 构建测试 Graph (简化版，只包含 RAG 和 LLM 回答逻辑)
	g := compose.NewGraph[TeamContentState, TeamContentState]()

	// RAG 检索节点
	_ = g.AddLambdaNode("rag_retriever", NewRAGNode())

	// LLM 回答节点 (我们直接复用 Filler 的一部分逻辑或模拟一个简单的回答节点)
	_ = g.AddLambdaNode("answer_generator", compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		chatModel := llm.GetContentModel()
		if chatModel == nil {
			return s, fmt.Errorf("content model not initialized")
		}

		prompt := fmt.Sprintf("基于以下背景资料，回答关于 '%s' 的问题。\n\n背景资料:\n%s\n\n请用简洁的语言总结其核心架构特点。", s.Theme, s.VDBContext)

		messages := []*schema.Message{
			schema.UserMessage(prompt),
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return s, err
		}

		s.FilledContentDraft = []string{resp.Content}
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "rag_retriever")
	_ = g.AddEdge("rag_retriever", "answer_generator")
	_ = g.AddEdge("answer_generator", compose.END)

	chain, err := g.Compile(ctx)
	if err != nil {
		t.Fatalf("failed to compile graph: %v", err)
	}

	// 5. 执行测试
	initialState := TeamContentState{
		Theme:     theme,
		VDBStatus: true,
	}

	finalState, err := chain.Invoke(ctx, initialState)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// 6. 验证结果
	if finalState.VDBContext == "" {
		t.Errorf("VDBContext should not be empty")
	}

	log.Printf("检索到的上下文信息:\n%s", finalState.VDBContext)

	if len(finalState.FilledContentDraft) == 0 {
		t.Errorf("FilledContentDraft should contain the LLM answer")
	} else {
		log.Printf("LLM 增强回答结果:\n%s", finalState.FilledContentDraft[0])
	}
}
