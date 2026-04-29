# Product Management Page — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a full-CRUD owner-only product management page at `/admin/products` with image upload support.

**Architecture:** Backend gains an `image_url` column on products, a `POST /api/v1/products/:id/image` upload endpoint that saves files as `<uuid>.<ext>`, and static serving of `./uploads/`. Frontend gains a React page with a product table and modal form (create/edit/delete/toggle availability/upload image), guarded by Keycloak `owner` realm role.

**Tech Stack:** Go/Gin, GORM, PostgreSQL, React + TypeScript, TanStack Query, Tailwind CSS, Zustand, Keycloak JS, sonner (toasts)

---

### Task 1: DB migration — add image_url to products

**Files:**
- Create: `backend/migrations/004_add_product_image.up.sql`
- Create: `backend/migrations/004_add_product_image.down.sql`

- [ ] **Step 1: Write the up migration**

```sql
-- backend/migrations/004_add_product_image.up.sql
ALTER TABLE products ADD COLUMN image_url TEXT;
```

- [ ] **Step 2: Write the down migration**

```sql
-- backend/migrations/004_add_product_image.down.sql
ALTER TABLE products DROP COLUMN image_url;
```

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/004_add_product_image.up.sql backend/migrations/004_add_product_image.down.sql
git commit -m "feat: add image_url column to products"
```

---

### Task 2: Domain model — add imageURL field and methods, accept available in New()

**Files:**
- Modify: `backend/internal/domain/product/product.go`

- [ ] **Step 1: Update product.go**

Replace the entire file:

```go
package product

import (
	"time"

	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
)

type ID = uuid.UUID

type Product struct {
	id          ID
	name        string
	description string
	priceCents  int64
	unit        string
	available   bool
	imageURL    string
	createdAt   time.Time
	updatedAt   time.Time
}

