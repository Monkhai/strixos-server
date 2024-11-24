package identity

import "sync"

type IndentitiesMap struct {
	identities map[string]Identity
	Mux        *sync.RWMutex
}

func NewIdentitiesMap() *IndentitiesMap {
	return &IndentitiesMap{
		identities: make(map[string]Identity),
		Mux:        &sync.RWMutex{},
	}
}

func (i *IndentitiesMap) AddIdentity(newIdentity Identity) error {
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

func (i *IndentitiesMap) RemoveIdentity(id string) {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	delete(i.identities, id)
}

func (i *IndentitiesMap) GetIdentity(id string) (Identity, error) {
	i.Mux.RLock()
	defer i.Mux.RUnlock()
	identity, ok := i.identities[id]
	if !ok {
		return Identity{}, ErrIdentityNotFound
	}
	return identity, nil
}
