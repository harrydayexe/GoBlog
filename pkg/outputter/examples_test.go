package outputter_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/outputter"
)

// Example demonstrates basic usage of the outputter package.
func Example() {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "blog-example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create some sample blog content
	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"hello-world": []byte("<h1>Hello World</h1><p>My first post</p>"),
		},
		Index: []byte("<h1>My Blog</h1><p>Welcome!</p>"),
		Tags: map[string][]byte{
			"intro": []byte("<h1>Intro Posts</h1>"),
		},
	}

	// Create a DirectoryWriter and write the blog
	writer := outputter.NewDirectoryWriter(tempDir)
	if err := writer.HandleGeneratedBlog(blog); err != nil {
		log.Fatal(err)
	}

	// Verify files were created
	files := []string{"index.html", "hello-world.html", "tags/intro.html"}
	for _, file := range files {
		path := filepath.Join(tempDir, file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("%s created\n", file)
		}
	}

	// Output:
	// index.html created
	// hello-world.html created
	// tags/intro.html created
}

// ExampleNewDirectoryWriter demonstrates creating a DirectoryWriter
// with default settings.
func ExampleNewDirectoryWriter() {
	writer := outputter.NewDirectoryWriter("/var/www/blog")

	fmt.Printf("Writer created with output path: %s\n", "/var/www/blog")
	_ = writer // Use the writer

	// Output:
	// Writer created with output path: /var/www/blog
}

// ExampleNewDirectoryWriter_withRawOutput demonstrates creating a
// DirectoryWriter with the RawOutput option enabled.
func ExampleNewDirectoryWriter_withRawOutput() {
	// Create writer with RawOutput option - this will skip creating
	// the tags directory
	writer := outputter.NewDirectoryWriter(
		"/var/www/blog",
		config.WithRawOutput(),
	)

	fmt.Printf("Writer created with RawOutput enabled\n")
	_ = writer // Use the writer

	// Output:
	// Writer created with RawOutput enabled
}

// ExampleDirectoryWriter_HandleGeneratedBlog demonstrates writing
// generated blog content to disk.
func ExampleDirectoryWriter_HandleGeneratedBlog() {
	tempDir, err := os.MkdirTemp("", "blog-handle")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create blog content
	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"my-post": []byte("<h1>My Post</h1>"),
		},
		Index: []byte("<h1>Blog Index</h1>"),
		Tags: map[string][]byte{
			"tech": []byte("<h1>Tech Posts</h1>"),
		},
	}

	// Write to disk
	writer := outputter.NewDirectoryWriter(tempDir)
	if err := writer.HandleGeneratedBlog(blog); err != nil {
		log.Fatal(err)
	}

	// Verify index was created
	indexPath := filepath.Join(tempDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		fmt.Println("Blog written successfully")
	}

	// Output:
	// Blog written successfully
}

// Example_fullWorkflow demonstrates a complete workflow of generating
// and outputting a blog. This example uses RawOutput mode since templates
// are not yet fully implemented in the generator.
func Example_fullWorkflow() {
	// Setup temporary directories
	postsDir, err := os.MkdirTemp("", "posts")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(postsDir)

	outputDir, err := os.MkdirTemp("", "output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	// Create a sample markdown post
	postContent := `---
title: Getting Started
description: A beginner's guide to getting started
tags: [tutorial, beginner]
date: 2024-01-01
---

# Getting Started

This is a simple blog post.
`
	postPath := filepath.Join(postsDir, "getting-started.md")
	if err := os.WriteFile(postPath, []byte(postContent), 0644); err != nil {
		log.Fatal(err)
	}

	// Generate the blog using os.DirFS to create a filesystem
	// Use RawOutput since templates are not fully implemented yet
	gen := generator.New(os.DirFS(postsDir), nil, config.WithRawOutput())
	blog, err := gen.Generate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Output the blog to disk with RawOutput to match generator settings
	writer := outputter.NewDirectoryWriter(outputDir, config.WithRawOutput())
	if err := writer.HandleGeneratedBlog(blog); err != nil {
		log.Fatal(err)
	}

	// Verify the output
	indexPath := filepath.Join(outputDir, "index.html")
	postPath = filepath.Join(outputDir, "getting-started.html")

	if _, err := os.Stat(indexPath); err == nil {
		fmt.Println("index.html created")
	}
	if _, err := os.Stat(postPath); err == nil {
		fmt.Println("getting-started.html created")
	}

	// Output:
	// index.html created
	// getting-started.html created
}

// Example_rawOutputMode demonstrates using RawOutput mode to skip
// template wrapping and tag generation.
func Example_rawOutputMode() {
	tempDir, err := os.MkdirTemp("", "blog-raw")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create blog content
	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"post": []byte("<article>Raw HTML content</article>"),
		},
		Index: []byte("<div>Index content</div>"),
		Tags: map[string][]byte{
			"tag": []byte("<div>Tag content</div>"),
		},
	}

	// Write with RawOutput enabled - tags directory won't be created
	writer := outputter.NewDirectoryWriter(tempDir, config.WithRawOutput())
	if err := writer.HandleGeneratedBlog(blog); err != nil {
		log.Fatal(err)
	}

	// Check what was created
	if _, err := os.Stat(filepath.Join(tempDir, "index.html")); err == nil {
		fmt.Println("index.html created")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "post.html")); err == nil {
		fmt.Println("post.html created")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "tags")); os.IsNotExist(err) {
		fmt.Println("tags directory not created (as expected)")
	}

	// Output:
	// index.html created
	// post.html created
	// tags directory not created (as expected)
}

// Example_multipleWrites demonstrates that DirectoryWriter is idempotent
// and can safely write to the same directory multiple times.
func Example_multipleWrites() {
	tempDir, err := os.MkdirTemp("", "blog-multi")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	writer := outputter.NewDirectoryWriter(tempDir)

	// First write
	blog1 := &generator.GeneratedBlog{
		Posts: map[string][]byte{"post": []byte("Version 1")},
		Index: []byte("Index V1"),
		Tags:  make(map[string][]byte),
	}
	if err := writer.HandleGeneratedBlog(blog1); err != nil {
		log.Fatal(err)
	}
	fmt.Println("First write completed")

	// Second write - updates the same files
	blog2 := &generator.GeneratedBlog{
		Posts: map[string][]byte{"post": []byte("Version 2")},
		Index: []byte("Index V2"),
		Tags:  make(map[string][]byte),
	}
	if err := writer.HandleGeneratedBlog(blog2); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Second write completed")

	// Read final content
	content, _ := os.ReadFile(filepath.Join(tempDir, "post.html"))
	fmt.Printf("Final content: %s\n", string(content))

	// Output:
	// First write completed
	// Second write completed
	// Final content: Version 2
}
