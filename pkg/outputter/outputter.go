package outputter

import (
	"context"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// Outputter defines the interface for handling generated blog content.
//
// Implementations process the GeneratedBlog structure from the generator
// package and perform output operations such as writing files to disk,
// serving content via HTTP, storing in a database, or any other desired
// output mechanism.
//
// Implementations should be safe for concurrent use.
type Outputter interface {
	// HandleGeneratedBlog processes the generated blog content and performs
	// the implementation-specific output operation.
	//
	// The blog parameter contains all generated content including posts, tags,
	// and the index page as HTML byte slices.
	//
	// Returns an error if the output operation fails (e.g., filesystem errors,
	// network errors, permission issues).
	HandleGeneratedBlog(context.Context, *generator.GeneratedBlog) error
}
