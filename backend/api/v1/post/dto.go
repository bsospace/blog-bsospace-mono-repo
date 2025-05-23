package post

type PostContentStructure struct {
	Type    string                 `json:"type"`
	Attrs   map[string]string      `json:"attrs,omitempty"`
	Marks   []map[string]string    `json:"marks,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Content []PostContentStructure `json:"content,omitempty"`
}
