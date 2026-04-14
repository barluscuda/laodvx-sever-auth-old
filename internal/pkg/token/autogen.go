package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// EnsureSeasonKeySet manages the lifecycle of EC key pairs inside dir:
//
//  1. Creates dir if it does not exist.
//  2. Prunes any season-format keys that are outside the allowed window
//     (removes their JWKS entry and deletes the private key file).
//  3. Generates a fresh P-256 key pair for the current season if one
//     does not already exist, persisting it as <kid>.private.pem.
//  4. Writes an up-to-date jwks.json to dir.
//  5. Returns a KeySet ready for signing and verification.
func EnsureSeasonKeySet(dir string) (*KeySet, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create keys dir %q: %w", dir, err)
	}

	now := currentSeason()
	currentKID := now.String()
	jwksPath := filepath.Join(dir, "jwks.json")
	privPath := filepath.Join(dir, currentKID+".private.pem")

	jwks, keys, err := loadOrCreateJWKS(jwksPath)
	if err != nil {
		return nil, err
	}

	jwks, keys = pruneExpiredKeys(jwks, keys, now, dir)

	// Generate a new key pair if the current season's private key is absent.
	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate EC key: %w", err)
		}
		if err := writePrivateKeyPEM(privPath, priv); err != nil {
			return nil, err
		}
		log.Printf("token: generated new EC key kid=%q → %s", currentKID, privPath)
		jwk := ecPublicKeyToJWK(currentKID, &priv.PublicKey)
		jwks.Keys = append(jwks.Keys, jwk)
		keys[currentKID] = &priv.PublicKey
	}

	// Defensive: private key file exists but JWKS entry is missing (e.g. after a
	// mid-write crash on a previous startup). Re-derive the public JWK from disk.
	if _, ok := keys[currentKID]; !ok {
		priv, err := readECPrivateKeyPEM(privPath)
		if err != nil {
			return nil, err
		}
		jwk := ecPublicKeyToJWK(currentKID, &priv.PublicKey)
		jwks.Keys = append(jwks.Keys, jwk)
		keys[currentKID] = &priv.PublicKey
	}

	if err := persistJWKS(jwksPath, jwks); err != nil {
		return nil, err
	}

	activeKey, err := readECPrivateKeyPEM(privPath)
	if err != nil {
		return nil, err
	}

	return &KeySet{
		keys:      keys,
		jwks:      jwks,
		activeKID: currentKID,
		activeKey: activeKey,
	}, nil
}

// pruneExpiredKeys removes season-format entries that are outside the allowed
// window from both the JWKS slice and the keys map, and deletes their private
// key files from disk.
func pruneExpiredKeys(jwks JWKS, keys map[string]*ecdsa.PublicKey, now seasonKID, dir string) (JWKS, map[string]*ecdsa.PublicKey) {
	kept := jwks.Keys[:0:0]
	for _, k := range jwks.Keys {
		sk, ok := parseSeasonKID(k.Kid)
		if ok && !sk.withinWindow(now) {
			delete(keys, k.Kid)
			privPath := filepath.Join(dir, k.Kid+".private.pem")
			if err := os.Remove(privPath); err != nil && !os.IsNotExist(err) {
				log.Printf("token: warning: could not remove expired key file %s: %v", privPath, err)
			}
			log.Printf("token: pruned expired key kid=%q", k.Kid)
			continue
		}
		kept = append(kept, k)
	}
	jwks.Keys = kept
	return jwks, keys
}

func loadOrCreateJWKS(path string) (JWKS, map[string]*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return JWKS{}, make(map[string]*ecdsa.PublicKey), nil
	}
	if err != nil {
		return JWKS{}, nil, fmt.Errorf("read jwks %q: %w", path, err)
	}
	var jwks JWKS
	if err := json.Unmarshal(data, &jwks); err != nil {
		return JWKS{}, nil, fmt.Errorf("parse jwks %q: %w", path, err)
	}
	keys := make(map[string]*ecdsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		pub, err := jwkToECPublicKey(k)
		if err != nil {
			return JWKS{}, nil, fmt.Errorf("jwk kid=%q: %w", k.Kid, err)
		}
		keys[k.Kid] = pub
	}
	return jwks, keys, nil
}

func persistJWKS(path string, jwks JWKS) error {
	data, err := json.MarshalIndent(jwks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal jwks: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write jwks %q: %w", path, err)
	}
	return nil
}

func writePrivateKeyPEM(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		return fmt.Errorf("write private key %q: %w", path, err)
	}
	return nil
}

func readECPrivateKeyPEM(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key %q: %w", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode PEM from %q", path)
	}
	// Try PKCS8 first (our write format), then fall back to SEC1.
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ecKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key from %q is not an EC private key", path)
		}
		return ecKey, nil
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func ecPublicKeyToJWK(kid string, pub *ecdsa.PublicKey) JWK {
	size := (pub.Curve.Params().BitSize + 7) / 8
	xPad := make([]byte, size)
	yPad := make([]byte, size)
	xB := pub.X.Bytes()
	yB := pub.Y.Bytes()
	copy(xPad[size-len(xB):], xB)
	copy(yPad[size-len(yB):], yB)
	return JWK{
		Kty: "EC",
		Kid: kid,
		Use: "sig",
		Alg: "ES256",
		Crv: pub.Curve.Params().Name,
		X:   base64.RawURLEncoding.EncodeToString(xPad),
		Y:   base64.RawURLEncoding.EncodeToString(yPad),
	}
}
