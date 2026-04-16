package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZeljkoBenovic/gozex/templ/blog"
	"github.com/ZeljkoBenovic/gozex/templ/posts"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

func main() {
	data := blog.Blog{
		Posts: make([]blog.Post, 0),
	}

	f, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config.yaml: %v", err)
	}

	if err = yaml.NewDecoder(bytes.NewBuffer(f)).Decode(&data); err != nil {
		log.Fatalf("Error parsing config.yaml: %v", err)
	}

	// Normalise site URL: strip trailing slash for consistent joining.
	siteURL := strings.TrimRight(data.SiteURL, "/")

	ctx := context.Background()
	rootFolder := "public"
	blogFolder := path.Join(rootFolder, "blog")

	// render posts based on MD files in posts folder, file names must be in slug format
	if err := filepath.Walk("./posts/.", func(postFilePath string, info os.FileInfo, err error) error {
		// skip processing dirs
		if info.IsDir() {
			return nil
		}

		// get filename without extension
		postSlug := strings.Split(info.Name(), ".")[0]
		// get post title — prefer the H1 from the file; fall back to slug-derived title
		postTitle := h1FromFile(postFilePath)
		if postTitle == "" {
			postTitle = cases.Title(language.English).String(strings.ReplaceAll(postSlug, "-", " "))
		}
		// get post URL
		postURL := path.Join(blogFolder, info.ModTime().Format("2006/01/02"), postSlug)

		data.Posts = append(data.Posts, blog.Post{
			Title:    postTitle,
			Date:     info.ModTime(),
			Slug:     postSlug,
			FilePath: postFilePath,
			URL:      strings.TrimPrefix(postURL, "public"),
		})

		return nil
	}); err != nil {
		log.Fatalf("could not fetch posts: %v", err)
	}

	//create public folder
	if err := os.Mkdir(rootFolder, 0755); err != nil && !os.IsExist(err) {
		log.Fatalf("failed to create public root dir: %v", err)
	}

	// create public/blog folder
	if err := os.Mkdir(blogFolder, 0755); err != nil && !os.IsExist(err) {
		log.Fatalf("failed to create blog root dir: %v", err)
	}

	// create public/index.html
	index, err := os.Create(path.Join(rootFolder, "index.html"))
	if err != nil {
		log.Fatalf("could not create public/index.html: %v", err)
	}

	homeDesc := data.SEODescription
	if homeDesc == "" {
		homeDesc = truncate(data.Bio, 160)
	}
	homeSEO := blog.SEOMeta{
		Title:        data.Name + " – Kubernetes Platform Engineer | SRE",
		Description:  homeDesc,
		CanonicalURL: siteURL + "/",
		OGType:       "website",
		JSONLD:       personJSONLD(data, siteURL),
	}

	// render public/index.html
	if err = blog.Index(homeSEO, blog.Body(data)).Render(ctx, index); err != nil {
		log.Fatalf("could not render public/index.html: %v", err)
	}

	// create public/blog/index.html
	blogIndex, err := os.Create(path.Join(blogFolder, "index.html"))
	if err != nil {
		log.Fatalf("failed to create public/blog/index.html file: %v", err)
	}

	blogSEO := blog.SEOMeta{
		Title:        "Kubernetes & Platform Engineering Blog – " + data.Name,
		Description:  "Engineering blog covering Kubernetes operations, SRE practices, platform automation, and DevSecOps — practical insights from production at Consensys Linea and Polygon Labs.",
		CanonicalURL: siteURL + "/blog/",
		OGType:       "website",
	}

	// render public/blog/index.html
	err = blog.Index(blogSEO, posts.Body(data)).Render(ctx, blogIndex)
	if err != nil {
		log.Fatalf("failed to write index page: %v", err)
	}

	for _, post := range data.Posts {
		// create dedicated folder for each post
		dir := path.Join(blogFolder, post.Date.Format("2006/01/02"), post.Slug)
		if err = os.MkdirAll(dir, 0755); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("failed to create dir %q: %v", dir, err)
		}

		// create index.html for each post
		name := path.Join(dir, "index.html")
		outFile, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}

		// read post content from posts MD file
		postContent, err := os.ReadFile(post.FilePath)
		if err != nil {
			log.Fatalf("failed to read post content: %v", err)
		}

		gm := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				parser.WithAttribute(),
			),
		)
		// and convert it to HTML
		var buf bytes.Buffer
		if err := gm.Convert(postContent, &buf); err != nil {
			log.Fatalf("failed to convert markdown to HTML: %v", err)
		}

		postURL := siteURL + post.URL + "/"
		postDesc := extractDescription(string(postContent))
		postSEO := blog.SEOMeta{
			Title:        post.Title + " – " + data.Name,
			Description:  postDesc,
			CanonicalURL: postURL,
			OGType:       "article",
			JSONLD:       blogPostingJSONLD(post, data, postURL, postDesc, siteURL),
		}

		// Create an unsafe component containing raw HTML.
		content := posts.Unsafe(buf.String())

		// Use templ to render the template containing the raw HTML.
		if err = blog.Index(postSEO, posts.Post(content, post.Title, post.Date)).Render(ctx, outFile); err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
	}

	// Generate robots.txt
	generateRobotsTxt(rootFolder, siteURL)

	// Generate sitemap.xml (only when a site URL is configured)
	if siteURL != "" {
		generateSitemap(rootFolder, siteURL, data.Posts)
	}
}

