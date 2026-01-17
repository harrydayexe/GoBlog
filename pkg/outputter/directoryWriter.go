package outputter

import (
	"fmt"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

type DirectoryWriterConfig struct {
	OutputPath string
}

type DirectoryWriterOption func(*DirectoryWriterConfig)

type DirectoryWriter struct {
	config DirectoryWriterConfig
}

func NewDirectoryWriter(outputDir string, opts ...DirectoryWriterOption) DirectoryWriter {
	config := DirectoryWriterConfig{
		OutputPath: outputDir,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return NewDirectoryWriterWithConfig(config)
}

func NewDirectoryWriterWithConfig(config DirectoryWriterConfig) DirectoryWriter {
	return DirectoryWriter{
		config: config,
	}
}

func (dw DirectoryWriter) HandleGeneratedBlog(blog *generator.GeneratedBlog) error {
	for slug, post := range blog.Posts {
		fmt.Printf("====== Post: %s ======\n", slug)
		fmt.Printf("%s\n\n", post)
	}

	return nil
}
