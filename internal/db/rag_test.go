package db

import (
	"context"
	"fmt"
	"os"
	"ppt-smasher/internal/config"
	"testing"
)

// 实际集成测试（需要本地或 CI 环境有 Postgres 运行并配置 .env 或环境变量）
func TestIntegration_VectorDB(t *testing.T) {
	// 如果没有配置文件或环境变量，则跳过
	if _, err := os.Stat("../../config.yaml"); os.IsNotExist(err) && os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("Skipping integration test; config.yaml or POSTGRES_HOST environment variable not set")
	}

	config.InitConfig([]string{"../../"})
	ctx := context.Background()

	// 初始化真实的数据库连接
	InitVectorDB(ctx)

	if vectorDB == nil {
		t.Fatal("Failed to initialize vectorDB")
	}

	theme := "integration_test_theme"
	docID := "test_doc_1"
	content := "这是一段用于集成测试的真实向量索引文本"

	// 1. 测试索引
	err := AddDocumentChunk(ctx, theme, docID, content)
	if err != nil {
		t.Fatalf("AddDocumentChunk failed: %v. Please check if Postgres is running and configuration is correct.", err)
	}

	// 2. 测试检索
	results, err := SearchDocument(ctx, theme, "集成测试", 5)
	if err != nil {
		t.Fatalf("SearchDocument failed: %v", err)
	}

	found := false
	for _, res := range results {
		if res == content {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Content not found in search results. Results: %v", results)
	} else {
		t.Log("Integration test passed: successfully indexed and retrieved document from real DB.")
	}

	// 3. 真实环境下的召回率测试
	t.Run("RealRecallRate", func(t *testing.T) {
		testData := []struct {
			theme   string
			content string
			query   string
		}{
			{"RAG_Test", "生成式人工智能正在改变世界", "什么是AI？"},
			{"RAG_Test", "大语言模型是深度学习的分支", "LLM是什么"},
			{"RAG_Test", "红烧肉是一道经典的中国菜", "怎么做红烧肉"},
			{"RAG_Test", "足球是世界上最受欢迎的运动", "体育项目"},
			{"RAG_Test", "Go语言以并发编程见长", "Golang特点"},
			{"RAG_Test", "中国正在加大对数字经济的投入", "国家政策对数字经济的支持"},
			{"RAG_Test", "唐朝是中国最强盛的朝代之一", "唐代的国际影响力"},
			{"RAG_Test", "青霉素的发现是医学史上伟大的时刻", "抗生素的起源"},
			{"RAG_Test", "火星探索是人类迈向多行星社会的第一步", "星际殖民计划"},
		}

		// 批量插入测试数据
		for i, td := range testData {
			err := AddDocumentChunk(ctx, td.theme, fmt.Sprintf("recall_%d", i), td.content)
			if err != nil {
				t.Fatalf("Failed to index test data [%d]: %v", i, err)
			}
		}

		// 执行检索并计算召回率
		hits := 0
		for _, td := range testData {
			results, err := SearchDocument(ctx, td.theme, td.query, 5)
			if err != nil {
				t.Logf("Search failed for query [%s]: %v", td.query, err)
				continue
			}

			found := false
			for _, res := range results {
				if res == td.content {
					found = true
					break
				}
			}
			if found {
				hits++
			}
		}

		recall := float64(hits) / float64(len(testData))
		t.Logf("Real DB Recall Evaluation - Total: %d, Hits: %d, Recall Rate: %.2f%%", len(testData), hits, recall*100)

		if recall < 0.6 {
			t.Errorf("Real DB recall rate %.2f is too low (expected >= 0.6)", recall)
		}
	})
}
