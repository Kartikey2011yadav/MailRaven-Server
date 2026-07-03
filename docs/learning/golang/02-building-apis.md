# Building REST APIs in Go (As Done in MailRaven)

---

## 1. HTTP Server Basics

Go's `net/http` is production-ready out of the box. MailRaven uses **chi** (a lightweight router) on top of it.

```go
// internal/adapters/http/server.go
import "github.com/go-chi/chi/v5"

router := chi.NewRouter()

// Register routes
router.Get("/api/v1/messages", messageHandler.ListMessages)
router.Post("/api/v1/auth/login", authHandler.Login)
router.Patch("/api/v1/messages/{id}", messageHandler.UpdateMessage)

// Start server
http.ListenAndServe(":8080", router)
```

**Chi URL params**: `{id}` in the route → `chi.URLParam(r, "id")` in the handler.

---

## 2. Handler Pattern

Every HTTP handler has this signature:

```go
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
    // 1. Parse input
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))

    // 2. Call service/repository
    messages, err := h.emailRepo.List(r.Context(), page, 20)
    if err != nil {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    // 3. Write response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}
```

**`w http.ResponseWriter`** — Write the response (status code, headers, body).
**`r *http.Request`** — Read the request (method, URL, headers, body, context).

---

## 3. Middleware

Middleware wraps handlers to add cross-cutting behavior:

```go
// internal/adapters/http/middleware/auth.go
func Auth(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Extract token from Authorization header
            token := r.Header.Get("Authorization")

            // 2. Validate
            claims, err := validateJWT(token, jwtSecret)
            if err != nil {
                http.Error(w, "Unauthorized", 401)
                return  // Stop here, don't call next
            }

            // 3. Add user info to context
            ctx := context.WithValue(r.Context(), "user_email", claims.Email)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usage:
router.Group(func(r chi.Router) {
    r.Use(middleware.Auth(jwtSecret))  // All routes below require auth
    r.Get("/api/v1/messages", handler.ListMessages)
})
```

**MailRaven middleware stack** (applied in order):
1. `Logging` — Log every request
2. `CORS` — Allow cross-origin requests
3. `MaxBodySize` — Limit request body to 10MB
4. `Compression` — Gzip responses
5. `RateLimit` — 100 req/min per IP
6. `Auth` — JWT validation (on protected routes)

---

## 4. JSON Request/Response

```go
// Parse JSON body (internal/adapters/http/handlers/auth.go)
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate
    if req.Email == "" || req.Password == "" {
        http.Error(w, "Email and password required", http.StatusBadRequest)
        return
    }

    // ... authenticate ...

    // Respond
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "token": token,
    })
}
```

---

## 5. Route Groups and Protection

```go
// Public routes (no auth)
router.Post("/api/v1/auth/login", authHandler.Login)
router.Get("/health", healthCheck)

// Protected routes (JWT required)
router.Group(func(r chi.Router) {
    r.Use(middleware.Auth(jwtSecret))

    r.Get("/api/v1/messages", messageHandler.List)
    r.Get("/api/v1/messages/{id}", messageHandler.Get)
    r.Post("/api/v1/messages/send", messageHandler.Send)
})

// Admin routes (JWT + admin role required)
router.Group(func(r chi.Router) {
    r.Use(middleware.Auth(jwtSecret))
    r.Use(middleware.RequireAdmin)

    r.Get("/api/v1/admin/users", adminHandler.ListUsers)
    r.Post("/api/v1/admin/users", adminHandler.CreateUser)
})
```

---

## 6. Graceful Shutdown

```go
// cmd/mailraven/serve.go
server := &http.Server{Addr: ":8080", Handler: router}

// Start in background
go server.ListenAndServe()

// Wait for SIGTERM
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan

// Graceful stop (finish in-flight requests, 30s timeout)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

---

## 7. JWT Authentication

```go
import "github.com/golang-jwt/jwt/v5"

// Generate token (on login)
claims := jwt.MapClaims{
    "email": user.Email,
    "role":  user.Role,
    "exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte(jwtSecret))

// Validate token (in middleware)
token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
    return []byte(jwtSecret), nil
})
claims := token.Claims.(jwt.MapClaims)
email := claims["email"].(string)
```
