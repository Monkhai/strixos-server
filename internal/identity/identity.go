package identity

type Identity struct {
	ID          string `json:"id"`
	Secret      string `json:"secret"`
	Avatar      string `json:"avatar"`
	DisplayName string `json:"displayName"`
}

func NewIdentity(id, secret string) *Identity {
	return &Identity{
		ID:          id,
		Secret:      secret,
		Avatar:      AVATAR_DEFAULT,
		DisplayName: "",
	}
}

type SafeIdentity struct {
	ID          string `json:"id"`
	Avatar      string `json:"avatar"`
	DisplayName string `json:"displayName"`
}

type InitialIdentity struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func NewInitialIdentity(id, secret string) *InitialIdentity {
	return &InitialIdentity{
		ID:     id,
		Secret: secret,
	}
}

func (i *Identity) GetSafeIdentity() *SafeIdentity {
	return &SafeIdentity{
		ID:          i.ID,
		Avatar:      i.Avatar,
		DisplayName: i.DisplayName,
	}
}
