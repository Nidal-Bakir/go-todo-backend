package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/otp"
	"github.com/Nidal-Bakir/go-todo-backend/internal/gateway"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/password_hasher"
	phonenumber "github.com/Nidal-Bakir/go-todo-backend/internal/utils/phone_number"
	usernaemgen "github.com/Nidal-Bakir/username_r_gen"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
)

const (
	OtpCodeLength             = 6
	PasswordRecommendedLength = 8
)

type Repository interface {
	GetUserById(ctx context.Context, id int) (User, error)
	GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (UserAndSession, error)
	CreateTempUser(ctx context.Context, tUser *TempUser) (*TempUser, error)
	CreateUser(ctx context.Context, tempUserId uuid.UUID, otp string) (User, error)
	PasswordLogin(ctx context.Context, accessKey PasswordLoginAccessKey, password string, installation database.Installation) (user User, token string, err error)
	GetInstallationUsingToken(ctx context.Context, installationToken string, attachedToSessionId *int32) (database.Installation, error)
	ChangePasswordForAllLoginOptions(ctx context.Context, userID int, oldPassword, newPassword string) error
	VerifyAuthToken(token string) (*AuthClaims, error)
	VerifyTokenForInstallation(token string) (*InstallationClaims, error)
	CreateInstallation(ctx context.Context, data CreateInstallationData) (installationToken string, err error)
	UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error
	Logout(ctx context.Context, userId, installationId, tokenId int, terminateAllOtherSessions bool) error
	ForgetPassword(ctx context.Context, accessKey PasswordLoginAccessKey) (uuid.UUID, error)
	ResetPassword(ctx context.Context, id uuid.UUID, providedOTP, newPassword string) error
	GetAllActiveLoginOptionForUser(ctx context.Context, userId int) ([]PublicLoginOptionForProfile, error)
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

	sentOtp, err := sendOtpToTempUserForAccountVerification(
		ctx,
		repo.gatewaysProvider,
		PasswordLoginAccessKey{LoginMethod: tUser.LoginMethod, Phone: tUser.Phone, Email: tUser.Email},
	)
	if err != nil {
		zlog.Err(err).Msg("error sending otp to temp user, for create account")
		return tUser, err
	}
	tUser.SentOTP = sentOtp

	err = repo.dataSource.StoreUserInTempCache(ctx, *tUser)
	if err != nil {
		zlog.Err(err).Msg("error creating temp user in the cache database")
	}

	return tUser, err
}

func (repo repositoryImpl) isUsedCredentials(ctx context.Context, tUser TempUser) (bool, error) {
	var accessKey string
	tUser.LoginMethod.Fold(
		func() { accessKey = tUser.Email },
		func() { accessKey = tUser.Phone.ToAppStdForm() },
	)
	return repo.isAccessKeyUsedInAnyLoginOption(ctx, accessKey)
}

func (repo repositoryImpl) isAccessKeyUsedInAnyLoginOption(ctx context.Context, accessKey string) (bool, error) {
	return repo.dataSource.IsAccessKeyUsedInAnyLoginOption(ctx, accessKey)
}

func sendOtpToTempUserForAccountVerification(
	ctx context.Context,
	gatewaysProvider gateway.Provider,
	passwordLoginAccessKey PasswordLoginAccessKey,
) (sentOTP string, err error) {
	otpSender := otp.NewOTPSender(ctx, gatewaysProvider, OtpCodeLength)

	passwordLoginAccessKey.LoginMethod.Fold(
		func() {
			sentOTP, err = otpSender.SendEmailOtpForAccountVerification(ctx, passwordLoginAccessKey.Email)
		},
		func() {
			sentOTP, err = otpSender.SendSmsOtpForAccountVerification(ctx, passwordLoginAccessKey.Phone)
		},
	)
	return sentOTP, err
}

