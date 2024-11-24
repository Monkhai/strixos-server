package identity

import (
	"log"
	"sync"
)

type IndentitiesMap struct {
	identities map[string]*Identity
	Mux        *sync.RWMutex
}

func NewIdentitiesMap() *IndentitiesMap {
	return &IndentitiesMap{
		identities: make(map[string]*Identity),
		Mux:        &sync.RWMutex{},
	}
}

func (i *IndentitiesMap) AddIdentity(newIdentity *Identity) error {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	identity, ok := i.identities[newIdentity.ID]
	if ok {
		if identity.Secret != newIdentity.Secret {
			return ErrSpoofedIdentity
		}
		return ErrIndentityExists
	}

	i.identities[newIdentity.ID] = newIdentity
	return nil
}

func (i *IndentitiesMap) ValidateIdentity(identity *Identity) (bool, error) {
	i.Mux.RLock()
	defer i.Mux.RUnlock()
	mappedIdentity, err := i.GetIdentity(identity.ID)
	if err != nil {
		return false, err
	}
	if mappedIdentity.Secret != identity.Secret {
		log.Printf("Spoofed identity: map %s, received %s\n", mappedIdentity.Secret, identity.Secret)
		return false, ErrSpoofedIdentity
	}
	return true, nil
}

func (i *IndentitiesMap) UpdateIdentity(updatedIdentity Identity) bool {
	i.Mux.Lock()
	identity, ok := i.identities[updatedIdentity.ID]
	if !ok {
		return false
	}
	i.Mux.Unlock()

	valid, err := i.ValidateIdentity(identity)
	if !valid || err != nil {
		return false
	}

	i.Mux.Lock()
	newIdentity := &Identity{
		ID:          identity.ID,
		Secret:      identity.Secret,
		Avatar:      updatedIdentity.Avatar,
		DisplayName: updatedIdentity.DisplayName,
	}
	i.identities[identity.ID] = newIdentity
	i.Mux.Unlock()

	return true
}

func (i *IndentitiesMap) RemoveIdentity(id string) {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	delete(i.identities, id)
}

func (i *IndentitiesMap) GetIdentity(id string) (*Identity, error) {
	i.Mux.RLock()
	defer i.Mux.RUnlock()
	identity, ok := i.identities[id]
	if !ok {
		return &Identity{}, ErrIdentityNotFound
	}
	return identity, nil
}
