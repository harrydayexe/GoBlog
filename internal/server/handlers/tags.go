package handlers

import (
	"net/http"

	"github.com/harrydayexe/GoBlog/internal/server/components"
	"github.com/harrydayexe/GoBlog/internal/server/content"
)

// TagHandlers handles tag-related requests
type TagHandlers struct {
	loader *content.Loader
}

// NewTagHandlers creates a new tag handler
func NewTagHandlers(loader *content.Loader) *TagHandlers {
	return &TagHandlers{
		loader: loader,
	}
}

// HandleTag handles tag page requests
func (h *TagHandlers) HandleTag(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	if tag == "" {
		http.Error(w, "Tag not specified", http.StatusBadRequest)
		return
	}

	posts := h.loader.GetByTag(tag)

	component := components.TagPage(tag, posts)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// HandleTagList handles tag list requests (HTMX partial)
func (h *TagHandlers) HandleTagList(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		// Return all posts if no tag
		posts := h.loader.GetAll()
		component := components.PostList(posts)
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	posts := h.loader.GetByTag(tag)

	component := components.PostList(posts)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
