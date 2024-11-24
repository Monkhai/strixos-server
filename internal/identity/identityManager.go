package identity

import (
	"crypto/rand"
	"encoding/hex"
)

type Identity struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type IdentityManager struct {
	IdentitiesMap *IndentitiesMap
}

func NewIdentityManager() *IdentityManager {
	return &IdentityManager{
		IdentitiesMap: NewIdentitiesMap(),
	}
}

func (i *IdentityManager) generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate user ID")
	}
	return hex.EncodeToString(bytes)
}

func (i *IdentityManager) generateSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate secret")
	}
	return hex.EncodeToString(bytes)
}

func (i *IdentityManager) RegisterIdentity() Identity {
	id := i.generateID()
	secret := i.generateSecret()
	identity := Identity{
		ID:     id,
		Secret: secret,
	}
	i.IdentitiesMap.AddIdentity(identity)
	return identity
}

func (i *IdentityManager) ValidateIdentity(identity Identity) (bool, error) {
	mappedIdentity, err := i.IdentitiesMap.GetIdentity(identity.ID)
	if err != nil {
		return false, err
	}
	if mappedIdentity.Secret != identity.Secret {
		return false, ErrSpoofedIdentity
	}

	return true, nil
}
