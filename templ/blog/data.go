package blog

import "time"

type Blog struct {
	Name        string
	GitHubName  string
	GitHubURL   string
	LinkedInURL string
	Email       string
	Roles       []string
	Projects    []GitHubProjects
	Posts       []Post
}

type GitHubProjects struct {
	Name        string
	ProjectURL  string
	Description string
}

type Post struct {
	Title    string
	Date     time.Time
	Slug     string
	FilePath string
	URL      string
}
