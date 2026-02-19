package server

import (
	"context"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

type Cache interface {
	Get(ctx context.Context) (*generator.GeneratedBlog, error)
	Set(ctx context.Context, blog *generator.GeneratedBlog) error
	Clear(ctx context.Context) error
}