func sendOtpToTempUserForForgetPassword(
	ctx context.Context,
	gatewaysProvider gateway.Provider,
	passwordLoginAccessKey PasswordLoginAccessKey,
) (sentOTP string, err error) {
	otpSender := otp.NewOTPSender(ctx, gatewaysProvider, OtpCodeLength)

	passwordLoginAccessKey.LoginMethod.Fold(
		func() {
			sentOTP, err = otpSender.SendEmailOtpForForgetPassword(ctx, passwordLoginAccessKey.Email)
		},
		func() {
			sentOTP, err = otpSender.SendSmsOtpForForgetPassword(ctx, passwordLoginAccessKey.Phone)
		},
	)
	return sentOTP, err
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
			accessKey = user.Phone.ToAppStdForm()
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

func (repo repositoryImpl) PasswordLogin(
	ctx context.Context,
	passwordLoginAccessKey PasswordLoginAccessKey,
	password string,
	installation database.Installation,
) (user User, token string, err error) {
	zlog := zerolog.Ctx(ctx)

	userWithLoginOption, err := repo.dataSource.GetActiveLoginOptionWithUser(
		ctx,
		passwordLoginAccessKey.accessKeyStr(),
		passwordLoginAccessKey.LoginMethod,
	)
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
	err = checkPassword()
	if err != nil {
		zlog.Err(err).Msg("error while checking the password for user to login")
		return User{}, "", err
	}

	expiresAt := time.Now().AddDate(0, 6, 0) // after 6 months from now
	token, err = repo.authJWT.GenWithClaimsForUser(userWithLoginOption.UserID, expiresAt)
	if err != nil {
		zlog.Err(err).Msg("error while generating a new session token using jwt, for login")
		return User{}, "", err
	}

	err = repo.dataSource.CreateNewSessionAndAttachUserToInstallation(ctx, userWithLoginOption.LoginOptionID, installation.ID, token, expiresAt)
	if err != nil {
		if !apperr.IsAppErr(err) {
			zlog.Err(err).Msg("error creating new session for user to login")
		}
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
	return repo.changePasswordForAllLoginOptions(ctx, userID, oldPassword, newPassword, true)
}

func (repo repositoryImpl) changePasswordForAllLoginOptions(ctx context.Context, userID int, oldPassword, newPassword string, shouldCheckOldPasswordWithCurrentOne bool) error {
	zlog := zerolog.Ctx(ctx)

	loginOptions, err := repo.dataSource.GetAllActiveLoginOptionForUserAndSupportPassword(ctx, int32(userID))

	if err != nil {
		zlog.Err(err).Msg("error while getting all the login options for a user")
		return err
	}

	if shouldCheckOldPasswordWithCurrentOne {
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
	zlog := zerolog.Ctx(ctx)

	expiresAt := time.Now().AddDate(0, 6, 0) // after 6 months from now
	token, err := repo.authJWT.GenWithClaimsForInstallation(expiresAt)
	if err != nil {
		zlog.Err(err).Msg("error while gen jwt token with claims for installation")
		return "", err
	}

	err = repo.dataSource.CreateInstallation(ctx, data, token)
	if err != nil {
		zlog.Err(err).Msg("error while creating installation")
		return "", err
	}

	return token, nil
}

func (repo repositoryImpl) UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error {
	return repo.dataSource.UpdateInstallation(ctx, installationToken, data)
}

func (repo repositoryImpl) Logout(ctx context.Context, userId, installationId, tokenId int, terminateAllOtherSessions bool) error {
	zlog := zerolog.Ctx(ctx).With().Bool("terminate_all_other_sessions", terminateAllOtherSessions).Logger()

	var err error
	if terminateAllOtherSessions {
		err = repo.dataSource.ExpAllTokensAndUnlinkThemFromInstallation(ctx, userId)
	} else {
		err = repo.dataSource.ExpTokenAndUnlinkFromInstallation(ctx, installationId, tokenId)
	}
	if err != nil {
		zlog.Err(err).Msg("error while loging out the user")
	}

	return err
}

func (repo repositoryImpl) ForgetPassword(ctx context.Context, accessKey PasswordLoginAccessKey) (uuid.UUID, error) {
	zlog := zerolog.Ctx(ctx)

	randomUUID := uuid.New()

	loginOption, err := repo.dataSource.GetActiveLoginOption(ctx, accessKey.accessKeyStr(), accessKey.LoginMethod)
	if err != nil {
		if errors.Is(err, apperr.ErrNoResult) {
			// send a random uuid and do not report that the user/accessKey is not present in the database.
			// Security in ambiguity
			err = nil
		} else {
			zlog.Err(err).Msg("error geting the login option, for forget password")
		}
		return randomUUID, err
	}

	sentOtp, err := sendOtpToTempUserForForgetPassword(
		ctx,
		repo.gatewaysProvider,
		PasswordLoginAccessKey{LoginMethod: accessKey.LoginMethod, Phone: accessKey.Phone, Email: accessKey.Email},
	)
	if err != nil {
		zlog.Err(err).Msg("error sending otp to user, for forget password")
		return randomUUID, err
	}

	forgetPassData := ForgetPasswordTmpDataStore{
		Id:            randomUUID,
		LoginOptionId: int(loginOption.ID),
		UserId:        int(loginOption.UserID),
		SentOTP:       sentOtp,
	}

	err = repo.dataSource.StoreForgetPasswordDataInTempCache(ctx, forgetPassData)
	if err != nil {
		zlog.Err(err).Msg("error can not store forget password data in the temp cache")
		return randomUUID, err
	}

	return randomUUID, nil
}

func (repo repositoryImpl) ResetPassword(ctx context.Context, id uuid.UUID, providedOTP, newPassword string) error {
	zlog := zerolog.Ctx(ctx)

	forgetPassData, err := repo.dataSource.GetForgetPasswordDataFromTempCache(ctx, id)
	if err != nil {
		if errors.Is(err, apperr.ErrNoResult) {
			return apperr.ErrInvalidId
		}
		zlog.Err(err).Msg("error can not get the forget password data from temp cache")
		return err
	}

	if forgetPassData.SentOTP != providedOTP {
		return apperr.ErrInvalidOtpCode
	}

	err = repo.changePasswordForAllLoginOptions(ctx, forgetPassData.UserId, "", newPassword, false)
	if err != nil {
		zlog.Err(err).Msg("error can not update the password for forget password flow")
		return err
	}

	repo.deleteForgetPasswordDataFromTempCache(ctx, forgetPassData)

	return nil
}

func (repo repositoryImpl) deleteForgetPasswordDataFromTempCache(ctx context.Context, forgetPassData *ForgetPasswordTmpDataStore) {
	zlog := zerolog.Ctx(ctx)
	// ignore any error because the temp user will be auto cleand by redis after sometime
	if err := repo.dataSource.DeleteForgetPasswordDataFromTempCache(ctx, forgetPassData.Id); err != nil {
		zlog.Err(err).Msg("error while deleting forget password temp data form temp cache. igonoring this error")
	}
}

func (repo repositoryImpl) GetAllActiveLoginOptionForUser(ctx context.Context, userId int) ([]PublicLoginOptionForProfile, error) {
	zlog := zerolog.Ctx(ctx)

	res, err := repo.dataSource.GetAllActiveLoginOptionForUser(ctx, int32(userId))
	if err != nil {
		zlog.Err(err).Msg("error while getting all the active login option for user")
		return nil, err
	}

	loginOptionSlice := make([]PublicLoginOptionForProfile, len(res))

	for i, v := range res {
		method, err := new(LoginMethod).FromString(v.LoginMethod)
		if err != nil {
			zlog.Err(err).Msg("error can not extract the login method from the str")
			return []PublicLoginOptionForProfile{}, err
		}

		var email string
		var phone phonenumber.PhoneNumber
		method.Fold(
			func() { email = v.AccessKey },
			func() {
				phone, err = phonenumber.NewPhoneNumberFromStdForm(v.AccessKey)
			},
		)
		if err != nil {
			zlog.Err(err).Msg("error can not extract the phone number")
			return []PublicLoginOptionForProfile{}, err
		}

		loginOptionSlice[i] = PublicLoginOptionForProfile{
			ID:          v.ID,
			Email:       email,
			Phone:       phone,
			LoginMethod: *method,
			IsVerified:  v.VerifiedAt.Valid,
		}
	}

	return loginOptionSlice, nil
}
