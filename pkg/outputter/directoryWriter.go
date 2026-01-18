package outputter

import (
	"fmt"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

type DirectoryWriterConfig struct {
	config.CommonConfig
	OutputPath string
}

type DirectoryWriter struct {
	config DirectoryWriterConfig
}

func NewDirectoryWriter(outputDir string, opts ...config.CommonOption) DirectoryWriter {
	config := DirectoryWriterConfig{
		OutputPath: outputDir,
	}

	for _, opt := range opts {
		opt(&config.CommonConfig)
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
