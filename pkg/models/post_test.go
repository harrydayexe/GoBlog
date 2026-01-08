package models

import (
	"html/template"
	"testing"
	"time"
)

// TestPost_Validate tests the post validation logic
func TestPost_Validate(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name      string
		post      Post
		expectErr bool
		errText   string
	}{
		{
			name: "valid post",
			post: Post{
				Title:       "Test Post",
				Date:        now,
				Description: "A test post",
			},
			expectErr: false,
		},
		{
			name: "missing title",
			post: Post{
				Title:       "",
				Date:        now,
				Description: "A test post",
				SourcePath:  "test.md",
			},
			expectErr: true,
			errText:   "title",
		},
		{
			name: "missing date",
			post: Post{
				Title:       "Test Post",
				Date:        time.Time{},
				Description: "A test post",
				SourcePath:  "test.md",
			},
			expectErr: true,
			errText:   "date",
		},
		{
			name: "missing description",
			post: Post{
				Title:       "Test Post",
				Date:        now,
				Description: "",
				SourcePath:  "test.md",
			},
			expectErr: true,
			errText:   "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.post.Validate()
			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errText != "" && !contains(err.Error(), tt.errText) {
					t.Errorf("expected error to contain %q, got %q", tt.errText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPost_GenerateSlug tests the slug generation logic
func TestPost_GenerateSlug(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		post         Post
		expectedSlug string
	}{
		{
			name: "generate from title",
			post: Post{
				Title: "Hello World",
			},
			expectedSlug: "hello-world",
		},
		{
			name: "existing slug preserved",
			post: Post{
				Title: "Hello World",
				Slug:  "custom-slug",
			},
			expectedSlug: "custom-slug",
		},
		{
			name: "title with special characters",
			post: Post{
				Title: "Go: A Programming Language!",
			},
			expectedSlug: "go-a-programming-language",
		},
		{
			name: "title with numbers",
			post: Post{
				Title: "Top 10 Tips for 2024",
			},
			expectedSlug: "top-10-tips-for-2024",
		},
		{
			name: "title with underscores",
			post: Post{
				Title: "Test_Post_Title",
			},
			expectedSlug: "test-post-title",
		},
		{
			name: "title with multiple spaces",
			post: Post{
				Title: "Multiple    Spaces    Here",
			},
			expectedSlug: "multiple-spaces-here",
		},
		{
			name: "fallback to filename",
			post: Post{
				Title:      "",
				SourcePath: "/path/to/my-post.md",
			},
			expectedSlug: "my-post",
		},
		{
			name: "unicode characters removed",
			post: Post{
				Title: "Café ☕ in München 🇩🇪",
			},
			expectedSlug: "caf-in-mnchen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.post.GenerateSlug()
			if tt.post.Slug != tt.expectedSlug {
				t.Errorf("expected slug %q, got %q", tt.expectedSlug, tt.post.Slug)
			}
		})
	}
}

// TestSlugify tests the slugify helper function
func TestSlugify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"UPPERCASE", "uppercase"},
		{"Mixed-Case_String", "mixed-case-string"},
		{"Special!@#$%Characters", "specialcharacters"},
		{"Numbers 123", "numbers-123"},
		{"Multiple---Hyphens", "multiple-hyphens"},
		{"--Leading-and-Trailing--", "leading-and-trailing"},
		{"", ""},
		{"    ", ""},
		{"a", "a"},
		{"123", "123"},
		{"---", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPost_HasTag tests the tag checking logic
func TestPost_HasTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		tags     []string
		checkTag string
		expected bool
	}{
		{
			name:     "has tag - exact match",
			tags:     []string{"go", "web", "tutorial"},
			checkTag: "go",
			expected: true,
		},
		{
			name:     "has tag - case insensitive",
			tags:     []string{"Go", "Web", "Tutorial"},
			checkTag: "go",
			expected: true,
		},
		{
			name:     "does not have tag",
			tags:     []string{"go", "web"},
			checkTag: "python",
			expected: false,
		},
		{
			name:     "empty tags",
			tags:     []string{},
			checkTag: "go",
			expected: false,
		},
		{
			name:     "nil tags",
			tags:     nil,
			checkTag: "go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			post := Post{Tags: tt.tags}
			if got := post.HasTag(tt.checkTag); got != tt.expected {
				t.Errorf("HasTag(%q) = %v, want %v", tt.checkTag, got, tt.expected)
			}
		})
	}
}

