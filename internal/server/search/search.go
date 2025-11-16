package search

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Index wraps a Bleve search index
type Index struct {
	index bleve.Index
	path  string
}

// NewIndex creates or opens a Bleve search index
func NewIndex(indexPath string, rebuild bool) (*Index, error) {
	var index bleve.Index
	var err error

	// If rebuild is requested or index doesn't exist, create new
	if rebuild || !indexExists(indexPath) {
		// Delete existing index if rebuilding
		if rebuild && indexExists(indexPath) {
			if err := os.RemoveAll(indexPath); err != nil {
				return nil, fmt.Errorf("failed to remove existing index: %w", err)
			}
		}

		// Create new index with default mapping
		mapping := buildIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	} else {
		// Open existing index
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open index: %w", err)
		}
	}

	return &Index{
		index: index,
		path:  indexPath,
	}, nil
}

// IndexPost indexes a single post
func (idx *Index) IndexPost(post *models.Post) error {
	return idx.index.Index(post.Slug, post)
}

// IndexPosts indexes multiple posts
func (idx *Index) IndexPosts(posts []*models.Post) error {
	batch := idx.index.NewBatch()

	for _, post := range posts {
		if err := batch.Index(post.Slug, post); err != nil {
			return fmt.Errorf("failed to add post to batch: %w", err)
		}
	}

	if err := idx.index.Batch(batch); err != nil {
		return fmt.Errorf("failed to execute batch: %w", err)
	}

	return nil
}

// Search performs a search query
func (idx *Index) Search(query string, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Create a query
	q := bleve.NewMatchQuery(query)
	search := bleve.NewSearchRequest(q)
	search.Size = limit
	search.Fields = []string{"title", "description", "slug"}

	// Execute search
	searchResults, err := idx.index.Search(search)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	results := make([]*SearchResult, len(searchResults.Hits))
	for i, hit := range searchResults.Hits {
		results[i] = &SearchResult{
			Slug:        hit.ID,
			Score:       hit.Score,
			Title:       getStringField(hit.Fields, "title"),
			Description: getStringField(hit.Fields, "description"),
		}
	}

	return results, nil
}

// Delete removes a post from the index
func (idx *Index) Delete(slug string) error {
	return idx.index.Delete(slug)
}

// Close closes the index
func (idx *Index) Close() error {
	return idx.index.Close()
}

// Count returns the number of documents in the index
func (idx *Index) Count() (uint64, error) {
	return idx.index.DocCount()
}

// SearchResult represents a search result
type SearchResult struct {
	Slug        string
	Score       float64
	Title       string
	Description string
}

// buildIndexMapping creates the index mapping for blog posts
func buildIndexMapping() mapping.IndexMapping {
	// Create a mapping for blog posts
	postMapping := bleve.NewDocumentMapping()

	// Title field - searchable text
	titleField := bleve.NewTextFieldMapping()
	titleField.Analyzer = "en"
	postMapping.AddFieldMappingsAt("title", titleField)

	// Description field - searchable text
	descriptionField := bleve.NewTextFieldMapping()
	descriptionField.Analyzer = "en"
	postMapping.AddFieldMappingsAt("description", descriptionField)

	// Content field - searchable text
	contentField := bleve.NewTextFieldMapping()
	contentField.Analyzer = "en"
	postMapping.AddFieldMappingsAt("content", contentField)

	// Slug field - keyword (exact match)
	slugField := bleve.NewKeywordFieldMapping()
	postMapping.AddFieldMappingsAt("slug", slugField)

	// Tags field - keyword array
	tagsField := bleve.NewKeywordFieldMapping()
	postMapping.AddFieldMappingsAt("tags", tagsField)

	// Date field - datetime
	dateField := bleve.NewDateTimeFieldMapping()
	postMapping.AddFieldMappingsAt("date", dateField)

	// Create index mapping
	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("_default", postMapping)

	return indexMapping
}

// indexExists checks if an index exists at the given path
func indexExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// getStringField safely extracts a string field from search results
func getStringField(fields map[string]interface{}, key string) string {
	if val, ok := fields[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
