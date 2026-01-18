package generator_test

import (
	"context"
	"fmt"
	"os"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// Example demonstrates basic usage of the generator package.
func Example() {
	// Create a generator with posts from testdata directory
	fsys := os.DirFS("testdata")
	gen := generator.New(fsys, config.WithRawOutput())

	// Generate the blog
	ctx := context.Background()
	blog, err := gen.Generate(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Generated %d post(s)\n", len(blog.Posts))
	// Output: Generated 3 post(s)
}

// Example_rawOutput demonstrates using raw output mode.
func Example_rawOutput() {
	// Raw output mode generates HTML without templates
	fsys := os.DirFS("testdata")
	gen := generator.New(fsys, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Raw HTML mode enabled")
	fmt.Printf("Posts generated: %d\n", len(blog.Posts))
	// Output: Raw HTML mode enabled
	// Posts generated: 3
}

// ExampleNew demonstrates creating a new generator.
func ExampleNew() {
	// Create a generator with default configuration
	fsys := os.DirFS("testdata")
	gen := generator.New(fsys)

	if gen != nil {
		fmt.Println("Generator created")
	}
	// Output: Generator created
}

// ExampleGenerator_Generate demonstrates generating a blog.
func ExampleGenerator_Generate() {
	fsys := os.DirFS("testdata")
	gen := generator.New(fsys, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Blog generation complete")
	fmt.Printf("Total posts: %d\n", len(blog.Posts))
	// Output: Blog generation complete
	// Total posts: 3
}

// ExampleGenerator_DebugConfig demonstrates debugging generator configuration.
func ExampleGenerator_DebugConfig() {
	fsys := os.DirFS("testdata")
	gen := generator.New(fsys, config.WithRawOutput())

	ctx := context.Background()
	gen.DebugConfig(ctx)

	fmt.Println("Debug output logged")
	// Output: Debug output logged
}

// ExampleWithRawOutput demonstrates using the WithRawOutput option.
func ExampleWithRawOutput() {
	fsys := os.DirFS("testdata")

	// Enable raw output mode (no templates)
	gen := generator.New(fsys, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Raw output enabled")
	fmt.Printf("Generated posts: %d\n", len(blog.Posts))
	// Output: Raw output enabled
	// Generated posts: 3
}

// ExampleGeneratedBlog demonstrates working with a GeneratedBlog.
func ExampleGeneratedBlog() {
	fsys := os.DirFS("testdata")
	gen := generator.New(fsys, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Access posts map and verify structure
	fmt.Printf("Total posts: %d\n", len(blog.Posts))
	fmt.Printf("Has Posts map: %t\n", blog.Posts != nil)
	fmt.Printf("Has Tags map: %t\n", blog.Tags != nil)
	// Output: Total posts: 3
	// Has Posts map: true
	// Has Tags map: true
}
