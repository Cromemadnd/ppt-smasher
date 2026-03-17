package research

import (
	"context"
	"os"
	"testing"

	"ppt-smasher/internal/config"
	"ppt-smasher/internal/db"

	"github.com/stretchr/testify/assert"
)

func TestResearchWorkflow_Basic(t *testing.T) {
	// 如果没有配置文件或环境变量，则跳过
	if _, err := os.Stat("../../../config.yaml"); os.IsNotExist(err) && os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("Skipping integration test; config.yaml or POSTGRES_HOST environment variable not set")
	}

	// Initialize real config
	config.InitConfig([]string{"../../../"})

	ctx := context.Background()

	// 初始化真实的数据库连接
	db.InitVectorDB(ctx)

	state := TeamResearchState{
		Theme:          "人工神经网络",
		GivenDocuments: []string{"test_intro.pdf"},
	}

	g := BuildResearchTeamGraph()
	compiled, err := g.Compile(ctx)
	assert.NoError(t, err)

	// Invoke the graph
	result, err := compiled.Invoke(ctx, state)

	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Verified state transitions
	assert.NotEmpty(t, result.DocQueries)
	assert.NotEmpty(t, result.ImageQueries)
	assert.NotEmpty(t, result.DataQueries)

	// ParallelTasks should have gathered documents
	assert.GreaterOrEqual(t, len(result.Documents), 1)

	// VDB should be indexed
	assert.True(t, result.VDBStatus)

	// Optional: Check if we can search for something just indexed
	// Note: result.Theme is used as collection name in AddDocumentChunk
	searchRes, err := db.SearchDocument(ctx, result.Theme, "神经网络", 1)
	if err == nil && len(searchRes) > 0 {
		t.Logf("Successfully retrieved indexed content: %s", searchRes[0])
	}
}
