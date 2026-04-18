package blog

import "time"

type Blog struct {
	Name           string          `yaml:"name"`
	Title          string          `yaml:"title"`
	Bio            string          `yaml:"bio"`
	SEODescription string          `yaml:"seo_description"`
	Company        string          `yaml:"company"`
	GitHubName     string          `yaml:"github_name"`
	GitHubURL      string          `yaml:"github_url"`
	LinkedInURL    string          `yaml:"linked_in_url"`
	Email          string          `yaml:"email"`
	SiteURL        string          `yaml:"site_url"`
	Roles          []string        `yaml:"roles"`
	Skills         []SkillCategory `yaml:"skills"`
	Projects       []GitHubProject `yaml:"projects"`
	Posts          []Post          `yaml:"-"`
	JobRoles       []JobRole       `yaml:"job_roles"`
	PostsMeta      []PostMeta      `yaml:"post_meta"`
}

// SEOMeta holds per-page SEO metadata for <head> rendering.
type SEOMeta struct {
	Title        string
	Description  string
	CanonicalURL string
	OGType       string
	JSONLD       string
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
	Kind        string   `yaml:"kind"`
}

type Post struct {
	Title    string
	Date     time.Time
	Slug     string
	FilePath string
	URL      string
	Tag      string
	ReadTime string
}

type JobRole struct {
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`
}

type PostMeta struct {
	Slug     string `yaml:"slug"`
	Tag      string `yaml:"tag"`
	ReadTime string `yaml:"read_time"`
}
