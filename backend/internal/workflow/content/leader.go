package content

import (
	"context"
	"log"
	"ppt-stasher-backend/internal/config"
	"ppt-stasher-backend/internal/workflow/research"
	"strings"

	"github.com/cloudwego/eino/compose"
)

func BuildContentTeamGraph() *compose.Graph[TeamContentState, TeamContentState] {
	g := compose.NewGraph[TeamContentState, TeamContentState]()
	modelID := config.GlobalConfig.LLM.ContentModel
	log.Printf("ContentTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("outline_director", NewOutlineDirectorNode())
	_ = g.AddLambdaNode("content_filler", NewContentFillerNode())
	_ = g.AddLambdaNode("content_critic", NewContentCriticNode())

	researchGraph, err := research.BuildResearchTeamGraph().Compile(context.Background())
	if err != nil {
		log.Fatalf("failed to compile research graph in content leader: %v", err)
	}

	_ = g.AddLambdaNode("call_research", compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Leader] 发现资料不足，调度 Research Team 进行补充调研...")

		// Summarize queries based on who asked for them
		queries := s.ResearchQueries
		if len(queries) == 0 && s.CriticFeedback != "" {
			queries = []string{s.Theme + " " + s.CriticFeedback}
		}

		var docQueries, dataQueries, imgQueries []string
		for _, q := range queries {
			if strings.Contains(q, "图") || strings.Contains(q, "img") {
				imgQueries = append(imgQueries, q)
			} else if strings.Contains(q, "数据") || strings.Contains(q, "data") {
				dataQueries = append(dataQueries, q)
			} else {
				docQueries = append(docQueries, q)
			}
		}

		rs, err := researchGraph.Invoke(ctx, research.TeamResearchState{
			Theme:        s.Theme,
			DocQueries:   docQueries,
			ImageQueries: imgQueries,
			DataQueries:  dataQueries,
		})
		if err != nil {
			log.Printf("[ContentTeam:Leader] Research sub-graph failed: %v", err)
			return s, err
		}

		if rs.VDBStatus {
			s.VDBStatus = true
			log.Printf("[ContentTeam:Leader] 补充调研完成，新资料已入库。")
		}

		// Reset state for re-entry
		s.FillerResultState = ""
		s.CriticDecision = ""
		s.ResearchQueries = nil

		return s, nil
	}))

	_ = g.AddEdge(compose.START, "outline_director")
	_ = g.AddEdge("outline_director", "content_filler")

	_ = g.AddBranch("content_filler", compose.NewGraphBranch(func(_ context.Context, s TeamContentState) (string, error) {
		if s.FillerResultState == "Needs_Research" {
			return "call_research", nil
		}
		return "content_critic", nil
	}, map[string]bool{
		"call_research":  false,
		"content_critic": false,
	}))

	_ = g.AddBranch("content_critic", compose.NewGraphBranch(func(_ context.Context, s TeamContentState) (string, error) {
		switch s.CriticDecision {
		case "Pass":
			return compose.END, nil
		case "Revise_Outline":
			return "outline_director", nil
		case "Revise_Content":
			return "content_filler", nil
		case "Needs_Research":
			return "call_research", nil
		default:
			return compose.END, nil
		}
	}, map[string]bool{
		compose.END:        true,
		"outline_director": false,
		"content_filler":   false,
		"call_research":    false,
	}))

	_ = g.AddEdge("call_research", "content_filler")

	return g
}