func New(name, description string, priceCents int64, unit string, available bool) (*Product, error) {
	if name == "" {
		return nil, domerrors.BadRequest("name is required")
	}
	if priceCents < 0 {
		return nil, domerrors.BadRequest("price cannot be negative")
	}
	if unit == "" {
		return nil, domerrors.BadRequest("unit is required")
	}
	now := time.Now().UTC()
	return &Product{
		id:          uuid.New(),
		name:        name,
		description: description,
		priceCents:  priceCents,
		unit:        unit,
		available:   available,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func Reconstitute(id ID, name, description string, priceCents int64, unit string, available bool, imageURL string, createdAt, updatedAt time.Time) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		priceCents:  priceCents,
		unit:        unit,
		available:   available,
		imageURL:    imageURL,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p *Product) ID() ID               { return p.id }
func (p *Product) Name() string         { return p.name }
func (p *Product) Description() string  { return p.description }
func (p *Product) PriceCents() int64    { return p.priceCents }
func (p *Product) Unit() string         { return p.unit }
func (p *Product) Available() bool      { return p.available }
func (p *Product) ImageURL() string     { return p.imageURL }
func (p *Product) CreatedAt() time.Time { return p.createdAt }
func (p *Product) UpdatedAt() time.Time { return p.updatedAt }

func (p *Product) Update(name, description string, priceCents int64, unit string) error {
	if name == "" {
		return domerrors.BadRequest("name is required")
	}
	if priceCents < 0 {
		return domerrors.BadRequest("price cannot be negative")
	}
	if unit == "" {
		return domerrors.BadRequest("unit is required")
	}
	p.name = name
	p.description = description
	p.priceCents = priceCents
	p.unit = unit
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Product) SetAvailable(available bool) {
	p.available = available
	p.updatedAt = time.Now().UTC()
}

func (p *Product) SetImageURL(url string) {
	p.imageURL = url
	p.updatedAt = time.Now().UTC()
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/domain/product/...
```
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/product/product.go
git commit -m "feat: add imageURL field and SetImageURL to Product domain"
```

---

### Task 3: Fix domain tests for updated New() and Reconstitute() signatures

**Files:**
- Modify: `backend/internal/domain/product/product_test.go`

- [ ] **Step 1: Read the current test file**

```bash
cat backend/internal/domain/product/product_test.go
```

- [ ] **Step 2: Update every call to `New(...)` to pass `available bool` as the 5th argument**

Every `product.New(name, desc, priceCents, unit)` call → `product.New(name, desc, priceCents, unit, true)`

Every `product.Reconstitute(id, name, desc, priceCents, unit, available, createdAt, updatedAt)` call → `product.Reconstitute(id, name, desc, priceCents, unit, available, "", createdAt, updatedAt)`

- [ ] **Step 3: Add a test for SetImageURL**

At the bottom of the test file, add:

```go
func TestProduct_SetImageURL(t *testing.T) {
	p, err := New("Sourdough", "", 500, "loaf", true)
	require.NoError(t, err)
	assert.Empty(t, p.ImageURL())

	p.SetImageURL("/uploads/some-uuid.jpg")
	assert.Equal(t, "/uploads/some-uuid.jpg", p.ImageURL())
}
```

- [ ] **Step 4: Run domain tests**

```bash
cd backend && go test ./internal/domain/product/... -v
```
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/product/product_test.go
git commit -m "test: update product domain tests for new New() signature and SetImageURL"
```

---

### Task 4: Service — accept available in Create, add SetImageURL method

**Files:**
- Modify: `backend/internal/domain/product/service.go`
- Modify: `backend/internal/domain/product/service_test.go`

- [ ] **Step 1: Update service.go**

Replace `Create` and add `SetImageURL`:

```go
package product

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, description string, priceCents int64, unit string, available bool) (*Product, error) {
	p, err := New(name, description, priceCents, unit, available)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetByID(ctx context.Context, id ID) (*Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, availableOnly bool) ([]*Product, error) {
	return s.repo.List(ctx, availableOnly)
}

func (s *Service) Update(ctx context.Context, id ID, name, description string, priceCents int64, unit string) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := p.Update(name, description, priceCents, unit); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) SetAvailable(ctx context.Context, id ID, available bool) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.SetAvailable(available)
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) SetImageURL(ctx context.Context, id ID, imageURL string) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.SetImageURL(imageURL)
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}
```

- [ ] **Step 2: Update service_test.go**

Read the file:
```bash
cat backend/internal/domain/product/service_test.go
```

Update every `svc.Create(ctx, name, desc, priceCents, unit)` call → `svc.Create(ctx, name, desc, priceCents, unit, true)`

Add a test for `SetImageURL`:

```go
func TestService_SetImageURL(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := NewService(repo)
	ctx := context.Background()

	p, err := svc.Create(ctx, "Rye", "", 300, "loaf", true)
	require.NoError(t, err)

	updated, err := svc.SetImageURL(ctx, p.ID(), "/uploads/"+p.ID().String()+".jpg")
	require.NoError(t, err)
	assert.Equal(t, "/uploads/"+p.ID().String()+".jpg", updated.ImageURL())
}
```

- [ ] **Step 3: Run service tests**

```bash
cd backend && go test ./internal/domain/product/... -v
```
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/product/service.go backend/internal/domain/product/service_test.go
git commit -m "feat: add SetImageURL to product service, accept available in Create"
```

---

### Task 5: Postgres layer — add ImageURL to model and repo

**Files:**
- Modify: `backend/internal/infra/postgres/models/product.go`
- Modify: `backend/internal/infra/postgres/product_repo.go`

- [ ] **Step 1: Update models/product.go**

```go
package models

import "time"

type Product struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	Name        string    `gorm:"not null"`
	Description string
	PriceCents  int64     `gorm:"not null"`
	Unit        string    `gorm:"not null"`
	Available   bool      `gorm:"not null;default:true"`
	ImageURL    string
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Product) TableName() string { return "products" }
```

- [ ] **Step 2: Update product_repo.go — productToDomain and productToModel**

Update the two mapping functions at the bottom of `backend/internal/infra/postgres/product_repo.go`:

```go
func productToDomain(m *models.Product) (*product.Product, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	return product.Reconstitute(id, m.Name, m.Description, m.PriceCents, m.Unit, m.Available, m.ImageURL, m.CreatedAt, m.UpdatedAt), nil
}

func productToModel(p *product.Product) models.Product {
	return models.Product{
		ID:          p.ID().String(),
		Name:        p.Name(),
		Description: p.Description(),
		PriceCents:  p.PriceCents(),
		Unit:        p.Unit(),
		Available:   p.Available(),
		ImageURL:    p.ImageURL(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}
```

- [ ] **Step 3: Build to confirm no compile errors**

