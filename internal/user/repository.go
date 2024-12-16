package user

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	GetUserById(ctx context.Context, id int) (User, error)
	GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error)
}

func NewRepository(db *database.Service, redis *redis.Client) Repository {
	return repositoryImpl{db: db, redis: redis}
}

type repositoryImpl struct {
	db    *database.Service
	redis *redis.Client
}

func (a repositoryImpl) GetUserById(ctx context.Context, id int) (User, error) {
	userId, err := utils.SafeIntToInt32(id)
	if err != nil {
		return User{}, err
	}

	dbUser, err := a.db.Queries.GetUserById(ctx, userId)
	user := NewUserFromDatabaseuser(dbUser)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (a repositoryImpl) GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error) {
	dbUser, err := a.db.Queries.GetUserBySessionToken(ctx, sessionToken)
	if err != nil {
		return User{}, err
	}

	user := NewUserFromDatabaseuser(dbUser)
	return user, err
}
