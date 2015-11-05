package signed

import (
	"crypto/rand"
	"errors"

	"github.com/agl/ed25519"
	"github.com/docker/notary/tuf/data"
)

type edCryptoKey struct {
	role    string
	privKey data.PrivateKey
}

// Ed25519 implements a simple in memory cryptosystem for ED25519 keys
type Ed25519 struct {
	keys map[string]edCryptoKey
}

// NewEd25519 initializes a new empty Ed25519 CryptoService that operates
// entirely in memory
func NewEd25519() *Ed25519 {
	return &Ed25519{
		make(map[string]edCryptoKey),
	}
}

// addKey allows you to add a private key
func (e *Ed25519) addKey(role string, k data.PrivateKey) {
	e.keys[k.ID()] = edCryptoKey{
		role:    role,
		privKey: k,
	}
}

// RemoveKey deletes a key from the signer
func (e *Ed25519) RemoveKey(keyID string) error {
	delete(e.keys, keyID)
	return nil
}

// ListKeys returns the list of keys IDs for the role
func (e *Ed25519) ListKeys(role string) []string {
	keyIDs := make([]string, 0, len(e.keys))
	for id := range e.keys {
		keyIDs = append(keyIDs, id)
	}
	return keyIDs
}

// ListKeys returns the list of keys IDs for the role
func (e *Ed25519) ListAllKeys() map[string]string {
	keys := make(map[string]string)
	for id, edKey := range e.keys {
		keys[id] = edKey.role
	}
	return keys
}

// Sign generates an Ed25519 signature over the data
func (e *Ed25519) Sign(keyIDs []string, toSign []byte) ([]data.Signature, error) {
	signatures := make([]data.Signature, 0, len(keyIDs))
	for _, keyID := range keyIDs {
		priv := [ed25519.PrivateKeySize]byte{}
		copy(priv[:], e.keys[keyID].privKey.Private())
		sig := ed25519.Sign(&priv, toSign)
		signatures = append(signatures, data.Signature{
			KeyID:     keyID,
			Method:    data.EDDSASignature,
			Signature: sig[:],
		})
	}
	return signatures, nil

}

// Create generates a new key and returns the public part
func (e *Ed25519) Create(role, algorithm string) (data.PublicKey, error) {
	if algorithm != data.ED25519Key {
		return nil, errors.New("only ED25519 supported by this cryptoservice")
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	public := data.NewED25519PublicKey(pub[:])
	private, err := data.NewED25519PrivateKey(*public, priv[:])
	if err != nil {
		return nil, err
	}

	e.addKey(role, private)
	return public, nil
}

// PublicKeys returns a map of public keys for the ids provided, when those IDs are found
// in the store.
func (e *Ed25519) PublicKeys(keyIDs ...string) (map[string]data.PublicKey, error) {
	k := make(map[string]data.PublicKey)
	for _, keyID := range keyIDs {
		if edKey, ok := e.keys[keyID]; ok {
			k[keyID] = data.PublicKeyFromPrivate(edKey.privKey)
		}
	}
	return k, nil
}

// GetKey returns a single public key based on the ID
func (e *Ed25519) GetKey(keyID string) data.PublicKey {
	return data.PublicKeyFromPrivate(e.keys[keyID].privKey)
}

// GetPrivateKey returns a single private key based on the ID
func (e *Ed25519) GetPrivateKey(keyID string) (data.PrivateKey, string, error) {
	return e.keys[keyID].privKey, "", nil
}