```bash
cd backend && go build ./...
```
Expected: no output (success)

- [ ] **Step 4: Run repo integration tests (requires Docker)**

```bash
cd backend && go test ./internal/infra/postgres/... -v -run TestProduct
```
Expected: PASS (or SKIP if Docker unavailable — `-short` will skip them)

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infra/postgres/models/product.go backend/internal/infra/postgres/product_repo.go
git commit -m "feat: add ImageURL to product postgres model and repo"
```

---

### Task 6: Handler — uploadsDir, updated response, updated Create, UploadImage

**Files:**
- Modify: `backend/internal/infra/http/handler/product.go`

- [ ] **Step 1: Replace product.go handler**

```go
package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/infra/http/respond"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	svc        *product.Service
	uploadsDir string
}

func NewProductHandler(svc *product.Service, uploadsDir string) *ProductHandler {
	return &ProductHandler{svc: svc, uploadsDir: uploadsDir}
}

type productResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Unit        string    `json:"unit"`
	Available   bool      `json:"available"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toProductResponse(p *product.Product) productResponse {
	return productResponse{
		ID:          p.ID().String(),
		Name:        p.Name(),
		Description: p.Description(),
		PriceCents:  p.PriceCents(),
		Unit:        p.Unit(),
		Available:   p.Available(),
		ImageURL:    p.ImageURL(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name"        binding:"required"`
		Description string `json:"description"`
		PriceCents  int64  `json:"price_cents" binding:"min=0"`
		Unit        string `json:"unit"        binding:"required"`
		Available   bool   `json:"available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.Create(c.Request.Context(), req.Name, req.Description, req.PriceCents, req.Unit, req.Available)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusCreated, toProductResponse(p))
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func (h *ProductHandler) List(c *gin.Context) {
	availableOnly := c.Query("available") == "true"
	products, err := h.svc.List(c.Request.Context(), availableOnly)
	if err != nil {
		respond.Error(c, err)
		return
	}
	resp := make([]productResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, toProductResponse(p))
	}
	respond.JSON(c, http.StatusOK, resp)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	var req struct {
		Name        string `json:"name"        binding:"required"`
		Description string `json:"description"`
		PriceCents  int64  `json:"price_cents" binding:"min=0"`
		Unit        string `json:"unit"        binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, req.Name, req.Description, req.PriceCents, req.Unit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func (h *ProductHandler) SetAvailable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	var req struct {
		Available bool `json:"available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.SetAvailable(c.Request.Context(), id, req.Available)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadImage accepts a multipart upload, saves the file as <product-uuid>.<ext>
// in h.uploadsDir, and updates the product's image_url.
func (h *ProductHandler) UploadImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}

	// Confirm product exists before touching the filesystem.
	if _, err := h.svc.GetByID(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		respond.Error(c, domerrors.BadRequest("image file is required"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	ext, err := imageExt(contentType)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("unsupported image type: %s", contentType))
		return
	}

	// Remove any existing image for this product (any extension).
	existing, _ := filepath.Glob(filepath.Join(h.uploadsDir, id.String()+".*"))
	for _, f := range existing {
		os.Remove(f) //nolint:errcheck
	}

	filename := id.String() + ext
	dst := filepath.Join(h.uploadsDir, filename)

	data, err := io.ReadAll(file)
	if err != nil {
		respond.Error(c, fmt.Errorf("reading upload: %w", err))
		return
	}
	if err := os.WriteFile(dst, data, 0644); err != nil { //nolint:gosec
		respond.Error(c, fmt.Errorf("saving upload: %w", err))
		return
	}

	p, err := h.svc.SetImageURL(c.Request.Context(), id, "/uploads/"+filename)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func imageExt(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	case "image/gif":
		return ".gif", nil
	default:
		return "", fmt.Errorf("unsupported image content-type: %s", contentType)
	}
}
```

- [ ] **Step 2: Build**

```bash
cd backend && go build ./...
```
Expected: compile error in `main.go` (NewProductHandler signature changed) — fix in Task 8.

- [ ] **Step 3: Commit handler changes (pre-fix)**

```bash
git add backend/internal/infra/http/handler/product.go
git commit -m "feat: add UploadImage handler, uploadsDir config, image_url in response"
```

---

### Task 7: Handler tests — fix existing, add UploadImage tests

**Files:**
- Modify: `backend/internal/infra/http/handler/product_test.go`

- [ ] **Step 1: Update setupProductRouter to pass uploadsDir**

Change:
```go
func setupProductRouter(repo *mock.ProductRepo) *gin.Engine {
	svc := product.NewService(repo)
	h := handler.NewProductHandler(svc)
```
To:
```go
func setupProductRouter(repo *mock.ProductRepo) (*gin.Engine, string) {
	svc := product.NewService(repo)
	uploadsDir := t.TempDir() // Note: must thread t into this function
	h := handler.NewProductHandler(svc, uploadsDir)
```

Wait — `setupProductRouter` currently doesn't take `t *testing.T`. Update signature:

```go
func setupProductRouter(t *testing.T, repo *mock.ProductRepo) (*gin.Engine, string) {
	t.Helper()
	svc := product.NewService(repo)
	uploadsDir := t.TempDir()
	h := handler.NewProductHandler(svc, uploadsDir)

	r := gin.New()
	r.POST("/products", h.Create)
	r.GET("/products", h.List)
	r.GET("/products/:id", h.GetByID)
	r.PUT("/products/:id", h.Update)
	r.PATCH("/products/:id/availability", h.SetAvailable)
	r.DELETE("/products/:id", h.Delete)
	r.POST("/products/:id/image", h.UploadImage)

	return r, uploadsDir
}
```

- [ ] **Step 2: Update registerProduct to match new setupProductRouter**

```go
func registerProduct(t *testing.T, router *gin.Engine, name string) map[string]any {
	t.Helper()
	w := postJSON(router, "/products", map[string]any{
		"name":        name,
		"description": "test product",
		"price_cents": 450,
		"unit":        "loaf",
		"available":   true,
	})
	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}
```

- [ ] **Step 3: Update all test functions that call setupProductRouter**

Every `router := setupProductRouter(mock.NewProductRepo())` → `router, _ := setupProductRouter(t, mock.NewProductRepo())`

Every `repo := mock.NewProductRepo(); router := setupProductRouter(repo)` → `repo := mock.NewProductRepo(); router, _ := setupProductRouter(t, repo)`

- [ ] **Step 4: Add UploadImage tests at the bottom of the file**

```go
// ---------------------------------------------------------------------------
// UploadImage
// ---------------------------------------------------------------------------

func multipartImageRequest(router *gin.Engine, path string, filename, contentType string, content []byte) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", filename)
	part.Write(content)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestProductHandler_UploadImage_Success(t *testing.T) {
	router, uploadsDir := setupProductRouter(t, mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")
	id := created["id"].(string)

	w := multipartImageRequest(router, "/products/"+id+"/image", "photo.jpg", "image/jpeg", []byte("fakejpeg"))
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "/uploads/"+id+".jpg", resp["image_url"])

	// File should exist on disk.
	_, err := os.Stat(filepath.Join(uploadsDir, id+".jpg"))
	assert.NoError(t, err)
}

func TestProductHandler_UploadImage_ReplacesOldFile(t *testing.T) {
	router, uploadsDir := setupProductRouter(t, mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")
	id := created["id"].(string)

	// Upload a PNG first.
	multipartImageRequest(router, "/products/"+id+"/image", "photo.png", "image/png", []byte("fakepng"))

	// Then upload a JPEG — the PNG should be removed.
	multipartImageRequest(router, "/products/"+id+"/image", "photo.jpg", "image/jpeg", []byte("fakejpeg"))

	_, pngErr := os.Stat(filepath.Join(uploadsDir, id+".png"))
	assert.True(t, os.IsNotExist(pngErr), "old PNG should have been deleted")

	_, jpgErr := os.Stat(filepath.Join(uploadsDir, id+".jpg"))
	assert.NoError(t, jpgErr, "new JPG should exist")
}

func TestProductHandler_UploadImage_InvalidUUID(t *testing.T) {
	router, _ := setupProductRouter(t, mock.NewProductRepo())

	w := multipartImageRequest(router, "/products/bad-id/image", "photo.jpg", "image/jpeg", []byte("x"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_UploadImage_ProductNotFound(t *testing.T) {
	router, _ := setupProductRouter(t, mock.NewProductRepo())

	w := multipartImageRequest(router, "/products/"+uuid.New().String()+"/image", "photo.jpg", "image/jpeg", []byte("x"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProductHandler_UploadImage_UnsupportedType(t *testing.T) {
	router, _ := setupProductRouter(t, mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")

	w := multipartImageRequest(router, "/products/"+created["id"].(string)+"/image", "file.pdf", "application/pdf", []byte("x"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_UploadImage_NoFile(t *testing.T) {
	router, _ := setupProductRouter(t, mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")

	req := httptest.NewRequest(http.MethodPost, "/products/"+created["id"].(string)+"/image", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

- [ ] **Step 5: Add missing imports**

The test file will need `"mime/multipart"`, `"os"`, and `"path/filepath"`. Add them to the import block at the top of `product_test.go`.

- [ ] **Step 6: Run handler tests**

```bash
cd backend && go test ./internal/infra/http/handler/... -v -run TestProduct
```
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/infra/http/handler/product_test.go
git commit -m "test: add UploadImage handler tests, update setupProductRouter"
```

---

### Task 8: Router, main.go, and uploads directory

**Files:**
- Modify: `backend/internal/infra/http/router.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `.gitignore`

- [ ] **Step 1: Update router.go — add upload route and static serving**

```go
package http

import (
	"fmt"
	"net/http"

	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	users *handler.UserHandler,
	products *handler.ProductHandler,
	orders *handler.OrderHandler,
	keycloakURL string,
	keycloakRealm string,
	uploadsDir string,
) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Serve uploaded images without auth — URLs are UUID-based and not guessable.
	r.Static("/uploads", uploadsDir)

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, keycloakRealm)
	ownerOnly := middleware.RequireOwner()

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(jwksURL))
	{
		u := v1.Group("/users")
		u.GET("", users.List)
		u.POST("", users.Register)
		u.GET("/:id", users.GetByID)
		u.PUT("/:id", users.Update)

		p := v1.Group("/products")
		p.GET("", products.List)
		p.GET("/:id", products.GetByID)
		p.POST("", ownerOnly, products.Create)
		p.PUT("/:id", ownerOnly, products.Update)
		p.PATCH("/:id/availability", ownerOnly, products.SetAvailable)
		p.DELETE("/:id", ownerOnly, products.Delete)
		p.POST("/:id/image", ownerOnly, products.UploadImage)

		o := v1.Group("/orders")
		o.GET("", orders.List)
		o.POST("", orders.Create)
		o.GET("/:id", orders.GetByID)
		o.POST("/:id/items", orders.AddItem)
		o.DELETE("/:id/items/:productID", orders.RemoveItem)
		o.POST("/:id/confirm", orders.Confirm)
		o.POST("/:id/fulfill", orders.Fulfill)
		o.POST("/:id/cancel", orders.Cancel)
	}

	return r
}
```

- [ ] **Step 2: Update main.go — pass uploadsDir, create the directory**

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fhardow/foodo/internal/config"
	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/domain/user"
	"github.com/fhardow/foodo/internal/infra/postgres"
	apphttp "github.com/fhardow/foodo/internal/infra/http"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/pkg/logger"
)

const uploadsDir = "./uploads"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Env)

	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Error("failed to create uploads directory", "err", err)
		os.Exit(1)
	}

	db, err := postgres.Connect(cfg.DSN, cfg.Env)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}

	if err := postgres.Migrate(db); err != nil {
		log.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	// Repositories
	userRepo    := postgres.NewUserRepo(db)
	productRepo := postgres.NewProductRepo(db)
	orderRepo   := postgres.NewOrderRepo(db)

	// Domain services
	userSvc    := user.NewService(userRepo)
	productSvc := product.NewService(productRepo)
	orderSvc   := order.NewService(orderRepo, productRepo)

	// HTTP handlers
	userHandler    := handler.NewUserHandler(userSvc)
	productHandler := handler.NewProductHandler(productSvc, uploadsDir)
	orderHandler   := handler.NewOrderHandler(orderSvc)

	router := apphttp.NewRouter(userHandler, productHandler, orderHandler, cfg.KeycloakURL, cfg.KeycloakRealm, uploadsDir)
	srv    := apphttp.NewServer(cfg.Port, router, log)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "err", err)
	}
}
```

