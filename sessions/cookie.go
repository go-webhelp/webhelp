// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package sessions

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"log"
	"net/http"

	"github.com/jtolds/webhelp"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	nonceLength = 24
	KeyLength   = 32
)

type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

type CookieStore struct {
	Options CookieOptions
	Secret  [KeyLength]byte
}

var _ Store = (*CookieStore)(nil)

// NewCookieStore creates a secure cookie store with default settings.
// Configure the Options field further if additional settings are required.
func NewCookieStore(secretKey []byte) *CookieStore {
	rv := &CookieStore{
		Options: CookieOptions{
			Path:   "/",
			MaxAge: 86400 * 30}}
	copy(rv.Secret[:], secretKey)
	return rv
}

// Load implements the Store interface. Not expected to be used directly.
func (cs *CookieStore) Load(r *http.Request, namespace string) (rv SessionData,
	err error) {
	empty := SessionData{New: true, Values: map[interface{}]interface{}{}}
	c, err := r.Cookie(namespace)
	if err != nil {
		log.Printf("error: %v\n", err)
		return empty, nil
	}
	log.Printf("loading cookie %#v %s", namespace, c.Value)
	data, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		log.Printf("error: %v\n", err)
		return empty, nil
	}
	var nonce [nonceLength]byte
	copy(nonce[:], data[:nonceLength])
	decrypted, ok := secretbox.Open(nil, data[nonceLength:], &nonce,
		&cs.Secret)
	if !ok {
		log.Printf("error: failed decrypting\n")
		return empty, nil
	}
	err = gob.NewDecoder(bytes.NewReader(decrypted)).Decode(&rv.Values)
	if err != nil {
		log.Printf("error: %v\n", err)
		return empty, nil
	}
	log.Printf("loaded session %#v %#v", namespace, rv)
	return rv, nil
}

// Save implements the Store interface. Not expected to be used directly.
func (cs *CookieStore) Save(w webhelp.ResponseWriter, namespace string,
	s SessionData) error {
	log.Printf("saving session %#v %#v", namespace, s)
	var out bytes.Buffer
	err := gob.NewEncoder(&out).Encode(&s.Values)
	if err != nil {
		log.Printf("failed encoding session: %v", err)
		return err
	}
	var nonce [nonceLength]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		log.Printf("failed generating nonce: %v", err)
		return err
	}
	value := base64.URLEncoding.EncodeToString(
		secretbox.Seal(nonce[:], out.Bytes(), &nonce, &cs.Secret))
	log.Printf("setting cookie %#v to %s", namespace, value)
	http.SetCookie(w, &http.Cookie{
		Name:     namespace,
		Value:    value,
		Path:     cs.Options.Path,
		Domain:   cs.Options.Domain,
		MaxAge:   cs.Options.MaxAge,
		Secure:   cs.Options.Secure,
		HttpOnly: cs.Options.HttpOnly})
	return nil
}
