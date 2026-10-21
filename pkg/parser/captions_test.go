// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"strings"
	"testing"
)

func TestParseFile_ImageCaptions(t *testing.T) {
	t.Parallel()

	const captioned = "<figure>\n" +
		`<img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" />` +
		"<figcaption><p>The <em>pipeline</em>.</p></figcaption>\n" +
		"</figure>\n"

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "markdown image with caption",
			body: "![A diagram](a.png)\nThe *pipeline*.\n",
			want: captioned,
		},
		{
			name: "wikilink embed with caption",
			body: "![[a.png|A diagram]]\nThe *pipeline*.\n",
			want: captioned,
		},
		{
			name: "multi-line caption",
			body: "![A diagram](a.png)\nFirst line\nsecond line\n",
			want: "<figcaption><p>First line<br />\nsecond line</p></figcaption>",
		},
		{
			name: "markdown image without caption is not a figure",
			body: "![A diagram](a.png)\n",
			want: `<p><img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" /></p>`,
		},
		{
			name: "wikilink embed without caption is not a figure",
			body: "![[a.png|A diagram]]\n",
			want: `<p><img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" /></p>`,
		},
		{
			name: "inline image is not a figure",
			body: "Before ![A diagram](a.png) after\nmore text\n",
			want: `<p>Before <img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" /> after<br />` + "\nmore text</p>",
		},
		{
			name: "wikilink embed followed by text on its line is not a figure",
			body: "![[a.png|A diagram]] after\nmore text\n",
			want: `<p><img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" /> after<br />` + "\nmore text</p>",
		},
		{
			name: "wikilink embed after text is not a figure",
			body: "Before\n![[a.png|A diagram]]\nafter\n",
			want: "<p>Before<br />\n" + `<img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" /><br />` + "\nafter</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseBodyWithOptions(t, tt.body)
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.want, got)
			}
		})
	}
}

func TestParseFile_ImageCaptionsDisabled(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		"![A diagram](a.png)\nThe caption\n",
		"![[a.png|A diagram]]\nThe caption\n",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			got := parseBodyWithOptions(t, body, WithDisableImageCaptions())
			want := `<p><img src="/images/a.png" alt="A diagram" loading="lazy" decoding="async" /><br />` + "\nThe caption</p>"
			if !strings.Contains(got, want) {
				t.Errorf("expected output to contain %q, got:\n%s", want, got)
			}
			if strings.Contains(got, "<figure>") {
				t.Errorf("expected no <figure> with captions disabled, got:\n%s", got)
			}
		})
	}
}