// h1FromFile reads a Markdown file and returns the text of the first H1 heading,
// or an empty string if none is found.
func h1FromFile(filePath string) string {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// truncate shortens s to at most max characters, appending "…" if trimmed.
func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// extractDescription returns the first meaningful non-heading paragraph from
// Markdown source, stripped of inline markers, capped at 160 characters.
func extractDescription(md string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") {
			continue
		}
		// Strip common inline Markdown markers.
		for _, marker := range []string{"**", "*", "`", "_"} {
			line = strings.ReplaceAll(line, marker, "")
		}
		return truncate(line, 160)
	}
	return ""
}

// personJSONLD returns a Schema.org Person JSON-LD string for the homepage.
// It includes knowsAbout (derived from all skill items) and worksFor (current company).
func personJSONLD(data blog.Blog, siteURL string) string {
	obj := map[string]any{
		"@context": "https://schema.org",
		"@type":    "Person",
		"name":     data.Name,
		"jobTitle": data.Title,
		"sameAs":   []string{data.GitHubURL, data.LinkedInURL},
	}
	if siteURL != "" {
		obj["url"] = siteURL + "/"
	}
	// Use the dedicated SEO description if available, otherwise fall back to bio.
	desc := data.SEODescription
	if desc == "" {
		desc = data.Bio
	}
	if desc != "" {
		obj["description"] = desc
	}
	// Flatten all skill items into knowsAbout for rich structured signal.
	var skills []string
	for _, cat := range data.Skills {
		skills = append(skills, cat.Items...)
	}
	if len(skills) > 0 {
		obj["knowsAbout"] = skills
	}
	// Current employer, if configured.
	if data.Company != "" {
		obj["worksFor"] = map[string]string{
			"@type": "Organization",
			"name":  data.Company,
		}
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(b)
}

// blogPostingJSONLD returns a Schema.org @graph JSON-LD string for a post page.
// It combines BlogPosting and BreadcrumbList in a single script block.
func blogPostingJSONLD(p blog.Post, data blog.Blog, postURL, description string, siteURL string) string {
	author := map[string]any{
		"@type": "Person",
		"name":  data.Name,
	}
	if siteURL != "" {
		author["url"] = siteURL + "/"
	}

	blogPosting := map[string]any{
		"@type":         "BlogPosting",
		"headline":      p.Title,
		"datePublished": p.Date.Format(time.RFC3339),
		"dateModified":  p.Date.Format(time.RFC3339),
		"author":        author,
		"publisher": map[string]any{
			"@type": "Person",
			"name":  data.Name,
		},
		// Surface the author's primary expertise as post keywords so crawlers
		// associate the content with the Kubernetes/SRE topic cluster.
		"keywords": "Kubernetes, SRE, Platform Engineering, DevSecOps, Go, DevOps",
	}
	if postURL != "" {
		blogPosting["url"] = postURL
		blogPosting["mainEntityOfPage"] = map[string]string{
			"@type": "WebPage",
			"@id":   postURL,
		}
	}
	if description != "" {
		blogPosting["description"] = description
	}

	breadcrumb := map[string]any{
		"@type": "BreadcrumbList",
		"itemListElement": []any{
			map[string]any{"@type": "ListItem", "position": 1, "name": "Home", "item": siteURL + "/"},
			map[string]any{"@type": "ListItem", "position": 2, "name": "Blog", "item": siteURL + "/blog/"},
			map[string]any{"@type": "ListItem", "position": 3, "name": p.Title},
		},
	}

	graph := map[string]any{
		"@context": "https://schema.org",
		"@graph":   []any{blogPosting, breadcrumb},
	}
	b, err := json.Marshal(graph)
	if err != nil {
		return ""
	}
	return string(b)
}

// generateRobotsTxt writes public/robots.txt.
func generateRobotsTxt(rootFolder, siteURL string) {
	var sb strings.Builder
	sb.WriteString("User-agent: *\n")
	sb.WriteString("Allow: /\n")
	if siteURL != "" {
		fmt.Fprintf(&sb, "Sitemap: %s/sitemap.xml\n", siteURL)
	}
	if err := os.WriteFile(path.Join(rootFolder, "robots.txt"), []byte(sb.String()), 0644); err != nil {
		log.Fatalf("failed to write robots.txt: %v", err)
	}
}

// generateSitemap writes public/sitemap.xml.
func generateSitemap(rootFolder, siteURL string, blogPosts []blog.Post) {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	writeURL := func(loc, lastmod, changefreq, priority string) {
		sb.WriteString("  <url>\n")
		fmt.Fprintf(&sb, "    <loc>%s</loc>\n", loc)
		if lastmod != "" {
			fmt.Fprintf(&sb, "    <lastmod>%s</lastmod>\n", lastmod)
		}
		fmt.Fprintf(&sb, "    <changefreq>%s</changefreq>\n", changefreq)
		fmt.Fprintf(&sb, "    <priority>%s</priority>\n", priority)
		sb.WriteString("  </url>\n")
	}

	writeURL(siteURL+"/", "", "monthly", "1.0")
	writeURL(siteURL+"/blog/", "", "weekly", "0.8")

	for _, p := range blogPosts {
		postURL := siteURL + p.URL + "/"
		writeURL(postURL, p.Date.Format("2006-01-02"), "yearly", "0.6")
	}

	sb.WriteString("</urlset>\n")

	if err := os.WriteFile(path.Join(rootFolder, "sitemap.xml"), []byte(sb.String()), 0644); err != nil {
		log.Fatalf("failed to write sitemap.xml: %v", err)
	}
}
