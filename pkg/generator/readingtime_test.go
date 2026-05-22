// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"testing"
)

func TestWordCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "plain prose",
			input: "one two three",
			want:  3,
		},
		{
			name:  "HTML tags stripped",
			input: "<p>hello <em>world</em></p>",
			want:  2,
		},
		{
			name:  "empty input",
			input: "",
			want:  0,
		},
		{
			name:  "adjacent tags do not glue words",
			input: "<p>a</p><p>b</p>",
			want:  2,
		},
		{
			name:  "code blocks counted",
			input: "<pre><code>foo bar</code></pre>",
			want:  2,
		},
		{
			name:  "HTML entity inside word counted once",
			input: "it&#39;s",
			want:  1,
		},
		{
			name:  "only whitespace",
			input: "   \t\n  ",
			want:  0,
		},
		{
			name:  "mixed content",
			input: "<h1>Title</h1><p>Some <strong>bold</strong> text.</p>",
			want:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wordCount([]byte(tt.input))
			if got != tt.want {
				t.Errorf("wordCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestMinutesFromWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		words int
		want  int
	}{
		{0, 1},
		{1, 1},
		{-1, 1},
		{219, 1},
		{220, 1},
		{221, 2},
		{440, 2},
		{441, 3},
		{1100, 5},
	}

	for _, tt := range tests {
		got := minutesFromWords(tt.words)
		if got != tt.want {
			t.Errorf("minutesFromWords(%d) = %d, want %d", tt.words, got, tt.want)
		}
	}
}
