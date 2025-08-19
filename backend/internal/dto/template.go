package dto

type RemoteTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags,omitempty"`
	ComposeURL  string   `json:"composeUrl"`
	EnvURL      string   `json:"envUrl,omitempty"`
	DocsURL     string   `json:"docsUrl,omitempty"`
	IconURL     string   `json:"iconUrl,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

type RemoteRegistry struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Templates   []RemoteTemplate `json:"templates"`
}
