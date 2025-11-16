package content

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/harrydayexe/GoBlog/internal/gen/parser"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Loader handles loading and managing blog posts
type Loader struct {
	contentPath string
	parser      *parser.Parser
	cache       *Cache
	mu          sync.RWMutex
	posts       []*models.Post
	watcher     *fsnotify.Watcher
	watcherDone chan bool
	onReload    func(int) // Callback when content is reloaded
}

// NewLoader creates a new content loader
func NewLoader(contentPath string, cache *Cache) (*Loader, error) {
	p := parser.New()

	loader := &Loader{
		contentPath: contentPath,
		parser:      p,
		cache:       cache,
		posts:       make([]*models.Post, 0),
	}

	// Initial load
	if err := loader.LoadAll(); err != nil {
		return nil, fmt.Errorf("failed to load posts: %w", err)
	}

	return loader, nil
}

// LoadAll loads all markdown posts from the content directory
func (l *Loader) LoadAll() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	posts, err := l.parser.ParseAllPosts(l.contentPath)
	if err != nil {
		return fmt.Errorf("failed to parse directory: %w", err)
	}

	// Filter out drafts and sort by date (newest first)
	published := make([]*models.Post, 0)
	for _, post := range posts {
		if post.IsPublished() {
			published = append(published, post)
		}
	}

	sort.Slice(published, func(i, j int) bool {
		return published[i].Date.After(published[j].Date)
	})

	l.posts = published
	return nil
}

// GetBySlug retrieves a post by its slug
func (l *Loader) GetBySlug(slug string) (*models.Post, error) {
	// Check cache first
	if l.cache != nil {
		if post := l.cache.Get(slug); post != nil {
			return post, nil
		}
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, post := range l.posts {
		if post.Slug == slug {
			// Cache the post
			if l.cache != nil {
				l.cache.Set(slug, post)
			}
			return post, nil
		}
	}

	return nil, fmt.Errorf("post not found: %s", slug)
}

// GetAll returns all published posts
func (l *Loader) GetAll() []*models.Post {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Return a copy to prevent external modification
	posts := make([]*models.Post, len(l.posts))
	copy(posts, l.posts)
	return posts
}

// GetPaginated returns a paginated slice of posts
func (l *Loader) GetPaginated(page, perPage int) ([]*models.Post, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := len(l.posts)
	if page < 1 {
		page = 1
	}

	start := (page - 1) * perPage
	if start >= total {
		return []*models.Post{}, total, nil
	}

	end := start + perPage
	if end > total {
		end = total
	}

	posts := make([]*models.Post, end-start)
	copy(posts, l.posts[start:end])

	return posts, total, nil
}

// GetByTag returns all posts with a specific tag
func (l *Loader) GetByTag(tag string) []*models.Post {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]*models.Post, 0)
	for _, post := range l.posts {
		if post.HasTag(tag) {
			result = append(result, post)
		}
	}

	return result
}

// GetAllTags returns all unique tags from all posts
func (l *Loader) GetAllTags() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	tagMap := make(map[string]int)
	for _, post := range l.posts {
		for _, tag := range post.Tags {
			tagMap[tag]++
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags
}

// Search performs a simple search across post titles, descriptions, and content
func (l *Loader) Search(query string) []*models.Post {
	if query == "" {
		return []*models.Post{}
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	query = strings.ToLower(query)
	result := make([]*models.Post, 0)

	for _, post := range l.posts {
		if l.matchesQuery(post, query) {
			result = append(result, post)
		}
	}

	return result
}

// matchesQuery checks if a post matches the search query
func (l *Loader) matchesQuery(post *models.Post, query string) bool {
	title := strings.ToLower(post.Title)
	description := strings.ToLower(post.Description)
	content := strings.ToLower(post.Content)

	return strings.Contains(title, query) ||
		strings.Contains(description, query) ||
		strings.Contains(content, query)
}

// Reload reloads all posts from disk
func (l *Loader) Reload() error {
	if err := l.LoadAll(); err != nil {
		return err
	}

	// Clear cache on reload
	if l.cache != nil {
		l.cache.Clear()
	}

	return nil
}

// SetReloadCallback sets the callback to be called when content is reloaded
func (l *Loader) SetReloadCallback(callback func(int)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onReload = callback
}

// Watch starts watching the content directory for changes
func (l *Loader) Watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	l.mu.Lock()
	l.watcher = watcher
	l.watcherDone = make(chan bool)
	l.mu.Unlock()

	// Watch the content directory
	err = watcher.Add(l.contentPath)
	if err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Start watching in a goroutine
	go l.watchLoop()

	log.Printf("Watching for changes in %s", l.contentPath)
	return nil
}

// watchLoop processes file system events
func (l *Loader) watchLoop() {
	// Debounce timer to avoid reloading too frequently
	var debounceTimer *time.Timer
	debounceDuration := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-l.watcher.Events:
			if !ok {
				return
			}

			// Only react to .md and .markdown files
			if !strings.HasSuffix(event.Name, ".md") && !strings.HasSuffix(event.Name, ".markdown") {
				continue
			}

			// Debounce: reset timer if it exists, otherwise create new one
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(debounceDuration, func() {
				log.Printf("Detected change in %s, reloading...", event.Name)
				if err := l.Reload(); err != nil {
					log.Printf("Error reloading content: %v", err)
				} else {
					l.mu.RLock()
					count := len(l.posts)
					callback := l.onReload
					l.mu.RUnlock()

					if callback != nil {
						callback(count)
					}
					log.Printf("Reloaded %d posts", count)
				}
			})

		case err, ok := <-l.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)

		case <-l.watcherDone:
			return
		}
	}
}

// StopWatching stops the file watcher
func (l *Loader) StopWatching() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.watcher != nil {
		close(l.watcherDone)
		if err := l.watcher.Close(); err != nil {
			return fmt.Errorf("failed to close watcher: %w", err)
		}
		l.watcher = nil
	}

	return nil
}

// findMarkdownFiles recursively finds all markdown files in a directory
func findMarkdownFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
