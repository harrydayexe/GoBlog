---
title: Building REST APIs with Go
date: 2026-01-08T16:45:00Z
description: A comprehensive guide to building RESTful APIs using Go's standard library and popular frameworks
tags:
  - go
  - api
  - web
  - backend
---

# Building REST APIs with Go

Go's standard library provides excellent tools for building HTTP servers and REST APIs. Let's explore how to create a production-ready API.

## Using the Standard Library

Go's `net/http` package is powerful enough for many use cases:

```go
package main

import (
    "encoding/json"
    "net/http"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func getUsers(w http.ResponseWriter, r *http.Request) {
    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func main() {
    http.HandleFunc("/api/users", getUsers)
    http.ListenAndServe(":8080", nil)
}
```

## Popular Frameworks

While the standard library is great, frameworks can speed up development:

### Chi Router

```go
import "github.com/go-chi/chi/v5"

func main() {
    r := chi.NewRouter()
    r.Get("/api/users", getUsers)
    r.Post("/api/users", createUser)
    http.ListenAndServe(":8080", r)
}
```

### Gin

```go
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/api/users", getUsers)
    r.POST("/api/users", createUser)
    r.Run(":8080")
}
```

## Request Validation

Always validate incoming data:

```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=2"`
    Email string `json:"email" binding:"required,email"`
}

func createUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Process valid request
    c.JSON(http.StatusCreated, gin.H{"status": "created"})
}
```

## Error Handling

Consistent error responses are crucial:

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}

func handleError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error:   http.StatusText(status),
        Message: message,
    })
}
```

## Middleware

Middleware is essential for cross-cutting concerns:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

## Testing

Always test your API endpoints:

```go
func TestGetUsers(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/users", nil)
    w := httptest.NewRecorder()

    getUsers(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

Building APIs in Go is straightforward and the ecosystem provides excellent tools for any scale of project!
