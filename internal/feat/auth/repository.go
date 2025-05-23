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

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/password_hasher"
	usernaemgen "github.com/Nidal-Bakir/username_r_gen"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
)

const (
	OtpCodeLength             = 6
	PasswordRecommendedLength = 8
)

type PasswordLoginAccessKey struct {
	Phone utils.PhoneNumber
	Email string
}

type CreateInstallationData struct {
	NotificationToken       string // e.g the FCM token
	Locale                  string // e.g: en-US ...
	TimezoneOffsetInMinutes int    // e.g: +180
	DeviceManufacturer      string // e.g: samsung
	DeviceOS                string // e.g: android
	DeviceOSVersion         string // e.g: 14
	AppVersion              string // e.g: 3.1.1
}

type UpdateInstallationData struct {
	NotificationToken       string // e.g the FCM token
	Locale                  string // e.g: en-US ...
	TimezoneOffsetInMinutes int    // e.g: +180
	AppVersion              string // e.g: 3.1.1
}

type Repository interface {
	GetUserById(ctx context.Context, id int) (User, error)
	// GetUserBySessionToken(ctx context.Context, sessionToken string) (User, error)
	GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (UserAndSession, error)
	CreateTempUser(ctx context.Context, tUser *TempUser) (*TempUser, error)
	CreateUser(ctx context.Context, tempUserId uuid.UUID, otp string) (User, error)
	PasswordLogin(ctx context.Context, accessKey PasswordLoginAccessKey, password string, loginMethod LoginMethod, installation database.Installation) (user User, token string, err error)
	GetInstallationUsingToken(ctx context.Context, installationToken string, attachedToSessionId *int32) (database.Installation, error)
	ChangePasswordForAllLoginOptions(ctx context.Context, userID int, oldPassword, newPassword string) error
	VerifyAuthToken(token string) (*AuthClaims, error)
	VerifyTokenForInstallation(token string) (*InstallationClaims, error)
	CreateInstallation(ctx context.Context, data CreateInstallationData) (installationToken string, err error)
	UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error
	Logout(ctx context.Context, userId, installationId int, token string, terminateAllOtherSessions bool) error
}