// TestPost_FormattedDate tests date formatting
func TestPost_FormattedDate(t *testing.T) {
	t.Parallel()
	date := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	post := Post{Date: date}

	expected := "March 15, 2024"
	if got := post.FormattedDate(); got != expected {
		t.Errorf("FormattedDate() = %q, want %q", got, expected)
	}
}

// TestPost_ShortDate tests short date formatting
func TestPost_ShortDate(t *testing.T) {
	t.Parallel()
	date := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	post := Post{Date: date}

	expected := "2024-03-15"
	if got := post.ShortDate(); got != expected {
		t.Errorf("ShortDate() = %q, want %q", got, expected)
	}
}

// TestPostList_FilterByTag tests filtering posts by tag
func TestPostList_FilterByTag(t *testing.T) {
	t.Parallel()
	now := time.Now()
	posts := PostList{
		{Title: "Post 1", Date: now, Description: "desc1", Tags: []string{"go", "web"}},
		{Title: "Post 2", Date: now, Description: "desc2", Tags: []string{"python", "api"}},
		{Title: "Post 3", Date: now, Description: "desc3", Tags: []string{"go", "cli"}},
		{Title: "Post 4", Date: now, Description: "desc4", Tags: []string{"rust"}},
	}

	tests := []struct {
		tag      string
		expected int
	}{
		{"go", 2},
		{"web", 1},
		{"python", 1},
		{"rust", 1},
		{"java", 0},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			t.Parallel()
			filtered := posts.FilterByTag(tt.tag)
			if len(filtered) != tt.expected {
				t.Errorf("expected %d posts with tag %q, got %d", tt.expected, tt.tag, len(filtered))
			}

			for _, post := range filtered {
				if !post.HasTag(tt.tag) {
					t.Errorf("post %q should have tag %q", post.Title, tt.tag)
				}
			}
		})
	}
}

// TestPostList_SortByDate tests sorting posts by date
func TestPostList_SortByDate(t *testing.T) {
	t.Parallel()
	date1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	posts := PostList{
		{Title: "Post 1", Date: date2, Description: "desc1"},
		{Title: "Post 2", Date: date1, Description: "desc2"},
		{Title: "Post 3", Date: date3, Description: "desc3"},
	}

	posts.SortByDate()

	// Should be sorted newest first
	if !posts[0].Date.Equal(date3) {
		t.Errorf("first post should have date %v, got %v", date3, posts[0].Date)
	}
	if !posts[1].Date.Equal(date2) {
		t.Errorf("second post should have date %v, got %v", date2, posts[1].Date)
	}
	if !posts[2].Date.Equal(date1) {
		t.Errorf("third post should have date %v, got %v", date1, posts[2].Date)
	}
}

// TestPostList_GetAllTags tests getting unique tags
func TestPostList_GetAllTags(t *testing.T) {
	t.Parallel()
	now := time.Now()
	posts := PostList{
		{Title: "Post 1", Date: now, Description: "desc1", Tags: []string{"go", "web"}},
		{Title: "Post 2", Date: now, Description: "desc2", Tags: []string{"python", "api"}},
		{Title: "Post 3", Date: now, Description: "desc3", Tags: []string{"go", "cli"}},
		{Title: "Post 4", Date: now, Description: "desc4", Tags: []string{}},
	}

	tags := posts.GetAllTags()

	// Should have unique tags
	expectedTags := map[string]bool{
		"go":     true,
		"web":    true,
		"python": true,
		"api":    true,
		"cli":    true,
	}

	if len(tags) != len(expectedTags) {
		t.Errorf("expected %d unique tags, got %d", len(expectedTags), len(tags))
	}

	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("unexpected tag: %q", tag)
		}
	}
}

// TestPostList_GetAllTags_Empty tests getting tags from empty list
func TestPostList_GetAllTags_Empty(t *testing.T) {
	t.Parallel()
	var posts PostList
	tags := posts.GetAllTags()

	if len(tags) != 0 {
		t.Errorf("expected 0 tags from empty list, got %d", len(tags))
	}
}

// TestPost_HTMLContent tests HTML content handling
func TestPost_HTMLContent(t *testing.T) {
	t.Parallel()
	htmlStr := "<p>Hello <strong>World</strong></p>"
	post := Post{
		Content:     htmlStr,
		HTMLContent: template.HTML(htmlStr),
	}

	if string(post.HTMLContent) != htmlStr {
		t.Errorf("HTMLContent = %q, want %q", post.HTMLContent, htmlStr)
	}

	if post.Content != htmlStr {
		t.Errorf("Content = %q, want %q", post.Content, htmlStr)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