- [ ] **Step 3: Add uploads dir to .gitignore**

Open `.gitignore` at the repo root and add:
```
backend/uploads/
```

- [ ] **Step 4: Full build and test**

```bash
cd backend && go build ./... && go test ./... -short
```
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infra/http/router.go backend/cmd/api/main.go .gitignore
git commit -m "feat: add upload route, static file serving, ensure uploads dir on startup"
```

---

### Task 9: Frontend — types and API

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/api/products.ts`

- [ ] **Step 1: Update types/index.ts — add image_url to Product**

```typescript
export type UUID = string

export interface Product {
  id: UUID
  name: string
  description: string
  unit: string
  available: boolean
  image_url?: string
}

export interface OrderItem {
  product_id: UUID
  product_name: string
  quantity: number
}

export interface Order {
  id: UUID
  status: 'pending' | 'confirmed' | 'fulfilled' | 'cancelled'
  items: OrderItem[]
  created_at: string
}
```

- [ ] **Step 2: Update api/products.ts — add all mutation functions**

```typescript
import keycloak from '../auth/keycloak'
import { apiFetch } from './client'
import type { Product } from '../types'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const getProducts = () =>
  apiFetch<Product[]>('/api/v1/products')

export interface ProductInput {
  name: string
  description: string
  unit: string
  available: boolean
  price_cents?: number
}

export const createProduct = (data: ProductInput) =>
  apiFetch<Product>('/api/v1/products', {
    method: 'POST',
    body: JSON.stringify({ price_cents: 0, ...data }),
  })

export const updateProduct = (id: string, data: ProductInput) =>
  apiFetch<Product>(`/api/v1/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ price_cents: 0, ...data }),
  })

