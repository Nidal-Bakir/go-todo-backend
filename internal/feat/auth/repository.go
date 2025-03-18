package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/otp"
	"github.com/Nidal-Bakir/go-todo-backend/internal/gateway"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/appjwt"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/password_hasher"
	usernaemgen "github.com/Nidal-Bakir/username_r_gen"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
)

type PasswordLoginAccessKey struct {
	Phone utils.PhoneNumber
	Email string
}

type Repository interface {
	GetUserById(ctx context.Context, id int) (User, error)
	GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error)
	CreateTempUser(ctx context.Context, tUser *TempUser) (*TempUser, error)
	CreateUser(ctx context.Context, tempUserId uuid.UUID, otp string) (User, error)
	PasswordLogin(ctx context.Context, accessKey PasswordLoginAccessKey, password string, loginMethod LoginMethod, installation database.Installation) (user User, token string, err error)
	GetInstallationUsingUuid(ctx context.Context, InstallationId uuid.UUID, attachedToUserId *int32) (database.Installation, error)
}

func NewRepository(ds DataSource, gatewaysProvider gateway.Provider, passwordHasher password_hasher.PasswordHasher) Repository {
	return repositoryImpl{dataSource: ds, gatewaysProvider: gatewaysProvider, PasswordHasher: passwordHasher}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	dataSource       DataSource
	gatewaysProvider gateway.Provider
	PasswordHasher   password_hasher.PasswordHasher
}

func (repo repositoryImpl) GetUserById(ctx context.Context, id int) (User, error) {
	userId, err := utils.SafeIntToInt32(id)
	if err != nil {
		return User{}, err
	}

	zlog := zerolog.Ctx(ctx).With().Int32("user_id", userId).Logger()

	dbUser, err := repo.dataSource.GetUserById(ctx, userId)
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

	if ok := tUser.ValidateForStore(); !ok {
		return tUser, apperr.ErrInvalidTempUserdata
	}

	sentOtp, err := sendOtpToTempUser(ctx, repo.gatewaysProvider, *tUser)
	if err != nil {
		zlog.Err(err).Msg("error sending otp to temp user")
		return tUser, err
	}
	tUser.SentOTP = sentOtp

	tUser, err = repo.dataSource.StoreUserInTempCache(ctx, tUser)
	if err != nil {
		zlog.Err(err).Msg("error creating temp user in the cache database")
	}

	return tUser, err
}

func sendOtpToTempUser(ctx context.Context, gatewaysProvider gateway.Provider, tUser TempUser) (sentOTP string, err error) {
	otpSender := otp.NewOTPSender(gatewaysProvider)

	if tUser.LoginMethod.isUsingEmail() {
		sentOTP, err = otpSender.SendEmailOTP(ctx, tUser.Email)
		return sentOTP, err
	}

	if tUser.LoginMethod.isUsingPhoneNumber() {
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
		return nil, err
	}

	if tUser == nil {
		return nil, apperr.ErrNoResult
	}

	return tUser, nil
}

func checkTempUserOTP(tUser *TempUser, providedOTP string) error {
	if tUser.SentOTP != providedOTP {
		return apperr.ErrInvalidOtpCode
	}
	return nil
}

