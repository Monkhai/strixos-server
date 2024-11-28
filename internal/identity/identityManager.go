package identity

import (
	"github.com/Monkhai/strixos-server.git/pkg/utils"
)

type IdentityManager struct {
	IdentitiesMap *IndentitiesMap
}

func NewIdentityManager() *IdentityManager {
	return &IdentityManager{
		IdentitiesMap: NewIdentitiesMap(),
	}
}

func (i *IdentityManager) RegisterIdentity() *Identity {
	id := utils.GenerateUniqueID()
	secret := utils.GenerateSecret()
	identity := NewIdentity(id, secret)
	i.IdentitiesMap.AddIdentity(identity)
	return identity
}
