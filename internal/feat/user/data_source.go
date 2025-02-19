package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperr "github.com/Nidal-Bakir/go-todo-backend/internal/app_error"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type dataSource struct {
	db    *database.Service
	redis *redis.Client
}

func NewDataSource(db *database.Service, redis *redis.Client) *dataSource {
	return &dataSource{db: db, redis: redis}
}

func (ds dataSource) GetUserById(ctx context.Context, id int) (database.User, error) {
	userId, err := utils.SafeIntToInt32(id)
	if err != nil {
		return database.User{}, err
	}

	dbUser, err := ds.db.Queries.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbUser, apperr.ErrNoResult
		}
		return dbUser, err
	}

	return dbUser, nil
}

func (ds dataSource) GetUserBySessionToken(ctx context.Context, sessionToken string) (database.User, error) {
	dbUser, err := ds.db.Queries.GetUserBySessionToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbUser, apperr.ErrNoResult
		}
		return dbUser, err
	}

	return dbUser, err
}

func (ds dataSource) genTempUserId(id uuid.UUID) string {
	return fmt.Sprint("user:tmp:", id.String())
}

func (ds dataSource) SetUserInTempCache(ctx context.Context, tUser *TempUser) (*TempUser, error) {
	key := ds.genTempUserId(tUser.Id)

	pip := ds.redis.TxPipeline()
	pip.Del(ctx, key)
	pip.HSet(ctx, key, tUser.ToMap())
	pip.Expire(ctx, key, time.Minute*30)
	resultArray, err := pip.Exec(ctx)

	if err != nil {
		return tUser, err
	}

	for _, cmdResult := range resultArray {
		if cmdResult.Err() != nil {
			return tUser, err
		}
	}

	return tUser, nil
}

func (ds dataSource) GetUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) (*TempUser, error) {
	key := ds.genTempUserId(tempUserId)

	result, err := ds.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, apperr.ErrNoResult
	}

	tUser := new(TempUser)
	tUser.FromMap(result)

	return tUser, err
}

func (ds dataSource) DeleteUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) error {
	return ds.redis.Del(ctx, ds.genTempUserId(tempUserId)).Err()
}

func (ds dataSource) CreateUser(ctx context.Context, tUser *TempUser) (database.User, error) {
	args := database.CreateNewUserParams{
		FirstName: tUser.Fname,
		Username:  tUser.Username,
		LastName:  pgtype.Text{String: tUser.Lname},
	}
	return ds.db.Queries.CreateNewUser(ctx, args)
}

func (ds dataSource) UpdateusernameForUser(ctx context.Context, id int32, newUsername string) error {
	return ds.db.Queries.UpdateUsernameForUser(ctx, database.UpdateUsernameForUserParams{ID: id, Username: newUsername})
}