func (repo repositoryImpl) storUser(ctx context.Context, tUser *TempUser) (User, error) {
	zlog := zerolog.Ctx(ctx)

	if ok := tUser.ValidateForStore(); !ok {
		return User{}, apperr.ErrInvalidTempUserdata
	}

	createUserArgs, err := generateUserArgsForCreateUser(tUser, repo.PasswordHasher)
	if err != nil {
		zlog.Err(err).Msg("error generating user args")
		return User{}, err
	}

	dbUser, err := repo.dataSource.CreateUser(ctx, createUserArgs)
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

func generateUserArgsForCreateUser(user *TempUser, passwordHasher password_hasher.PasswordHasher) (CreateUserArgs, error) {
	var loginMethod LoginMethod
	var accessKey, hashedPass, passSalt string
	var supportPassword bool

	user.LoginMethod.Fold(
		func() {
			supportPassword = true
			accessKey = user.Email
		}, func() {
			supportPassword = true
			accessKey = user.Phone.ToAppStanderdForm()
		},
	)

	if supportPassword {
		var err error
		hashedPass, passSalt, err = passwordHasher.GeneratePasswordHashWithSalt((user.Password))
		if err != nil {
			return CreateUserArgs{}, nil
		}
	}

	createUserArgs := CreateUserArgs{
		Username:    user.Username,
		Fname:       user.Fname,
		Lname:       user.Lname,
		LoginMethod: loginMethod,
		AccessKey:   accessKey,
		HashedPass:  hashedPass,
		PassSalt:    passSalt,
	}

	return createUserArgs, nil
}

func (repo repositoryImpl) deleteTempUserFromCache(ctx context.Context, tUser *TempUser) {
	zlog := zerolog.Ctx(ctx)

	// ignore any error because the temp user will be auto cleand by redis after sometime
	if err := repo.dataSource.DeleteUserFromTempCache(ctx, tUser.Id); err != nil {
		zlog.Err(err).Msg("error while deleting user form temp cache. igonoring this error")
	}
}

func (repo repositoryImpl) PasswordLogin(ctx context.Context, passwordLoginAccessKey PasswordLoginAccessKey, password string, loginMethod LoginMethod, installation database.Installation) (user User, token string, err error) {
	zlog := zerolog.Ctx(ctx)

	var accessKey string

	loginMethod.Fold(
		func() {
			accessKey = passwordLoginAccessKey.Email
		}, func() {
			accessKey = passwordLoginAccessKey.Phone.ToAppStanderdForm()
		},
	)

	userWithLoginOption, err := repo.dataSource.GetActiveLoginOptionWithUser(ctx, accessKey, loginMethod)
	if err != nil {
		if errors.Is(err, apperr.ErrNoResult) {
			err = apperr.ErrInvalidLoginCredentials
		} else {
			zlog.Err(err).Msg("error geting active login option with user data")
		}
		return User{}, "", err
	}

	checkPassword := func() error {
		hashedPassword := userWithLoginOption.LoginOptionHashedPass.String
		salt := userWithLoginOption.LoginOptionPassSalt.String
		if ok, err := repo.PasswordHasher.CompareHashAndPassword(hashedPassword, salt, password); !ok || err != nil {
			if err != nil {
				return err
			}
			return apperr.ErrInvalidLoginCredentials
		}
		return nil
	}

	switch loginMethod {
	case LoginMethodPhoneNumber, LoginMethodEmail:
		err := checkPassword()
		if err != nil {
			zlog.Err(err).Msg("error while checking the password for user to login")
			return User{}, "", err
		}

	default:
		panic(fmt.Sprintf("Not supported login method %s", loginMethod.String()))
	}

	expiresAt := time.Now().AddDate(0, 6, 0) // after 6 months from now
	token, err = appjwt.NewAppJWT().GenWithClaims(expiresAt, int(userWithLoginOption.UserID), "auth")
	if err != nil {
		zlog.Err(err).Msg("error while generating a new session token using jwt, for login")
		return User{}, "", err
	}

	err = repo.dataSource.CreateNewSession(ctx, userWithLoginOption.LoginOptionID, installation.ID, token, expiresAt)
	if err != nil {
		zlog.Err(err).Msg("error creating new session for user to login")
		return User{}, "", err
	}

	user = User{
		ID:           userWithLoginOption.UserID,
		Username:     userWithLoginOption.UserUsername,
		ProfileImage: userWithLoginOption.UserProfileImage,
		FirstName:    userWithLoginOption.UserFirstName,
		MiddleName:   userWithLoginOption.UserMiddleName,
		LastName:     userWithLoginOption.UserLastName,
		RoleID:       userWithLoginOption.UserRoleID,
		DeletedAt:    userWithLoginOption.UserDeletedAt,
		BlockedAt:    userWithLoginOption.UserBlockedAt,
		CreatedAt:    userWithLoginOption.UserCreatedAt,
		UpdatedAt:    userWithLoginOption.UserUpdatedAt,
	}

	return user, token, nil
}

func (repo repositoryImpl) GetInstallationUsingUuid(ctx context.Context, InstallationId uuid.UUID, attachedToUserId *int32) (installation database.Installation, err error) {
	zlog := zerolog.Ctx(ctx)

	if attachedToUserId == nil {
		installation, err = repo.dataSource.GetInstallationUsingUUID(ctx, InstallationId)
	} else {
		installation, err = repo.dataSource.GetInstallationUsingUUIdAndWhereAttachTo(ctx, InstallationId, *attachedToUserId)
	}

	if err != nil {
		if !errors.Is(err, apperr.ErrNoResult) {
			zlog.Err(err).Msg("error geting an installation from the database")
		}
		return database.Installation{}, err
	}

	return installation, nil
}
