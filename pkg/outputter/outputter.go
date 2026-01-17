package outputter

import "github.com/harrydayexe/GoBlog/v2/pkg/generator"

type Outputter interface {
	HandleGeneratedBlog(*generator.GeneratedBlog) error
}
