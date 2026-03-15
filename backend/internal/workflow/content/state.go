package content

type TeamContentState struct {
	Theme              string
	Outline            string
	VDBStatus          bool
	AvailableLayouts   []string
	FilledContentDraft []string

	// For loop and feedback
	FillerResultState string   // "Success" or "Needs_Research"
	ResearchQueries   []string // From Filler when needing more context
	CriticFeedback    string   // Raw feedback from Critic
	CriticDecision    string   // "Pass", "Revise_Outline", "Revise_Content", "Needs_Research"
	CurrentFeedback   string   // Synthesized feedback from Leader
	VDBContext        string   // Context fetched from VDB
}
