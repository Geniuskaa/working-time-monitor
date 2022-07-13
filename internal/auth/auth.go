package auth

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
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

func GetUserPrincipal(r *http.Request, ctx context.Context) (*UserPrincipal, error) {
	tr := otel.Tracer("handler-GetUserPrincipal")
	ctx, span := tr.Start(ctx, "handler-GetUserPrincipal")
	defer span.End()

	principal, ok := r.Context().Value(userPrincipalContextKey).(*UserPrincipal)
	if !ok {
		return nil, errors.New("request context does not contain user principal")
	}
	return principal, nil
}
