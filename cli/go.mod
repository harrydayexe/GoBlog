module github.com/harrydayexe/GoBlog/v2/cli

go 1.26.3

// The CLI is never published as a Go module; it is always built from source in
// this repository, so it tracks the library at the same commit.
replace github.com/harrydayexe/GoBlog/v2 => ../

require (
	github.com/fatih/color v1.18.0
	github.com/harrydayexe/GoBlog/v2 v2.0.0
	github.com/harrydayexe/GoWebUtilities v1.5.1
	github.com/urfave/cli/v3 v3.6.1
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/alecthomas/chroma/v2 v2.22.0 // indirect
	github.com/caarlos0/env/v11 v11.3.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/gorilla/feeds v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/yuin/goldmark v1.7.16 // indirect
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc // indirect
	go.abhg.dev/goldmark/frontmatter v0.3.0 // indirect
	go.abhg.dev/goldmark/wikilink v0.6.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
