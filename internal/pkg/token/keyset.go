package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWK is a single JSON Web Key (public EC key only).
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// JWKS is the JSON Web Key Set container served at /.well-known/jwks.json.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// KeySet holds all EC public keys by kid for JWT verification,
// plus one active private key for signing new tokens.
type KeySet struct {
	keys      map[string]*ecdsa.PublicKey
	jwks      JWKS
	activeKID string
	activeKey *ecdsa.PrivateKey
}

// LoadKeySet parses jwksPath (a JWKS JSON file) and returns a ready KeySet.
// activeKID must be present in the JWKS; activeKey is the corresponding private key.
func LoadKeySet(jwksPath, activeKID string, activeKey *ecdsa.PrivateKey) (*KeySet, error) {
	data, err := os.ReadFile(jwksPath)
	if err != nil {
		return nil, fmt.Errorf("read jwks %q: %w", jwksPath, err)
	}
	var jwks JWKS
	if err := json.Unmarshal(data, &jwks); err != nil {
		return nil, fmt.Errorf("parse jwks %q: %w", jwksPath, err)
	}

	now := currentSeason()
	keys := make(map[string]*ecdsa.PublicKey, len(jwks.Keys))
	filtered := jwks.Keys[:0:0] // same backing array, zero length
	for _, k := range jwks.Keys {
		// Drop season-format kids that are outside the allowed window.
		if sk, ok := parseSeasonKID(k.Kid); ok && !sk.withinWindow(now) {
			continue
		}
		pub, err := jwkToECPublicKey(k)
		if err != nil {
			return nil, fmt.Errorf("jwk kid=%q: %w", k.Kid, err)
		}
		keys[k.Kid] = pub
		filtered = append(filtered, k)
	}
	jwks.Keys = filtered

	if _, ok := keys[activeKID]; !ok {
		return nil, fmt.Errorf("active kid %q not found in %s (expired or missing)", activeKID, jwksPath)
	}

	return &KeySet{
		keys:      keys,
		jwks:      jwks,
		activeKID: activeKID,
		activeKey: activeKey,
	}, nil
}

// ActiveKID returns the kid used to sign new tokens.
func (ks *KeySet) ActiveKID() string { return ks.activeKID }

// ActiveKey returns the private key used to sign new tokens.
func (ks *KeySet) ActiveKey() *ecdsa.PrivateKey { return ks.activeKey }

// PublicJWKS returns the public JWKS suitable for serving at /.well-known/jwks.json.
func (ks *KeySet) PublicJWKS() JWKS { return ks.jwks }

// KeyFunc is a jwt.Keyfunc that resolves the EC public key by the token's kid header.
// When the kid matches the season format "YYYY-sN", it also enforces that the kid
// is within the allowed window: current season or the immediately preceding season.
func (ks *KeySet) KeyFunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
		return nil, jwt.ErrSignatureInvalid
	}
	kid, _ := t.Header["kid"].(string)
	if kid == "" {
		return nil, fmt.Errorf("missing kid header")
	}
	if sk, ok := parseSeasonKID(kid); ok {
		if !sk.withinWindow(currentSeason()) {
			return nil, fmt.Errorf("kid %q is outside the allowed season window", kid)
		}
	}
	pub, ok := ks.keys[kid]
	if !ok {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}
	return pub, nil
}

// seasonKID represents a parsed "YYYY-sN" key ID (N ∈ 1..4).
type seasonKID struct {
	year   int
	season int
}

// parseSeasonKID parses a kid of the form "2026-s3".
// Returns (parsed, true) on success, (zero, false) if the format doesn't match.
func parseSeasonKID(kid string) (seasonKID, bool) {
	// Exactly "YYYY-sN" — 7 characters.
	if len(kid) != 7 {
		return seasonKID{}, false
	}
	year, err := strconv.Atoi(kid[:4])
	if err != nil || kid[4] != '-' || kid[5] != 's' {
		return seasonKID{}, false
	}
	s := int(kid[6] - '0')
	if s < 1 || s > 4 {
		return seasonKID{}, false
	}
	return seasonKID{year: year, season: s}, true
}

// currentSeason returns the season kid for the current wall-clock time.
// s1 = Jan–Mar, s2 = Apr–Jun, s3 = Jul–Sep, s4 = Oct–Dec.
func currentSeason() seasonKID {
	now := time.Now()
	return seasonKID{
		year:   now.Year(),
		season: (int(now.Month())-1)/3 + 1,
	}
}

// String returns the canonical kid string, e.g. "2026-s2".
func (s seasonKID) String() string { return fmt.Sprintf("%d-s%d", s.year, s.season) }

// index converts a season to a monotonically increasing integer so seasons
// can be compared across year boundaries (e.g. 2025-s4 < 2026-s1).
func (s seasonKID) index() int { return s.year*4 + s.season - 1 }

// withinWindow returns true when s is the current season or exactly one season behind it.
func (s seasonKID) withinWindow(current seasonKID) bool {
	diff := current.index() - s.index()
	return diff == 0 || diff == 1
}

func jwkToECPublicKey(jwk JWK) (*ecdsa.PublicKey, error) {
	if jwk.Kty != "EC" {
		return nil, fmt.Errorf("unsupported kty %q, want EC", jwk.Kty)
	}

	var curve elliptic.Curve
	switch jwk.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported crv %q", jwk.Crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
