package discord

// Message contains all the information for a webhook message.
type Message struct {
	Content   string  `json:"content"`
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Embeds    []Embed `json:"embeds"`
}

// Embed contains all the information for an embed object.
type Embed struct {
	Color  int          `json:"color,omitempty"`
	Footer EmbedFooter  `json:"footer"`
	Author EmbedAuthor  `json:"author"`
	Fields []EmbedField `json:"fields,omitempty"`
}

// EmbedFooter contains all the information for an embed footer.
type EmbedFooter struct {
	Text string `json:"text"`
}

// EmbedAuthor contains all the information for an embed author.
type EmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

// EmbedField contains all the information for an embed field.
type EmbedField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
