// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package whsess

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"sync"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/net/context"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/wherr"
)

const (
	nonceLength  = 24
	keyLength    = 32
	minKeyLength = 10
)

type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

type CookieStore struct {
	Options  CookieOptions
	secretCB func(context.Context) ([]byte, error)

	secretMtx   sync.Mutex
	secretSetup bool
	secret      [keyLength]byte
}

var _ Store = (*CookieStore)(nil)

// NewCookieStore creates a secure cookie store with default settings.
// Configure the Options field further if additional settings are required.
func NewCookieStore(secretKey []byte) *CookieStore {
	rv := &CookieStore{
		Options: CookieOptions{
			Path:   "/",
			MaxAge: 86400 * 30},
		secretCB: func(context.Context) ([]byte, error) {
			return secretKey, nil
		},
	}
	return rv
}

// NewLazyCookieStore is like NewCookieStore but loads the secretKey using
// the provided callback once. This is useful for delayed initialization after
// the first request for something like App Engine where you can't interact
// with a database without a context.
func NewLazyCookieStore(secretKey func(context.Context) ([]byte, error)) (
	cs *CookieStore) {
	rv := &CookieStore{
		Options: CookieOptions{
			Path:   "/",
			MaxAge: 86400 * 30},
		secretCB: secretKey,
	}
	return rv
}

func (cs *CookieStore) getSecret(ctx context.Context) (
	*[keyLength]byte, error) {
	cs.secretMtx.Lock()
	defer cs.secretMtx.Unlock()
	if cs.secretSetup {
		return &cs.secret, nil
	}
	secret, err := cs.secretCB(ctx)
	if err != nil {
		return nil, err
	}
	if len(secret) < minKeyLength {
		return nil, wherr.InternalServerError.New("cookie secret not long enough")
	}
	secretHash := sha256.Sum256(secret)
	copy(cs.secret[:], secretHash[:])
	cs.secretSetup = true
	return &cs.secret, nil
}

// Load implements the Store interface. Not expected to be used directly.
func (cs *CookieStore) Load(ctx context.Context, r *http.Request,
	namespace string) (rv SessionData, err error) {
	empty := SessionData{New: true, Values: map[interface{}]interface{}{}}
	secret, err := cs.getSecret(whcompat.Context(r))
	if err != nil {
		return empty, err
	}
	c, err := r.Cookie(namespace)
	if err != nil || c.Value == "" {
		return empty, nil
	}
	data, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		return empty, nil
	}
	var nonce [nonceLength]byte
	copy(nonce[:], data[:nonceLength])
	decrypted, ok := secretbox.Open(nil, data[nonceLength:], &nonce,
		secret)
	if !ok {
		return empty, nil
	}
	err = gob.NewDecoder(bytes.NewReader(decrypted)).Decode(&rv.Values)
	if err != nil {
		return empty, nil
	}
	return rv, nil
}

// Save implements the Store interface. Not expected to be used directly.
func (cs *CookieStore) Save(ctx context.Context, w http.ResponseWriter,
	namespace string, s SessionData) error {
	secret, err := cs.getSecret(ctx)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = gob.NewEncoder(&out).Encode(&s.Values)
	if err != nil {
		return err
	}
	var nonce [nonceLength]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		return err
	}
	value := base64.URLEncoding.EncodeToString(
		secretbox.Seal(nonce[:], out.Bytes(), &nonce, secret))

	return setCookie(w, &http.Cookie{
		Name:     namespace,
		Value:    value,
		Path:     cs.Options.Path,
		Domain:   cs.Options.Domain,
		MaxAge:   cs.Options.MaxAge,
		Secure:   cs.Options.Secure,
		HttpOnly: cs.Options.HttpOnly})
}

func setCookie(w http.ResponseWriter, cookie *http.Cookie) error {
	v := cookie.String()
	if v == "" {
		return SessionError.New("invalid cookie %#v", cookie.Name)
	}
	w.Header().Add("Set-Cookie", v)
	return nil
}

// Clear implements the Store interface. Not expected to be used directly.
func (cs *CookieStore) Clear(ctx context.Context, w http.ResponseWriter,
	namespace string) error {
	return setCookie(w, &http.Cookie{
		Name:     namespace,
		Value:    "",
		Path:     cs.Options.Path,
		Domain:   cs.Options.Domain,
		MaxAge:   -1,
		Secure:   cs.Options.Secure,
		HttpOnly: cs.Options.HttpOnly})
}
