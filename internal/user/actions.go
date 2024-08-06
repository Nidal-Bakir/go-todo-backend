package user

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

type userCtxKeysType int

const (
	currentUserCtxKey userCtxKeysType = iota
)

func ContextWithUser(ctx context.Context, user database.User) context.Context {
	return context.WithValue(ctx, currentUserCtxKey, user)
}

func UserFromContext(ctx context.Context) (database.User, bool) {
	user, ok := ctx.Value(currentUserCtxKey).(database.User)
	return user, ok
}

type Actions interface {
	GetUserById(ctx context.Context, id int) (database.User, error)
	GetUserBySessionToken(ctx context.Context, sessionToken string) (database.User, error)
}

func NewActions(db database.Querier) Actions {
	return actionsImpl{db: db}
}

type actionsImpl struct {
	db database.Querier
}

func (a actionsImpl) GetUserById(ctx context.Context, id int) (database.User, error) {
	userId, err := utils.SafeIntToInt32(id)
	if err != nil {
		return database.User{}, err
	}

	user, err := a.db.GetUserById(ctx, userId)
	if err != nil {
		return database.User{}, err
	}

	return user, nil
}

func (a actionsImpl) GetUserBySessionToken(ctx context.Context, sessionToken string) (database.User, error) {
	// TODO: implement getting user using session token
	return database.User{}, nil
}
