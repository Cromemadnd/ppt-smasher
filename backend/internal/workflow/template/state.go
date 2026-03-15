package template

type Placeholder struct {
	ID    int    `json:"id,omitempty"`
	Type  string `json:"type"`
	Index uint32 `json:"index,omitempty"`
	Name  string `json:"name,omitempty"` // The semantic name predicted by LLM
}

type SlideLayoutSchema struct {
	LayoutName   string        `json:"layout_name"`
	Placeholders []Placeholder `json:"placeholders"`
}

type SlideHTMLSchema struct {
	LayoutName string `json:"layout_name"`
	HTML       string `json:"html"`
}

type TeamTemplateState struct {
	ReferencePPT   string
	HTMLViews      []SlideHTMLSchema // Rendered HTML strings for the slide layouts
	Schemas        []string          // LLM semantic interpretations (the "name" of elements)
	ExtractedStyle []SlideLayoutSchema
}
