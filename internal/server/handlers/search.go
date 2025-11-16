package handlers

import (
	"net/http"

	"github.com/harrydayexe/GoBlog/internal/server/components"
	"github.com/harrydayexe/GoBlog/internal/server/content"
	"github.com/harrydayexe/GoBlog/internal/server/search"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// SearchHandlers handles search requests
type SearchHandlers struct {
	loader *content.Loader
	index  *search.Index
}

// NewSearchHandlers creates a new search handler
func NewSearchHandlers(loader *content.Loader, index *search.Index) *SearchHandlers {
	return &SearchHandlers{
		loader: loader,
		index:  index,
	}
}

// HandleSearch handles search requests (HTMX partial)
func (h *SearchHandlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		// Return all posts if no query
		posts := h.loader.GetAll()
		component := components.PostList(posts)
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	var posts = h.loader.Search(query)

	// If search index is available, use it instead
	if h.index != nil {
		results, err := h.index.Search(query, 20)
		if err == nil && len(results) > 0 {
			// Convert search results to posts
			posts = make([]*models.Post, 0, len(results))
			for _, result := range results {
				if post, err := h.loader.GetBySlug(result.Slug); err == nil {
					posts = append(posts, post)
				}
			}
		}
	}

	component := components.SearchResults(posts, query)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
