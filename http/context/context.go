package context

import (
	"context"

	"github.com/go-thread-7/commonlib/http/constants"
	"github.com/google/uuid"
)

type UserCtxKey struct{}

type User struct {
	UserID uuid.UUID `json:"user_id"`
	Role   *string   `json:"role,omitempty"`
}

func GetUserFromCtx(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(UserCtxKey{}).(*User)
	if !ok {
		return nil, constants.ErrorUnauthorized
	}
	return user, nil
}
