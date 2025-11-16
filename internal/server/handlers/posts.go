package handlers

import (
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/harrydayexe/GoBlog/internal/server/components"
	"github.com/harrydayexe/GoBlog/internal/server/content"
)

// PostHandlers handles blog post requests
type PostHandlers struct {
	loader *content.Loader
	htmx   *htmx.HTMX
}

// NewPostHandlers creates a new post handler
func NewPostHandlers(loader *content.Loader) *PostHandlers {
	return &PostHandlers{
		loader: loader,
		htmx:   htmx.New(),
	}
}

// HandleIndex handles the blog index page
func (h *PostHandlers) HandleIndex(w http.ResponseWriter, r *http.Request) {
	posts := h.loader.GetAll()

	// For now, show first 10 posts
	perPage := 10
	displayPosts := posts
	if len(posts) > perPage {
		displayPosts = posts[:perPage]
	}

	hasMore := len(posts) > perPage

	component := components.Index(displayPosts, 1, hasMore)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// HandlePost handles individual post requests
func (h *PostHandlers) HandlePost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	post, err := h.loader.GetBySlug(slug)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	component := components.Post(post)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// HandlePagination handles pagination requests (HTMX partial)
func (h *PostHandlers) HandlePagination(w http.ResponseWriter, r *http.Request) {
	// Parse page number from query
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		// Simple page parsing - in production use strconv
		page = parseInt(pageStr, 1)
	}

	perPage := 10
	posts, _, err := h.loader.GetPaginated(page, perPage)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Render just the post list (partial update)
	component := components.PostList(posts)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// parseInt parses an int from a string with a default value
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}

	val := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			val = val*10 + int(c-'0')
		} else {
			return defaultVal
		}
	}

	if val == 0 {
		return defaultVal
	}

	return val
}
