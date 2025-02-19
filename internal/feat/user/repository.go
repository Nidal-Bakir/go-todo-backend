package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/otp"
	"github.com/Nidal-Bakir/go-todo-backend/internal/gateway"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	GetUserById(ctx context.Context, id int) (User, error)
	GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error)
	CreateTempUser(ctx context.Context, tUser TempUser) (TempUser, error)
}

func NewRepository(db *database.Service, redis *redis.Client, gatewaysProvider gateway.Provider) Repository {
	return repositoryImpl{db: db, redis: redis, gatewaysProvider: gatewaysProvider}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	db               *database.Service
	redis            *redis.Client
	gatewaysProvider gateway.Provider
}

func (repo repositoryImpl) GetUserById(ctx context.Context, id int) (User, error) {
	userId, err := utils.SafeIntToInt32(id)
	if err != nil {
		return User{}, err
	}

	dbUser, err := repo.db.Queries.GetUserById(ctx, userId)
	user := NewUserFromDatabaseuser(dbUser)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (repo repositoryImpl) GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error) {
	dbUser, err := repo.db.Queries.GetUserBySessionToken(ctx, sessionToken)
	if err != nil {
		return User{}, err
	}

	user := NewUserFromDatabaseuser(dbUser)
	return user, err
}

func (repo repositoryImpl) CreateTempUser(ctx context.Context, tUser TempUser) (TempUser, error) {
	otpSender := otp.NewOTPSender(repo.gatewaysProvider)
	
	
	// TODO: Add logger to this code

	var otp string
	var err error
	if tUser.isUsingEmail() {
		otp, err = otpSender.SendEmailOTP(ctx, tUser.email)
	} else if tUser.isUsingPhoneNumber() {
		otp, err = otpSender.SendSMSOTP(ctx, tUser.phone)
	} else {
		panic("we should not be here!")
	}

	if err != nil {
		return TempUser{}, err
	}
	tUser.otp = otp

	key := fmt.Sprint("user:tmp:", tUser.Key())
	pip := repo.redis.TxPipeline()
	pip.Del(ctx, key)
	pip.HSet(ctx, key)
	pip.Expire(ctx, key, time.Hour*1)
	resultArray, err := pip.Exec(ctx)

	if err != nil {
		return TempUser{}, err
	}
	for _, cmdResult := range resultArray {
		if cmdResult.Err() != nil {
			return TempUser{}, err
		}
	}

	return tUser, nil
}
