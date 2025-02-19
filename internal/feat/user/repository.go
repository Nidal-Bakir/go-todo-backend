package user

import (
	"context"
	"errors"

	apperr "github.com/Nidal-Bakir/go-todo-backend/internal/app_error"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/otp"
	"github.com/Nidal-Bakir/go-todo-backend/internal/gateway"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	usernaemgen "github.com/Nidal-Bakir/username_r_gen"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
)

type Repository interface {
	GetUserById(ctx context.Context, id int) (User, error)
	GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error)
	CreateTempUser(ctx context.Context, tUser *TempUser) (*TempUser, error)
	CreateUser(ctx context.Context, tempUserId uuid.UUID, otp string) (User, error)
}

func NewRepository(ds *dataSource, gatewaysProvider gateway.Provider) Repository {
	return repositoryImpl{dataSource: ds, gatewaysProvider: gatewaysProvider}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	dataSource       *dataSource
	gatewaysProvider gateway.Provider
}

func (repo repositoryImpl) GetUserById(ctx context.Context, id int) (User, error) {
	zlog := zerolog.Ctx(ctx).With().Int("user_id", id).Logger()

	dbUser, err := repo.dataSource.GetUserById(ctx, id)
	if err != nil {
		if !errors.Is(err, apperr.ErrNoResult) {
			zlog.Err(err).Msg("error geting the user by user id")
		}
		return User{}, err
	}

	user := NewUserFromDatabaseUser(dbUser)
	return user, nil
}

func (repo repositoryImpl) GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error) {
	zlog := zerolog.Ctx(ctx)

	dbUser, err := repo.dataSource.GetUserBySessionToken(ctx, sessionToken)
	if err != nil {
		if !errors.Is(err, apperr.ErrNoResult) {
			zlog.Err(err).Msg("error geting the user by session token")
		}
		return User{}, err
	}

	user := NewUserFromDatabaseUser(dbUser)
	return user, nil
}

func (repo repositoryImpl) CreateTempUser(ctx context.Context, tUser *TempUser) (*TempUser, error) {
	zlog := zerolog.Ctx(ctx)

	tUser.Id = uuid.New()
	tUser.Username = tUser.Id.String()

	sentOtp, err := sendOtpToTempUser(ctx, repo.gatewaysProvider, *tUser)
	if err != nil {
		zlog.Err(err).Msg("error sending otp to temp user")
		return tUser, err
	}
	tUser.SentOTP = sentOtp

	tUser, err = repo.dataSource.SetUserInTempCache(ctx, tUser)
	if err != nil {
		zlog.Err(err).Msg("error creating temp user in the cache database")
	}

	return tUser, err
}

func sendOtpToTempUser(ctx context.Context, gatewaysProvider gateway.Provider, tUser TempUser) (sentOTP string, err error) {
	otpSender := otp.NewOTPSender(gatewaysProvider)

	if tUser.UsernameType.isUsingEmail() {
		sentOTP, err = otpSender.SendEmailOTP(ctx, tUser.Email)
		return sentOTP, err
	}

	if tUser.UsernameType.isUsingPhoneNumber() {
		sentOTP, err = otpSender.SendSMSOTP(ctx, tUser.Phone)
		return sentOTP, err
	}

	panic("we should not be here!")
}

func (repo repositoryImpl) CreateUser(ctx context.Context, tempUserId uuid.UUID, providedOTP string) (User, error) {
	tUser, err := repo.getTempUser(ctx, tempUserId)
	if err != nil {
		return User{}, err
	}

	err = checkTempUserOTP(tUser, providedOTP)
	if err != nil {
		return User{}, err
	}

	user, err := repo.storUser(ctx, tUser)
	if err != nil {
		return User{}, err
	}

	repo.deleteTempUserFromCache(ctx, tUser)

	return user, err
}

func (repo repositoryImpl) getTempUser(ctx context.Context, id uuid.UUID) (*TempUser, error) {
	zlog := zerolog.Ctx(ctx)

	tUser, err := repo.dataSource.GetUserFromTempCache(ctx, id)
	if err != nil {
		if !errors.Is(err, apperr.ErrNoResult) {
			zlog.Err(err).Msg("error geting user from cache")
		}
		return nil, apperr.ErrNoResult
	}

	utils.AssertDev(tUser != nil, "tuser should not be nil! something is wrong")
	if tUser == nil {
		return nil, apperr.ErrNoResult
	}

	return tUser, nil
}

func checkTempUserOTP(tUser *TempUser, providedOTP string) error {
	if tUser.SentOTP != providedOTP {
		return ErrInvalidOtpCode
	}
	return nil
}

func (repo repositoryImpl) storUser(ctx context.Context, tUser *TempUser) (User, error) {
	zlog := zerolog.Ctx(ctx)

	if ok := tUser.ValidateForStore(); !ok {
		return User{}, ErrInvalidTempUserdata
	}

	dbUser, err := repo.dataSource.CreateUser(ctx, tUser)
	if err != nil {
		zlog.Err(err).Msg("error while create new user in the database")
		return User{}, err
	}

	username := usernaemgen.NewUsernameGen().Generate(int64(dbUser.ID))
	err = repo.dataSource.UpdateusernameForUser(ctx, dbUser.ID, username)
	if err != nil {
		zlog.Err(err).Msg("error while updating the username for user for the first time")
		// we do not have to return with error because the username is set with some random UUID
	} else {
		// use the new username
		dbUser.Username = username
	}

	user := NewUserFromDatabaseUser(dbUser)
	return user, nil
}

func (repo repositoryImpl) deleteTempUserFromCache(ctx context.Context, tUser *TempUser) {
	zlog := zerolog.Ctx(ctx)

	// ignore any error because the temp user will be auto cleand by the redis after sometime
	if err := repo.dataSource.DeleteUserFromTempCache(ctx, tUser.Id); err != nil {
		zlog.Err(err).Msg("error while deleting user form temp cache")
	}
}
