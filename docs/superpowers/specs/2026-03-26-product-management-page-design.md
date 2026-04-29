# Product Management Page — Design Spec

**Date:** 2026-03-26
**Status:** Approved

## Overview

An owner-only product management page at `/admin/products`. Owners can create, edit, delete, toggle availability, and upload an image for each product. Regular users cannot see or reach this page.

## Access Control

- The Nav renders a "Manage" link only when `keycloak.hasRealmRole('owner')` is true.
- The `/admin/products` route redirects non-owners to `/` as a safety net.
- All mutating API calls already require the `owner` Keycloak realm role on the backend.

## Page Layout

A full-width table listing all products (including unavailable ones — unlike the Store which filters them out). Each row shows:

- Thumbnail (if an image exists), name, unit, description
- Availability toggle — calls `PATCH /api/v1/products/:id/availability`
- Edit button — opens the modal pre-filled with the product's current data
- Delete button — shows an inline "Are you sure?" confirmation before calling `DELETE /api/v1/products/:id`

An "Add product" button at the top of the page opens the modal in create mode.

## Modal

Single modal used for both create and edit. Fields:

| Field | Type | Required |
|-------|------|----------|
| Name | text input | yes |
| Unit | text input (e.g. "loaf", "dozen") | yes |
| Description | textarea | no |
| Available | checkbox | no (defaults to false for new products) |
| Image | file picker (image/*) | no |

**Submit flow:**
1. POST (create) or PUT (update) the product JSON fields.
2. If a file was selected, upload it separately: `POST /api/v1/products/:id/image` as multipart/form-data.
3. On success: close modal, invalidate the `['products']` TanStack Query cache, show a toast.

The two-step upload keeps the backend simple — no mixed multipart+JSON request needed.

## Frontend Changes

### `src/api/products.ts`
Add:
- `createProduct(data: { name, description, unit, available })` → `POST /api/v1/products`
- `updateProduct(id, data: { name, description, unit, available })` → `PUT /api/v1/products/:id`
- `deleteProduct(id)` → `DELETE /api/v1/products/:id`
- `setAvailability(id, available: boolean)` → `PATCH /api/v1/products/:id/availability`
- `uploadImage(id, file: File)` → `POST /api/v1/products/:id/image` (multipart)

### `src/types/index.ts`
Add `image_url?: string` to the `Product` interface.

### `src/pages/admin/Products.tsx`
New page component — product table + modal. All TanStack Query mutations invalidate `['products']` on success.

### `src/components/Nav.tsx`
Add "Manage" link to `/admin/products`, rendered only when `keycloak.hasRealmRole('owner')`.

### `src/App.tsx`
Add route: `<Route path="/admin/products" element={<AdminProducts />} />`

## Backend Changes

### Migration: `004_add_product_image.up.sql`
```sql
ALTER TABLE products ADD COLUMN image_url TEXT;
```

### Upload endpoint: `POST /api/v1/products/:id/image`
- Accepts `multipart/form-data` with a single file field (`image`).
- Validates the product exists.
- Saves the file as `./uploads/<product-uuid>.<ext>` — UUID from the product ID, extension from the uploaded file's content type. Never uses the original filename.
- Deletes any existing file for that product UUID (any extension) before saving the new one.
- Updates `products.image_url` to `/uploads/<product-uuid>.<ext>`.
- Returns the updated product.
- Owner-only (behind `RequireOwner` middleware).

### Static file serving
The router serves `./uploads/` at `/uploads/` with no auth required. URLs are opaque (UUID-based) so they are not guessable.

### `Product` domain model
Add `imageURL string` field (empty string = no image). Exposed via getter `ImageURL() string`.

### `Product` HTTP response
Add `image_url` to `productResponse` JSON struct (omitempty).

## Image Filename Convention

Images are saved as `<product-uuid>.<ext>` where:
- `<product-uuid>` is the product's existing UUID (no new UUID generated)
- `<ext>` is derived from the uploaded file's MIME type (`image/jpeg` → `.jpg`, `image/png` → `.png`, `image/webp` → `.webp`)
- The original filename from the client is never used

This eliminates collisions (one image per product) and path traversal risk.

## Out of Scope

- Image deletion (removing an image without replacing it)
- Multiple images per product
- Image resizing / CDN
- Role management UI
