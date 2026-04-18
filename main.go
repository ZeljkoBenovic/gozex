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

	// Build a slug→meta lookup for tags and read times.
	metaMap := make(map[string]blog.PostMeta)
	for _, m := range data.PostsMeta {
		metaMap[m.Slug] = m
	}

	// Normalise site URL: strip trailing slash for consistent joining.
	siteURL := strings.TrimRight(data.SiteURL, "/")

	ctx := context.Background()
	rootFolder := "public"
	blogFolder := path.Join(rootFolder, "blog")

	// Render posts based on MD files in posts folder; file names must be in slug format.
	if err := filepath.Walk("./posts/.", func(postFilePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		postSlug := strings.Split(info.Name(), ".")[0]
		postTitle := h1FromFile(postFilePath)
		if postTitle == "" {
			postTitle = cases.Title(language.English).String(strings.ReplaceAll(postSlug, "-", " "))
		}
		postURL := path.Join(blogFolder, info.ModTime().Format("2006/01/02"), postSlug)

		p := blog.Post{
			Title:    postTitle,
			Date:     info.ModTime(),
			Slug:     postSlug,
			FilePath: postFilePath,
			URL:      strings.TrimPrefix(postURL, "public") + "/",
		}

		if m, ok := metaMap[postSlug]; ok {
			p.Tag = m.Tag
			p.ReadTime = m.ReadTime
		}

		data.Posts = append(data.Posts, p)
		return nil
	}); err != nil {
		log.Fatalf("could not fetch posts: %v", err)
	}

	// Create output directories.
	for _, dir := range []string{
		rootFolder,
		blogFolder,
		path.Join(rootFolder, "projects"),
		path.Join(rootFolder, "about"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			log.Fatalf("failed to create dir %q: %v", dir, err)
		}
	}

	// ── public/index.html ────────────────────────────────────────────────────
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
	if err = blog.Index(homeSEO, data, blog.Body(data)).Render(ctx, index); err != nil {
		log.Fatalf("could not render public/index.html: %v", err)
	}

	// ── public/blog/index.html ───────────────────────────────────────────────
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
	if err = blog.Index(blogSEO, data, posts.Body(data)).Render(ctx, blogIndex); err != nil {
		log.Fatalf("failed to write blog index page: %v", err)
	}

	// ── public/projects/index.html ───────────────────────────────────────────
	projFile, err := os.Create(path.Join(rootFolder, "projects", "index.html"))
	if err != nil {
		log.Fatalf("failed to create public/projects/index.html: %v", err)
	}
	projSEO := blog.SEOMeta{
		Title:        "Open-Source Projects – " + data.Name,
		Description:  "Open-source Go tools and Kubernetes utilities by Željko Benović — platform automation, DevSecOps, and SRE tooling.",
		CanonicalURL: siteURL + "/projects/",
		OGType:       "website",
	}
	if err = blog.Index(projSEO, data, blog.Projects(data)).Render(ctx, projFile); err != nil {
		log.Fatalf("failed to write projects page: %v", err)
	}

	// ── public/about/index.html ──────────────────────────────────────────────
	aboutFile, err := os.Create(path.Join(rootFolder, "about", "index.html"))
	if err != nil {
		log.Fatalf("failed to create public/about/index.html: %v", err)
	}
	aboutSEO := blog.SEOMeta{
		Title:        "About – " + data.Name,
		Description:  "Staff Platform Engineer specializing in Kubernetes, SRE, and DevSecOps. Career history, expertise, and background.",
		CanonicalURL: siteURL + "/about/",
		OGType:       "website",
	}
	if err = blog.Index(aboutSEO, data, blog.About(data)).Render(ctx, aboutFile); err != nil {
		log.Fatalf("failed to write about page: %v", err)
	}

	// ── public/404.html ──────────────────────────────────────────────────────
	notFoundFile, err := os.Create(path.Join(rootFolder, "404.html"))
	if err != nil {
		log.Fatalf("failed to create public/404.html: %v", err)
	}
	notFoundSEO := blog.SEOMeta{
		Title:       "404 – CrashLoopBackOff | " + data.Name,
		Description: "The resource you requested does not exist in this namespace.",
		OGType:      "website",
	}
	if err = blog.Index(notFoundSEO, data, blog.NotFound(data)).Render(ctx, notFoundFile); err != nil {
		log.Fatalf("failed to write 404 page: %v", err)
	}

	// ── Individual blog posts ────────────────────────────────────────────────
	for i, post := range data.Posts {
		dir := path.Join(blogFolder, post.Date.Format("2006/01/02"), post.Slug)
		if err = os.MkdirAll(dir, 0755); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("failed to create dir %q: %v", dir, err)
		}

		name := path.Join(dir, "index.html")
		outFile, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}

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
		var buf bytes.Buffer
		if err := gm.Convert(postContent, &buf); err != nil {
			log.Fatalf("failed to convert markdown to HTML: %v", err)
		}

		postURL := siteURL + post.URL
		postDesc := extractDescription(string(postContent))
		postSEO := blog.SEOMeta{
			Title:        post.Title + " – " + data.Name,
			Description:  postDesc,
			CanonicalURL: postURL,
			OGType:       "article",
			JSONLD:       blogPostingJSONLD(post, data, postURL, postDesc, siteURL),
		}

		// Compute prev/next (by index position).
		var prevPost, nextPost blog.Post
		if i > 0 {
			prevPost = data.Posts[i-1]
		}
		if i < len(data.Posts)-1 {
			nextPost = data.Posts[i+1]
		}

		content := posts.Unsafe(buf.String())
		if err = blog.Index(postSEO, data, posts.Post(content, post, data, prevPost, nextPost)).Render(ctx, outFile); err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
	}

	// Generate robots.txt and sitemap.xml.
	generateRobotsTxt(rootFolder, siteURL)
	if siteURL != "" {
		generateSitemap(rootFolder, siteURL, data.Posts)
	}
}

// h1FromFile reads a Markdown file and returns the text of the first H1 heading.
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

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func extractDescription(md string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") {
			continue
		}
		for _, marker := range []string{"**", "*", "`", "_"} {
			line = strings.ReplaceAll(line, marker, "")
		}
		return truncate(line, 160)
	}
	return ""
}

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
	desc := data.SEODescription
	if desc == "" {
		desc = data.Bio
	}
	if desc != "" {
		obj["description"] = desc
	}
	var skills []string
	for _, cat := range data.Skills {
		skills = append(skills, cat.Items...)
	}
	if len(skills) > 0 {
		obj["knowsAbout"] = skills
	}
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
	writeURL(siteURL+"/projects/", "", "monthly", "0.7")
	writeURL(siteURL+"/about/", "", "monthly", "0.7")

	for _, p := range blogPosts {
		postURL := siteURL + p.URL
		writeURL(postURL, p.Date.Format("2006-01-02"), "yearly", "0.6")
	}

	sb.WriteString("</urlset>\n")

	if err := os.WriteFile(path.Join(rootFolder, "sitemap.xml"), []byte(sb.String()), 0644); err != nil {
		log.Fatalf("failed to write sitemap.xml: %v", err)
	}
}
