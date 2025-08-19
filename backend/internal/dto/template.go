package dto

type RemoteTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	ComposeURL  string   `json:"composeUrl"`
	EnvURL      string   `json:"envUrl"`
	DocsURL     string   `json:"docsUrl"`
	IconURL     string   `json:"iconUrl"`
	UpdatedAt   string   `json:"updatedAt"`
}

type RemoteRegistry struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Templates   []RemoteTemplate `json:"templates"`
}
