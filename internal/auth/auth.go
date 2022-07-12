package auth

import (
	"errors"
	"net/http"
)

type requestContextKey uint

const (
	userPrincipalContextKey requestContextKey = iota
)

type UserPrincipal struct {
	Id       int
	Username string
	Email    string
}

func GetUserPrincipal(r *http.Request) (*UserPrincipal, error) {
	principal, ok := r.Context().Value(userPrincipalContextKey).(*UserPrincipal)
	if !ok {
		return nil, errors.New("request context does not contain user principal")
	}
	return principal, nil
}
