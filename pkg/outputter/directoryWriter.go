package outputter

import (
	"os"
	"path/filepath"

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
	writeMapToFiles(blog.Posts, dw.config.OutputPath)
	if !dw.config.RawOutput {
		writeMapToFiles(blog.Tags, filepath.Join(dw.config.OutputPath, "tags"))
	}
	if err := os.WriteFile(dw.config.OutputPath, blog.Index, 0644); err != nil {
		return err
	}

	return nil
}

func writeMapToFiles(data map[string][]byte, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for filename, content := range data {
		htmlFile := filename + ".html"
		path := filepath.Join(outputDir, htmlFile)
		if err := os.WriteFile(path, content, 0644); err != nil {
			return err
		}
	}
	return nil
}
