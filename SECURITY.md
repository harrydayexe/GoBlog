# Security Policy

## Supported Versions

**GoBlog follows a forward-only security patch policy.**

This means:
- ✅ Security fixes are **only** applied to the latest release
- ❌ **No backports** to older versions
- 📈 Users must upgrade to the latest version to receive security fixes

**Example:** If a vulnerability is discovered in v0.9.0 and the current release is v1.2.0, the fix will be released as v1.2.1. There will be no v0.9.1 or v1.1.x patch release.

| Version        | Supported          | Notes |
| -------------- | ------------------ | ----- |
| Latest release | :white_check_mark: | Security patches released as patch versions (e.g., 1.2.0 → 1.2.1) |
| Older releases | :x:                | Must upgrade to latest to receive security fixes |
| Pre-1.0.0      | :x:                | Development versions - upgrade to 1.0.0+ for production use |

**Current Status:** Pre-1.0.0 development. Only `main` branch is supported.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **contact (at) harryday (dot) xyz**

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Affected versions (if known)
- Suggested fix (if you have one)

### What to Expect

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix timeline**: Depends on severity
  - Critical: 1-7 days
  - High: 7-30 days
  - Medium/Low: Next release cycle

We will keep you informed of progress and coordinate disclosure timing with you.

## Security Best Practices

When using GoBlog in production:

### For GoBlogGen (Static Site Generator)
- Sanitize user input in markdown frontmatter if processing untrusted content
- Run in isolated environments if processing untrusted markdown
- Validate file paths to prevent path traversal attacks
- Review generated HTML before deploying to production

### For GoBlogServ (Dynamic Server)
- **Use HTTPS in production** (terminate TLS at reverse proxy like nginx/Caddy)
- **Never expose to internet without authentication** if using file watching in development mode
- Set appropriate CORS policies with `WithCORS()` - avoid wildcard origins in production
- Implement rate limiting via middleware using `WithMiddleware()`
- Run with minimal file system permissions (read-only access to content directory)
- Consider running in a container with restricted capabilities
- Enable caching to reduce DoS attack surface
- Disable file watching in production (`WatchFiles: false`)

### For Embedded Use Cases

```go
// Example: Add authentication middleware
authMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !isAuthenticated(r) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Example: Add rate limiting
rateLimitMiddleware := func(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(10, 20) // 10 req/s, burst of 20
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

blog, _ := server.New(
    server.DefaultOptions().
        WithMiddleware(authMiddleware, rateLimitMiddleware).
        WithCORS(&server.CORSConfig{
            AllowedOrigins: []string{"https://yourdomain.com"},
        }),
)
```

## Known Security Considerations

### Content Injection
- **Markdown content** is processed through [goldmark](https://github.com/yuin/goldmark) which sanitizes HTML by default
- **HTML in markdown** is escaped by default to prevent XSS
- If you enable raw HTML in goldmark configuration, be aware of XSS risks from untrusted content
- User-provided frontmatter (title, description, tags) is properly escaped in templates

### File Watching
- The file watching feature (`WithFileWatching(true)`) is intended for **development only**
- In production, disable file watching to prevent unauthorized content modification
- An attacker with write access to the content directory could inject malicious content
- Always run production servers with read-only access to the content directory

### Search Index
- The Bleve search index is stored locally at `SearchIndexPath`
- Protect the index directory with appropriate file system permissions
- Consider storing the index outside your webroot to prevent unauthorized access
- Index corruption could lead to DoS via repeated rebuild attempts

### Cache Poisoning
- The Ristretto cache stores parsed markdown in memory
- Cache keys are based on post slugs (user-controlled via filenames)
- Extremely long slugs or malformed filenames could impact cache performance
- Validate filenames if allowing untrusted users to create posts

### CORS Misconfigurations
- Default CORS is disabled (nil) - no cross-origin requests allowed
- Wildcard origins (`["*"]`) allow any domain - only use for public APIs
- Always specify exact origins for production: `AllowedOrigins: []string{"https://yourdomain.com"}`
- Avoid enabling credentials with wildcard origins (browsers reject this)

### Denial of Service
- Large markdown files (>10MB) could consume excessive memory during parsing
- Implement file size limits if accepting user-uploaded markdown
- The search index can become corrupted if disk space is exhausted
- Rate limiting is not built-in - use middleware or reverse proxy

### Dependency Vulnerabilities
- GoBlog depends on third-party packages (goldmark, bleve, ristretto, etc.)
- Run `go list -m -u all` to check for updates
- Review security advisories: `go list -json -m all | nancy sleuth`
- Keep dependencies updated, especially goldmark and bleve

## Out of Scope

The following are **not** considered security vulnerabilities:

- Denial of service via local file system exhaustion (user controls deployment)
- Theoretical attacks without proof of concept
- Issues in example code or documentation
- Vulnerabilities in dependencies (report to upstream projects)
- Social engineering attacks
- Physical access to the server

## Disclosure Policy

When a security vulnerability is fixed:

1. A patch will be released on the latest version only (forward-only policy)
2. A security advisory will be published on GitHub
3. The reporter will be credited (unless they prefer anonymity)
4. CVE will be requested for high/critical severity issues
5. Downstream users will be notified via GitHub releases and security advisories

**Coordinated Disclosure:** We request a 90-day coordinated disclosure period. If you plan to publish details of a vulnerability, please give us 90 days to release a fix before public disclosure.

## Security Hall of Fame

We appreciate security researchers who help make GoBlog safer:

<!-- This section will list contributors who responsibly disclose vulnerabilities -->

*No reports yet - be the first!*

## Contact

- **Security issues:** contact (at) harryday (dot) xyz
- **General questions:** [Open a GitHub issue](https://github.com/harrydayexe/GoBlog/issues)
- **Project maintainer:** [@harrydayexe](https://github.com/harrydayexe)

---

*This security policy was last updated: 2025-01-16*
