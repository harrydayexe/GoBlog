<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/harrydayexe/GoBlog">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">GoBlog</h3>

  <p align="center">
    Create a blog feed from posts written in Markdown!
    <br />
    <a href="https://github.com/harrydayexe/GoBlog/wiki"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://harryday.xyz/posts">View Example</a>
    &middot;
    <!-- TODO: Create bug template -->
    <a href="https://github.com/harrydayexe/GoBlog/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <!-- TODO: Create feature request template -->
    <a href="https://github.com/harrydayexe/GoBlog/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#gobloggen-usage">GoBlogGen Usage</a></li>
      </ul>
    </li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

GoBlog is a two-part system for creating and serving markdown-based blogs. Write your posts in markdown and let GoBlog handle the rest.

### GoBlogGen

**Unopinionated static site generator for markdown blogs.**

GoBlogGen converts your markdown posts into static HTML pages. You control the templates, styling, and structure. The generated output is pure HTML/CSS that can be deployed anywhere.

**Features:**
- YAML frontmatter for post metadata
- Customizable HTML templates
- Syntax highlighting for code blocks
- Tag-based organization
- Pagination support

**Perfect for:** Developers who want full control over their blog's appearance and hosting.

### GoBlogServ

**Opinionated web server with HTMX-powered interactivity.**

GoBlogServ provides two ways to add dynamic features to your blog:

1. **Go Library**: Import into your Go application for custom integration
2. **Standalone Server**: Run as a binary or Docker container for turnkey deployment

**Features:**
- Real-time search (powered by Bleve)
- Tag filtering and pagination
- HTMX for dynamic updates without page reloads
- In-memory caching (Ristretto)
- Single binary deployment
- Docker support for sidecar pattern

**Perfect for:** Developers who want interactive features without complex JavaScript frameworks.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

