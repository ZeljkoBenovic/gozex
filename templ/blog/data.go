package blog

import "time"

type Blog struct {
	Name        string          `yaml:"name"`
	Title       string          `yaml:"title"`
	Bio         string          `yaml:"bio"`
	GitHubName  string          `yaml:"github_name"`
	GitHubURL   string          `yaml:"github_url"`
	LinkedInURL string          `yaml:"linked_in_url"`
	Email       string          `yaml:"email"`
	Roles       []string        `yaml:"roles"`
	Skills      []SkillCategory `yaml:"skills"`
	Projects    []GitHubProject `yaml:"projects"`
	Posts       []Post          `yaml:"-"`
	JobRoles    []JobRole       `yaml:"job_roles"`
}

type SkillCategory struct {
	Category string   `yaml:"category"`
	Items    []string `yaml:"items"`
}

type GitHubProject struct {
	Name        string   `yaml:"name"`
	ProjectURL  string   `yaml:"project_url"`
	Description string   `yaml:"description"`
	Stars       int      `yaml:"stars"`
	Tags        []string `yaml:"tags"`
}

type Post struct {
	Title    string
	Date     time.Time
	Slug     string
	FilePath string
	URL      string
}

type JobRole struct {
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`
}
