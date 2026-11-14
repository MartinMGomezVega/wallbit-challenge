package http

import (
	"context"
	"strings"
)

// Roles de la demo.
const (
	RoleAccountant = "accountant"
	RoleClient     = "client"
)

// Identity es la identidad de demo tomada del header X-Demo-User.
type Identity struct {
	Role     string
	ClientID string
}

type identityKey struct{}

// WithIdentity guarda la identidad en el contexto.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// FromContext devuelve la identidad del contexto, si hay.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

// ParseIdentity interpreta "accountant" o "client:<clientId>".
func ParseIdentity(h string) (Identity, bool) {
	h = strings.TrimSpace(h)
	if h == RoleAccountant {
		return Identity{Role: RoleAccountant}, true
	}
	if id, ok := strings.CutPrefix(h, RoleClient+":"); ok && id != "" && !strings.ContainsAny(id, " :") {
		return Identity{Role: RoleClient, ClientID: id}, true
	}
	return Identity{}, false
}