export const deleteProduct = (id: string) =>
  apiFetch<void>(`/api/v1/products/${id}`, { method: 'DELETE' })

export const setAvailability = (id: string, available: boolean) =>
  apiFetch<Product>(`/api/v1/products/${id}/availability`, {
    method: 'PATCH',
    body: JSON.stringify({ available }),
  })

// uploadImage uses a raw fetch because apiFetch forces Content-Type: application/json,
// which breaks multipart/form-data uploads.
export const uploadImage = async (id: string, file: File): Promise<Product> => {
  await keycloak.updateToken(30).catch(() => keycloak.login())
  const form = new FormData()
  form.append('image', file)
  const res = await fetch(`${BASE_URL}/api/v1/products/${id}/image`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${keycloak.token}` },
    body: form,
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json() as Promise<Product>
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/api/products.ts
git commit -m "feat: add product mutation API functions and image_url type"
```

---

### Task 10: Frontend — Admin Products page

**Files:**
- Create: `frontend/src/pages/admin/Products.tsx`

- [ ] **Step 1: Create the page**

```tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import keycloak from '../../auth/keycloak'
import { getProducts, createProduct, updateProduct, deleteProduct, setAvailability, uploadImage } from '../../api/products'
import type { ProductInput } from '../../api/products'
import type { Product } from '../../types'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

const emptyForm: ProductInput = { name: '', description: '', unit: '', available: false }

export default function AdminProducts() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  if (!keycloak.hasRealmRole('owner')) {
    navigate('/')
    return null
  }

  const { data: products = [], isLoading, isError } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const [form, setForm] = useState<ProductInput>(emptyForm)
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['products'] })

  const availabilityMutation = useMutation({
    mutationFn: ({ id, available }: { id: string; available: boolean }) =>
      setAvailability(id, available),
    onSuccess: invalidate,
    onError: () => toast.error('Failed to update availability'),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteProduct,
    onSuccess: () => {
      invalidate()
      setDeleteConfirm(null)
      toast.success('Product deleted')
    },
    onError: () => toast.error('Failed to delete product'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setImageFile(null)
    setModalOpen(true)
  }

  const openEdit = (p: Product) => {
    setEditing(p)
    setForm({ name: p.name, description: p.description, unit: p.unit, available: p.available })
    setImageFile(null)
    setModalOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      let product: Product
      if (editing) {
        product = await updateProduct(editing.id, form)
        toast.success('Product updated')
      } else {
        product = await createProduct(form)
        toast.success('Product created')
      }
      if (imageFile) {
        await uploadImage(product.id, imageFile)
      }
      invalidate()
      setModalOpen(false)
    } catch {
      toast.error('Something went wrong')
    } finally {
      setSubmitting(false)
    }
  }

  if (isLoading) {
    return <div className="text-center py-16 text-[#8a6a50]">Loading…</div>
  }

  if (isError) {
    return <div className="text-center py-16 text-[#8a6a50]">Failed to load products.</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-[#5c3d1e]">Products</h1>
        <button
          onClick={openCreate}
          className="bg-[#5c3d1e] text-white rounded px-4 py-2 text-sm hover:bg-[#3d2b1a] transition-colors"
        >
          + Add product
        </button>
      </div>

      {products.length === 0 ? (
        <p className="text-[#8a6a50] text-center py-16">No products yet.</p>
      ) : (
        <div className="bg-white rounded-lg border border-[#e8ddd0] overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[#e8ddd0] text-left text-[#8a6a50]">
                <th className="px-4 py-3 w-12"></th>
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3 hidden sm:table-cell">Unit</th>
                <th className="px-4 py-3 hidden md:table-cell">Description</th>
                <th className="px-4 py-3">Available</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id} className="border-b border-[#e8ddd0] last:border-0">
                  <td className="px-4 py-3">
                    {p.image_url ? (
                      <img
                        src={`${BASE_URL}${p.image_url}`}
                        alt={p.name}
                        className="w-10 h-10 object-cover rounded"
                      />
                    ) : (
                      <div className="w-10 h-10 bg-[#f0e8de] rounded flex items-center justify-center text-[#c4a882] text-xs">
                        No img
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3 font-medium text-[#3d2b1a]">{p.name}</td>
                  <td className="px-4 py-3 hidden sm:table-cell text-[#8a6a50]">{p.unit}</td>
                  <td className="px-4 py-3 hidden md:table-cell text-[#8a6a50] max-w-xs truncate">
                    {p.description}
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => availabilityMutation.mutate({ id: p.id, available: !p.available })}
                      className={`w-10 h-6 rounded-full transition-colors ${
                        p.available ? 'bg-[#5c3d1e]' : 'bg-[#e8ddd0]'
                      }`}
                      aria-label={p.available ? 'Mark unavailable' : 'Mark available'}
                    >
                      <span
                        className={`block w-4 h-4 bg-white rounded-full mx-1 transition-transform ${
                          p.available ? 'translate-x-4' : 'translate-x-0'
                        }`}
                      />
                    </button>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 justify-end">
                      {deleteConfirm === p.id ? (
                        <>
                          <span className="text-[#8a6a50] text-xs">Delete?</span>
                          <button
                            onClick={() => deleteMutation.mutate(p.id)}
                            className="text-red-600 text-xs font-medium hover:underline"
                          >
                            Yes
                          </button>
                          <button
                            onClick={() => setDeleteConfirm(null)}
                            className="text-[#8a6a50] text-xs hover:underline"
                          >
                            No
                          </button>
                        </>
                      ) : (
                        <>
                          <button
                            onClick={() => openEdit(p)}
                            className="text-[#5c3d1e] text-xs font-medium hover:underline"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => setDeleteConfirm(p.id)}
                            className="text-red-500 text-xs font-medium hover:underline"
                          >
                            Delete
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Modal */}
      {modalOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md">
            <div className="px-6 py-4 border-b border-[#e8ddd0] flex justify-between items-center">
              <h2 className="font-semibold text-[#3d2b1a]">
                {editing ? 'Edit product' : 'Add product'}
              </h2>
              <button
                onClick={() => setModalOpen(false)}
                className="text-[#8a6a50] hover:text-[#3d2b1a] text-xl leading-none"
              >
                ×
              </button>
            </div>
            <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">
                  Name <span className="text-red-400">*</span>
                </label>
                <input
                  required
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="w-full border border-[#e8ddd0] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#5c3d1e]"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">
                  Unit <span className="text-red-400">*</span>
                </label>
                <input
                  required
                  placeholder="e.g. loaf, dozen"
                  value={form.unit}
                  onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}
                  className="w-full border border-[#e8ddd0] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#5c3d1e]"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">Description</label>
                <textarea
                  rows={3}
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  className="w-full border border-[#e8ddd0] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#5c3d1e] resize-none"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  id="available"
                  type="checkbox"
                  checked={form.available}
                  onChange={(e) => setForm((f) => ({ ...f, available: e.target.checked }))}
                  className="accent-[#5c3d1e]"
                />
                <label htmlFor="available" className="text-sm text-[#3d2b1a]">
                  Available for ordering
                </label>
              </div>
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">Image</label>
                {editing?.image_url && !imageFile && (
                  <img
                    src={`${BASE_URL}${editing.image_url}`}
                    alt="current"
                    className="w-16 h-16 object-cover rounded mb-2"
                  />
                )}
                <input
                  type="file"
                  accept="image/*"
                  onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
                  className="text-sm text-[#8a6a50]"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setModalOpen(false)}
                  className="px-4 py-2 text-sm text-[#8a6a50] hover:text-[#3d2b1a]"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="bg-[#5c3d1e] text-white rounded px-4 py-2 text-sm hover:bg-[#3d2b1a] disabled:opacity-50 transition-colors"
                >
                  {submitting ? 'Saving…' : editing ? 'Save changes' : 'Create product'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/admin/Products.tsx
git commit -m "feat: add AdminProducts page with CRUD table and modal form"
```

---

### Task 11: Frontend — Nav link and App routing

**Files:**
- Modify: `frontend/src/components/Nav.tsx`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Update Nav.tsx — add Manage link for owners**

```tsx
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import { useBasketStore } from '../store/basket'
import keycloak from '../auth/keycloak'

export default function Nav() {
  const basketOrderId = useBasketStore((s) => s.basketOrderId)
  const { data: order } = useQuery({
    queryKey: ['order', basketOrderId],
    queryFn: () => getOrder(basketOrderId!),
    enabled: !!basketOrderId,
  })

  const itemCount = order?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0
  const isOwner = keycloak.hasRealmRole('owner')

  return (
    <nav className="bg-white border-b border-[#e8ddd0] px-4 py-3 flex justify-between items-center sticky top-0 z-10">
      <Link to="/" className="font-bold text-[#5c3d1e] text-lg">
        Foodo
      </Link>
      <div className="flex items-center gap-4">
        <Link to="/" className="hidden sm:block text-sm text-[#5c3d1e] hover:text-[#3d2b1a]">
          Store
        </Link>
        <Link to="/orders" className="hidden sm:block text-sm text-[#5c3d1e] hover:text-[#3d2b1a]">
          History
        </Link>
        {isOwner && (
          <Link to="/admin/products" className="hidden sm:block text-sm text-[#5c3d1e] hover:text-[#3d2b1a]">
            Manage
          </Link>
        )}
        <Link
          to="/basket"
          aria-label={`Basket${itemCount > 0 ? `, ${itemCount} item${itemCount !== 1 ? 's' : ''}` : ''}`}
          className="bg-[#5c3d1e] text-white rounded-full px-3 py-1 text-sm hover:bg-[#3d2b1a] transition-colors"
        >
          🛒{itemCount > 0 ? ` ${itemCount}` : ''}
        </Link>
      </div>
    </nav>
  )
}
```

- [ ] **Step 2: Update App.tsx — add the admin route**

```tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import Nav from './components/Nav'
import Store from './pages/Store'
import Basket from './pages/Basket'
import OrderStatus from './pages/OrderStatus'
import OrderHistory from './pages/OrderHistory'
import AdminProducts from './pages/admin/Products'

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen bg-[#faf7f2]">
          <Nav />
          <main className="max-w-5xl mx-auto px-4 py-6">
            <Routes>
              <Route path="/" element={<Store />} />
              <Route path="/basket" element={<Basket />} />
              <Route path="/orders/:id" element={<OrderStatus />} />
              <Route path="/orders" element={<OrderHistory />} />
              <Route path="/admin/products" element={<AdminProducts />} />
            </Routes>
          </main>
        </div>
        <Toaster richColors />
      </BrowserRouter>
    </QueryClientProvider>
  )
}
```

- [ ] **Step 3: Build frontend to check for type errors**

```bash
cd frontend && npm run build
```
Expected: Build succeeds with no TypeScript errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/Nav.tsx frontend/src/App.tsx
git commit -m "feat: add Manage nav link for owners, register /admin/products route"
```

---

## Done

End state:
- `GET /uploads/:file` serves product images (no auth)
- `POST /api/v1/products/:id/image` (owner only) saves `<uuid>.<ext>`, removes old file
- All existing product endpoints now include `image_url` in responses
- `/admin/products` is a full-CRUD product table visible only to Keycloak `owner` role holders
- "Manage" nav link appears only for owners
