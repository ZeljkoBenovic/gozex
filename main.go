package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

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
		// get post title
		postTitle := cases.Title(language.English).String(strings.ReplaceAll(postSlug, "-", " "))
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

	// render public/index.html
	if err = blog.Index(data.Name, blog.Body(data)).Render(ctx, index); err != nil {
		log.Fatalf("could not render public/index.html: %v", err)
	}

	// create public/blog/index.html
	blogIndex, err := os.Create(path.Join(blogFolder, "index.html"))
	if err != nil {
		log.Fatalf("failed to create public/blog/index.html file: %v", err)
	}

	// render public/blog/index.html
	err = blog.Index(data.Name, posts.Body(data)).Render(ctx, blogIndex)
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
		f, err := os.Create(name)
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

		// Create an unsafe component containing raw HTML.
		content := posts.Unsafe(buf.String())

		// Use templ to render the template containing the raw HTML.
		if err = blog.Index(data.Name, posts.Post(content)).Render(ctx, f); err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
	}
}
