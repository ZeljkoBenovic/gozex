package blog

import "time"

type Blog struct {
	Name        string           `yaml:"name"`
	GitHubName  string           `yaml:"github_name"`
	GitHubURL   string           `yaml:"github_url"`
	LinkedInURL string           `yaml:"linked_in_url"`
	Email       string           `yaml:"email"`
	Roles       []string         `yaml:"roles"`
	Projects    []GitHubProjects `yaml:"projects"`
	Posts       []Post           `yaml:"-"`
}

type GitHubProjects struct {
	Name        string `yaml:"name"`
	ProjectURL  string `yaml:"project_url"`
	Description string `yaml:"description"`
}

type Post struct {
	Title    string
	Date     time.Time
	Slug     string
	FilePath string
	URL      string
}
