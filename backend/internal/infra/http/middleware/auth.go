package middleware

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"
const RolesKey = "roles"

// jwksCache caches RSA public keys fetched from a JWKS endpoint.
type jwksCache struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	ttl       time.Duration
	url       string
}

type rawJWK struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type rawJWKS struct {
	Keys []rawJWK `json:"keys"`
}

func newJWKSCache(jwksURL string) *jwksCache {
	return &jwksCache{
		url:  jwksURL,
		ttl:  10 * time.Minute,
		keys: make(map[string]*rsa.PublicKey),
	}
}

func (c *jwksCache) getKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	fresh := time.Since(c.fetchedAt) < c.ttl
	key, found := c.keys[kid]
	c.mu.RUnlock()

	if fresh && found {
		return key, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Re-check after acquiring write lock.
	if time.Since(c.fetchedAt) < c.ttl {
		if key, ok := c.keys[kid]; ok {
			return key, nil
		}
	}

	if err := c.fetch(); err != nil {
		return nil, fmt.Errorf("JWKS fetch: %w", err)
	}

	key, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("kid %q not found in JWKS", kid)
	}
	return key, nil
}

func (c *jwksCache) fetch() error {
	resp, err := http.Get(c.url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var set rawJWKS
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := parseJWKRSA(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}

	c.keys = keys
	c.fetchedAt = time.Now()
	return nil
}

func parseJWKRSA(k rawJWK) (*rsa.PublicKey, error) {
	if len(k.X5c) > 0 {
		der, err := base64.StdEncoding.DecodeString(k.X5c[0])
		if err != nil {
			return nil, err
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, err
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("x5c cert is not RSA")
		}
		return pub, nil
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}

// JWTAuth validates Keycloak-issued RS256 JWT bearer tokens on every request.
func JWTAuth(jwksURL string) gin.HandlerFunc {
	cache := newJWKSCache(jwksURL)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		sub, roles, err := parseAndValidate(tokenStr, cache)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(UserIDKey, sub)
		c.Set(RolesKey, roles)
		c.Next()
	}
}

type tokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

type tokenClaims struct {
	Sub         string      `json:"sub"`
	Exp         int64       `json:"exp"`
	RealmAccess realmAccess `json:"realm_access"`
}

func parseAndValidate(raw string, cache *jwksCache) (string, []string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return "", nil, fmt.Errorf("malformed JWT")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", nil, err
	}
	var hdr tokenHeader
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		return "", nil, err
	}
	if hdr.Alg != "RS256" {
		return "", nil, fmt.Errorf("unsupported alg: %s", hdr.Alg)
	}

	pub, err := cache.getKey(hdr.Kid)
	if err != nil {
		return "", nil, err
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", nil, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sigBytes); err != nil {
		return "", nil, fmt.Errorf("signature verification failed: %w", err)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, err
	}
	var claims tokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", nil, err
	}

	if claims.Exp < time.Now().Unix() {
		return "", nil, fmt.Errorf("token expired")
	}
	if claims.Sub == "" {
		return "", nil, fmt.Errorf("missing sub")
	}

	return claims.Sub, claims.RealmAccess.Roles, nil
}
