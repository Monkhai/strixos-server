package identity

import "errors"

var (
	ErrSpoofedIdentity  = errors.New("spoofed identity")
	ErrIndentityExists  = errors.New("identity already exists")
	ErrIdentityNotFound = errors.New("identity not found")
)