func NewRepository(ds DataSource, gatewaysProvider gateway.Provider, passwordHasher password_hasher.PasswordHasher, authJWT *AuthJWT) Repository {
	return repositoryImpl{dataSource: ds, gatewaysProvider: gatewaysProvider, passwordHasher: passwordHasher, authJWT: authJWT}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	dataSource       DataSource
	gatewaysProvider gateway.Provider
	passwordHasher   password_hasher.PasswordHasher
	authJWT          *AuthJWT
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

func (repo repositoryImpl) GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (UserAndSession, error) {
	zlog := zerolog.Ctx(ctx)

	userAndSessionDataFromDB, err := repo.dataSource.GetUserAndSessionDataBySessionToken(ctx, sessionToken)
	if err != nil {
		if !errors.Is(err, apperr.ErrNoResult) {
			zlog.Err(err).Msg("error geting the user by session token")
		}
		return UserAndSession{}, err
	}

	userAndSession := NewUserAndSessionFromDatabaseUserAndSessionRow(userAndSessionDataFromDB)
	return userAndSession, nil
}

func (repo repositoryImpl) CreateTempUser(ctx context.Context, tUser *TempUser) (*TempUser, error) {
	zlog := zerolog.Ctx(ctx)

	tUser.Id = uuid.New()
	tUser.Username = tUser.Id.String()

	if ok := tUser.ValidateForStore(); !ok {
		return tUser, apperr.ErrInvalidTempUserdata
	}

	// check if the user is already present in the database with this Credentials
	if isUsed, err := repo.isUsedCredentials(ctx, *tUser); isUsed || err != nil {
		if err != nil {
			return tUser, err
		}

		tUser.LoginMethod.Fold(
			func() { err = apperr.ErrAlreadyUsedEmail },
			func() { err = apperr.ErrAlreadyUsedPhoneNumber },
		)
		return tUser, err
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

func (repo repositoryImpl) isUsedCredentials(ctx context.Context, tUser TempUser) (bool, error) {
	var accessKey string
	tUser.LoginMethod.Fold(
		func() { accessKey = tUser.Email },
		func() { accessKey = tUser.Phone.ToAppStanderdForm() },
	)
	return repo.isAccessKeyUsedInAnyLoginOption(ctx, accessKey)
}

func (repo repositoryImpl) isAccessKeyUsedInAnyLoginOption(ctx context.Context, accessKey string) (bool, error) {
	return repo.dataSource.IsAccessKeyUsedInAnyLoginOption(ctx, accessKey)
}

func sendOtpToTempUser(ctx context.Context, gatewaysProvider gateway.Provider, tUser TempUser) (sentOTP string, err error) {
	otpSender := otp.NewOTPSender(gatewaysProvider, OtpCodeLength)

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
		if errors.Is(err, apperr.ErrNoResult) {
			return User{}, apperr.ErrInvalidId
		}
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

	createUserArgs, err := generateUserArgsForCreateUser(tUser, repo.passwordHasher)
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
		LoginMethod: user.LoginMethod,
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
		func() { accessKey = passwordLoginAccessKey.Email },
		func() { accessKey = passwordLoginAccessKey.Phone.ToAppStanderdForm() },
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
		if ok, err := repo.passwordHasher.CompareHashAndPassword(hashedPassword, salt, password); !ok || err != nil {
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
	token, err = repo.authJWT.GenWithClaimsForUser(userWithLoginOption.UserID, expiresAt)
	if err != nil {
		zlog.Err(err).Msg("error while generating a new session token using jwt, for login")
		return User{}, "", err
	}

	err = repo.dataSource.CreateNewSessionAndAttachUserToInstallation(ctx, userWithLoginOption.LoginOptionID, installation.ID, token, expiresAt)
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

func (repo repositoryImpl) GetInstallationUsingToken(ctx context.Context, installationToken string, attachedToSessionId *int32) (installation database.Installation, err error) {
	zlog := zerolog.Ctx(ctx)

	if attachedToSessionId == nil {
		installation, err = repo.dataSource.GetInstallationUsingToken(ctx, installationToken)
	} else {
		installation, err = repo.dataSource.GetInstallationUsingTokenAndWhereAttachTo(ctx, installationToken, *attachedToSessionId)
	}

	if err != nil {
		if !errors.Is(err, apperr.ErrNoResult) {
			zlog.Err(err).Msg("error geting an installation from the database")
		}
		return database.Installation{}, err
	}

	return installation, nil
}

func (repo repositoryImpl) ChangePasswordForAllLoginOptions(ctx context.Context, userID int, oldPassword, newPassword string) error {
	zlog := zerolog.Ctx(ctx)

	loginOptions, err := repo.dataSource.GetAllActiveLoginOptionByUserIdAndSupportPassword(ctx, int32(userID))

	if err != nil {
		zlog.Err(err).Msg("error while getting all the login options for a user")
		return err
	}

	// all the login options should have the same password
	for _, op := range loginOptions {
		ok, err := repo.passwordHasher.CompareHashAndPassword(op.HashedPass.String, op.PassSalt.String, oldPassword)
		if err != nil {
			zlog.Err(err).Msg("error while comparing password hash with salt and password to change a password for logged in user")
			return err
		}
		if !ok {
			return apperr.ErrOldPasswordDoesNotMatchCurrentOne
		}
	}

	hashedPass, salt, err := repo.passwordHasher.GeneratePasswordHashWithSalt(newPassword)
	if err != nil {
		zlog.Err(err).Msg("error while generating password hash with salt to change a password for logged in user")
		return err
	}

	err = repo.dataSource.ChangeAllPasswordsForLoginOptions(ctx, int32(userID), hashedPass, salt)
	if err != nil {
		zlog.Err(err).Msg("error while changing the password for login options to logged in user")
		return err
	}

	return nil
}

func (repo repositoryImpl) VerifyAuthToken(token string) (*AuthClaims, error) {
	return repo.authJWT.VerifyTokenForUser(token)
}

func (repo repositoryImpl) VerifyTokenForInstallation(token string) (*InstallationClaims, error) {
	return repo.authJWT.VerifyTokenForInstallation(token)
}

func (repo repositoryImpl) CreateInstallation(ctx context.Context, data CreateInstallationData) (installationToken string, err error) {
	expiresAt := time.Now().AddDate(0, 6, 0) // after 6 months from now
	token, err := repo.authJWT.GenWithClaimsForInstallation(expiresAt)
	if err != nil {
		return "", err
	}

	err = repo.dataSource.CreateInstallation(ctx, data, token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (repo repositoryImpl) UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error {
	return repo.dataSource.UpdateInstallation(ctx, installationToken, data)
}

func (repo repositoryImpl) Logout(ctx context.Context, userId, installationId int, token string, terminateAllOtherSessions bool) error {
	// repo.dataSource.
	// TODO: implent this
	panic("not implemented")
}