<!-- TODO: Add any new frameworks in here -->
* [![GoLang][GoModVer]][Go-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

### Installation

#### Option 1: Homebrew (macOS/Linux)

```bash
brew install harrydayexe/goblog/gobloggen
```

#### Option 2: Direct Download

Download the latest release for your platform from the [releases page](https://github.com/harrydayexe/GoBlog/releases).

**Linux/macOS:**
```bash
# Download and extract
tar -xzf gobloggen_*_<your-platform>.tar.gz

# Move to PATH
sudo mv gobloggen /usr/local/bin/

# Verify installation
gobloggen -help
```

**Windows:**
```powershell
# Extract the .zip file
# Add the directory to your PATH or move gobloggen.exe to a directory in PATH

# Verify installation
gobloggen.exe -help
```

#### Option 3: Go Install (For Go Developers)

```bash
go install github.com/harrydayexe/GoBlog/cmd/GoBlogGen@latest
```

### Quick Start

1. Create a directory for your blog posts:
   ```bash
   mkdir posts
   ```

2. Create a sample post with YAML frontmatter:
   ```bash
   cat > posts/hello-world.md << 'EOF'
   ---
   title: "Hello World"
   date: 2025-01-15
   description: "My first blog post"
   tags: ["intro", "blogging"]
   draft: false
   ---

   # Hello World

   This is my first blog post using GoBlog!

   ```go
   package main

   func main() {
       println("Hello, Blog!")
   }
   ```
   EOF
   ```

3. Create a configuration file:
   ```bash
   cat > config.yaml << 'EOF'
   input_folder: "./posts"
   output_folder: "./site"
   site:
     title: "My Blog"
     description: "A blog about development"
     author: "Your Name"
   EOF
   ```

4. Generate your site:
   ```bash
   gobloggen -config config.yaml
   ```

5. Your static site is now in `./site/` and ready to deploy!

### Customizing Templates

GoBlog uses HTML templates for rendering your blog. You can customize the look and feel by creating your own templates.

#### Using Default Templates

By default, GoBlog uses templates from `./templates/defaults/`. These include:
- `post.html` - Individual blog post page
- `index.html` - Blog index/list page
- `tag.html` - Tag filter page

#### Creating Custom Templates

1. Create a new directory for your templates:
   ```sh
   mkdir ./my-templates
   ```

2. Copy the default templates as a starting point:
   ```sh
   cp templates/defaults/* ./my-templates/
   ```

3. Edit the templates to match your design. Each template must exist:
   - `post.html`
   - `index.html`
   - `tag.html`

4. Update your `config.yaml`:
   ```yaml
   template_dir: "./my-templates"
   ```

#### Template Data

Templates have access to the following data:

**post.html:**
- `.Post.Title` - Post title
- `.Post.Description` - Post description
- `.Post.FormattedDate` - Human-readable date
- `.Post.ShortDate` - ISO date (YYYY-MM-DD)
- `.Post.Tags` - Array of tags
- `.Post.HTMLContent` - Rendered HTML content
- `.Site.Title` - Site title
- `.Site.Description` - Site description
- `.Site.Author` - Site author
- `.BlogPath` - Blog URL path

**index.html:**
- `.Posts` - Array of posts for this page
- `.AllTags` - Array of all tags across all posts
- `.Page` - Current page number
- `.TotalPages` - Total number of pages
- `.HasNext` - Boolean, true if there's a next page
- `.HasPrev` - Boolean, true if there's a previous page
- `.Site.*` - Site metadata (same as post.html)
- `.BlogPath` - Blog URL path

**tag.html:**
- `.Tag` - Current tag being filtered
- `.Posts` - Array of posts with this tag
- `.AllTags` - Array of all tags across all posts
- `.Site.*` - Site metadata
- `.BlogPath` - Blog URL path

#### Template Functions

The following functions are available in templates:
- `{{add .Page 1}}` - Add two numbers
- `{{sub .Page 1}}` - Subtract two numbers

See `templates/defaults/` for complete examples.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## GoBlogServ Usage

GoBlogServ adds dynamic features to your blog through two distribution models.

### Option 1: Go Library (For Developers)

Import GoBlogServ into your existing Go application:

```go
package main

import (
    "net/http"
    "github.com/harrydayexe/GoBlog/pkg/server"
)

func main() {
    // Create blog server instance
    blog := server.New(server.Config{
        ContentFolder: "./posts",
        CacheSize: 100,
    })

    // Mount to your existing router
    mux := http.NewServeMux()
    mux.Handle("/blog/", blog.Routes())

    // Or use individual handlers for custom routing
    mux.HandleFunc("GET /blog/{slug}", blog.PostHandler)
    mux.HandleFunc("GET /api/search", blog.SearchHandler)

    http.ListenAndServe(":8080", mux)
}
```

Install the library:

```bash
go get github.com/harrydayexe/GoBlog/pkg/server
```

### Option 2: Standalone Binary

Run GoBlogServ as a standalone server:

```bash
# Install via Homebrew
brew install harrydayexe/goblog/goblogserv

# Or via Go install
go install github.com/harrydayexe/GoBlog/cmd/GoBlogServ@latest

# Run with configuration file
goblogserv -config server.yaml

# Or with command-line flags
goblogserv -content ./posts -port 8080
```

### Option 3: Docker Container

Run GoBlogServ in a container:

```bash
# Pull and run
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblogserv:latest

# With custom configuration
docker run -v ./config.yaml:/config.yaml \
           -v ./posts:/posts \
           -p 8080:8080 \
           harrydayexe/goblogserv:latest -config /config.yaml
```

#### Docker Compose (Sidecar Pattern)

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"

  blog:
    image: harrydayexe/goblogserv:latest
    volumes:
      - ./posts:/posts
    environment:
      - GOBLOG_CONTENT_FOLDER=/posts
      - GOBLOG_PORT=8080
    ports:
      - "8080:8080"
```

### Configuration

Create a `server.yaml` configuration file:

```yaml
server:
  host: "localhost"
  port: 8080

content_folder: "./posts"

cache:
  max_size_mb: 100
  ttl_minutes: 60

search:
  index_path: "./blog.bleve"
  rebuild_on_start: false

blog:
  path: "/blog"
  posts_per_page: 10
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Top contributors:

<a href="https://github.com/harrydayexe/GoBlog/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=harrydayexe/GoBlog" alt="contrib.rocks image" />
</a>



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Harry Day - [@harrydayexe](https://twitter.com/harrydayexe) - contact@harryday.xyz

Project Link: [https://github.com/harrydayexe/GoBlog](https://github.com/harrydayexe/GoBlog)

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/harrydayexe/GoBlog.svg?style=for-the-badge
[contributors-url]: https://github.com/harrydayexe/GoBlog/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/harrydayexe/GoBlog.svg?style=for-the-badge
[forks-url]: https://github.com/harrydayexe/GoBlog/network/members
[stars-shield]: https://img.shields.io/github/stars/harrydayexe/GoBlog.svg?style=for-the-badge
[stars-url]: https://github.com/harrydayexe/GoBlog/stargazers
[issues-shield]: https://img.shields.io/github/issues/harrydayexe/GoBlog.svg?style=for-the-badge
[issues-url]: https://github.com/harrydayexe/GoBlog/issues
[license-shield]: https://img.shields.io/github/license/harrydayexe/GoBlog.svg?style=for-the-badge
[license-url]: https://github.com/harrydayexe/GoBlog/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/harrydayexe
[product-screenshot]: images/screenshot.png
[JQuery.com]: https://img.shields.io/badge/jQuery-0769AD?style=for-the-badge&logo=jquery&logoColor=white
[JQuery-url]: https://jquery.com
[GoModVer]: https://img.shields.io/github/go-mod/go-version/harrydayexe/GoBlog?style=for-the-badge
[Go-URL]: https://go.dev
