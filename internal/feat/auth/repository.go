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
	CreateTempPasswordUser(ctx context.Context, tUser *TempPasswordUser) (*TempPasswordUser, error)
	CreatePasswordUser(ctx context.Context, tempUserId uuid.UUID, otp string) (User, error)
	PasswordLogin(ctx context.Context, accessKey PasswordLoginAccessKey, password, ipAddress string, installation database.Installation) (user User, token string, err error)
	GetInstallationUsingToken(ctx context.Context, installationToken string, attachedToSessionId *int32) (database.Installation, error)
	ChangePasswordForAllPasswordLoginIdentities(ctx context.Context, userID int, oldPassword, newPassword string) error
	VerifyAuthToken(token string) (*AuthClaims, error)
	VerifyTokenForInstallation(token string) (*InstallationClaims, error)
	CreateInstallation(ctx context.Context, data CreateInstallationData) (installationToken string, err error)
	UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error
	Logout(ctx context.Context, userId, installationId, tokenId int, terminateAllOtherSessions bool) error
	ForgetPassword(ctx context.Context, accessKey PasswordLoginAccessKey) (uuid.UUID, error)
	ResetPassword(ctx context.Context, id uuid.UUID, providedOTP, newPassword string) error
	GetAllLoginIdentitiesForUser(ctx context.Context, userId int) ([]PublicLoginOptionForProfile, error)
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

	user := User{
		ID:           dbUser.ID,
		Username:     dbUser.Username,
		ProfileImage: dbUser.ProfileImage,
		FirstName:    dbUser.FirstName,
		MiddleName:   dbUser.MiddleName,
		LastName:     dbUser.LastName,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		BlockedAt:    dbUser.BlockedAt,
		RoleID:       dbUser.RoleID,
	}
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

func (repo repositoryImpl) CreateTempPasswordUser(ctx context.Context, tUser *TempPasswordUser) (*TempPasswordUser, error) {
	zlog := zerolog.Ctx(ctx)

	tUser.Id = uuid.New()
	tUser.Username = tUser.Id.String()

	if ok := tUser.ValidateForStore(); !ok {
		return tUser, apperr.ErrInvalidTempUserdata
	}

	// check if the user is already present in the database with this Credentials
	if err := repo.isUsedCredentialsPasswordUser(ctx, *tUser); err != nil {
		return tUser, err
	}

	sentOtp, err := sendOtpToTempUserForAccountVerification(
		ctx,
		repo.gatewaysProvider,
		PasswordLoginAccessKey{LoginIdentityType: tUser.LoginIdentityType, Phone: tUser.Phone, Email: tUser.Email},
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

func (repo repositoryImpl) isUsedCredentialsPasswordUser(ctx context.Context, tUser TempPasswordUser) error {
	var resultError error

	tUser.LoginIdentityType.Fold(
		LoginIdentityFoldActions{
			OnEmail: func() {
				isUsed, err := repo.dataSource.IsEmailUsedInPasswordLoginIdentity(ctx, tUser.Email)
				if err != nil {
					resultError = err
					return
				}
				if isUsed {
					resultError = apperr.ErrAlreadyUsedEmail
					return
				}

				isUsed, err = repo.dataSource.IsEmailUsedInOidcLoginIdentity(ctx, tUser.Email)
				if err != nil {
					resultError = err
					return
				}
				if isUsed {
					resultError = apperr.ErrAlreadyUsedEmailWithOidc
					return
				}
			},
			OnPhone: func() {
				isUsed, err := repo.dataSource.IsPhoneUsedInPasswordLoginIdentity(ctx, tUser.Email)
				if err != nil {
					resultError = err
					return
				}
				if isUsed {
					resultError = apperr.ErrAlreadyUsedPhoneNumber
					return
				}
			},
		},
	)

	return resultError
}

func sendOtpToTempUserForAccountVerification(
	ctx context.Context,
	gatewaysProvider gateway.Provider,
	passwordLoginAccessKey PasswordLoginAccessKey,
) (sentOTP string, err error) {
	otpSender := otp.NewOTPSender(ctx, gatewaysProvider, OtpCodeLength)

	passwordLoginAccessKey.LoginIdentityType.Fold(
		LoginIdentityFoldActions{
			OnEmail: func() { sentOTP, err = otpSender.SendEmailOtpForAccountVerification(ctx, passwordLoginAccessKey.Email) },
			OnPhone: func() { sentOTP, err = otpSender.SendSmsOtpForAccountVerification(ctx, passwordLoginAccessKey.Phone) },
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

	passwordLoginAccessKey.LoginIdentityType.Fold(
		LoginIdentityFoldActions{
			OnEmail: func() { sentOTP, err = otpSender.SendEmailOtpForForgetPassword(ctx, passwordLoginAccessKey.Email) },
			OnPhone: func() { sentOTP, err = otpSender.SendSmsOtpForForgetPassword(ctx, passwordLoginAccessKey.Phone) },
		},
	)
	return sentOTP, err
}

func (repo repositoryImpl) CreatePasswordUser(ctx context.Context, tempUserId uuid.UUID, providedOTP string) (User, error) {
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

	user, err := repo.storPasswordUser(ctx, tUser)
	if err != nil {
		return User{}, err
	}

	repo.deleteTempUserFromCache(ctx, tUser)

	return user, err
}

func (repo repositoryImpl) getTempUser(ctx context.Context, id uuid.UUID) (*TempPasswordUser, error) {
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

func checkTempUserOTP(tUser *TempPasswordUser, providedOTP string) error {
	if tUser.SentOTP != providedOTP {
		return apperr.ErrInvalidOtpCode
	}
	return nil
}

func (repo repositoryImpl) storPasswordUser(ctx context.Context, tUser *TempPasswordUser) (User, error) {
	zlog := zerolog.Ctx(ctx)

	if ok := tUser.ValidateForStore(); !ok {
		return User{}, apperr.ErrInvalidTempUserdata
	}

	createUserArgs, err := generatePassworUserArgsForCreateUser(tUser, repo.passwordHasher)
	if err != nil {
		zlog.Err(err).Msg("error generating user args")
		return User{}, err
	}

	dbUser, err := repo.dataSource.CreatePasswordUser(ctx, createUserArgs)
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

func generatePassworUserArgsForCreateUser(user *TempPasswordUser, passwordHasher password_hasher.PasswordHasher) (CreatePasswordUserArgs, error) {
	var phone, email string

	user.LoginIdentityType.Fold(
		LoginIdentityFoldActions{
			OnEmail: func() { email = user.Email },
			OnPhone: func() { phone = user.Phone.ToAppStdForm() },
		},
	)

	hashedPass, passSalt, err := passwordHasher.GeneratePasswordHashWithSalt((user.Password))
	if err != nil {
		return CreatePasswordUserArgs{}, nil
	}

	createUserArgs := CreatePasswordUserArgs{
		Username:          user.Username,
		Fname:             user.Fname,
		Lname:             user.Lname,
		LoginIdentityType: user.LoginIdentityType,
		Email:             email,
		Phone:             phone,
		HashedPass:        hashedPass,
		PassSalt:          passSalt,
	}

	return createUserArgs, nil
}

func (repo repositoryImpl) deleteTempUserFromCache(ctx context.Context, tUser *TempPasswordUser) {
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
	ipAddress string,
	installation database.Installation,
) (user User, token string, err error) {
	zlog := zerolog.Ctx(ctx)

	userWithLoginOption, err := repo.dataSource.GetPasswordLoginIdentityWithUser(
		ctx,
		passwordLoginAccessKey.accessKeyStr(),
		passwordLoginAccessKey.LoginIdentityType,
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
		hashedPassword := userWithLoginOption.HashedPass
		salt := userWithLoginOption.PassSalt
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

	err = repo.dataSource.CreateNewSessionAndAttachUserToInstallation(ctx, userWithLoginOption.LoginIdentityID, installation.ID, token, ipAddress, expiresAt)
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
		BlockedAt:    userWithLoginOption.UserBlockedAt,
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

func (repo repositoryImpl) ChangePasswordForAllPasswordLoginIdentities(ctx context.Context, userID int, oldPassword, newPassword string) error {
	return repo.changePasswordForAllPasswordLoginIdentities(ctx, userID, oldPassword, newPassword, true)
}

func (repo repositoryImpl) changePasswordForAllPasswordLoginIdentities(ctx context.Context, userID int, oldPassword, newPassword string, shouldCheckOldPasswordWithCurrentOne bool) error {
	zlog := zerolog.Ctx(ctx)

	loginOptions, err := repo.dataSource.GetAllPasswordLoginIdentitiesForUser(ctx, int32(userID))

	if err != nil {
		zlog.Err(err).Msg("error while getting all the login options for a user")
		return err
	}

	if shouldCheckOldPasswordWithCurrentOne {
		// all the login options should have the same password
		for _, op := range loginOptions {
			ok, err := repo.passwordHasher.CompareHashAndPassword(op.PasswordHashedPass.String, op.PasswordPassSalt.String, oldPassword)
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

	err = repo.dataSource.ChangePasswordLoginIdentityForUser(ctx, int32(userID), hashedPass, salt)
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

	loginOption, err := repo.dataSource.GetPasswordLoginIdentity(ctx, accessKey.accessKeyStr(), accessKey.LoginIdentityType)
	if err != nil {
		if errors.Is(err, apperr.ErrNoResult) {
			// send a random uuid and do not report that the user/accessKey is not present in the database.
			// Security by obscurity
			err = nil
		} else {
			zlog.Err(err).Msg("error geting the login option, for forget password")
		}
		return randomUUID, err
	}

	sentOtp, err := sendOtpToTempUserForForgetPassword(
		ctx,
		repo.gatewaysProvider,
		PasswordLoginAccessKey{LoginIdentityType: accessKey.LoginIdentityType, Phone: accessKey.Phone, Email: accessKey.Email},
	)
	if err != nil {
		zlog.Err(err).Msg("error sending otp to user, for forget password")
		return randomUUID, err
	}

	forgetPassData := ForgetPasswordTmpDataStore{
		Id:      randomUUID,
		UserId:  int(loginOption.UserID),
		SentOTP: sentOtp,
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

	err = repo.changePasswordForAllPasswordLoginIdentities(ctx, forgetPassData.UserId, "", newPassword, false)
	if err != nil {
		zlog.Err(err).Msg("error can not update the password for forget password flow")
		return err
	}

	repo.deleteForgetPasswordDataFromTempCache(ctx, forgetPassData)

	// logout all the devices, do not returen any erros, jsut log them
	err = repo.dataSource.ExpAllTokensAndUnlinkThemFromInstallation(ctx, forgetPassData.UserId)
	if err != nil {
		zlog.Err(err).Msg("error can not exp all the tokens and unlink them from installation after a Reset Passowrd operation")
	}

	return nil
}

func (repo repositoryImpl) deleteForgetPasswordDataFromTempCache(ctx context.Context, forgetPassData *ForgetPasswordTmpDataStore) {
	zlog := zerolog.Ctx(ctx)
	// ignore any error because the temp user will be auto cleand by redis after sometime
	if err := repo.dataSource.DeleteForgetPasswordDataFromTempCache(ctx, forgetPassData.Id); err != nil {
		zlog.Err(err).Msg("error while deleting forget password temp data form temp cache. igonoring this error")
	}
}

func (repo repositoryImpl) GetAllLoginIdentitiesForUser(ctx context.Context, userId int) ([]PublicLoginOptionForProfile, error) {
	zlog := zerolog.Ctx(ctx)

	res, err := repo.dataSource.GetAllLoginIdentitiesForUser(ctx, int32(userId))
	if err != nil {
		zlog.Err(err).Msg("error while getting all the active login option for user")
		return nil, err
	}

	loginOptionSlice := make([]PublicLoginOptionForProfile, len(res))

	for i, v := range res {
		identityType, err := new(LoginIdentityType).FromString(v.LoginIdentityIdentityType)
		if err != nil {
			zlog.Err(err).Msg("error can not extract the login method from the str")
			return []PublicLoginOptionForProfile{}, err
		}

		var email string
		var phone phonenumber.PhoneNumber
		identityType.Fold(
			LoginIdentityFoldActions{
				OnEmail: func() { email = v.PasswordEmail.String },
				OnPhone: func() { phone, err = phonenumber.NewPhoneNumberFromStdForm(v.PasswordPhone.String) },
				OnOcid:  func() { email = v.OidcDataEmail.String },
			},
		)
		if err != nil {
			zlog.Err(err).Msg("error can not extract the phone number")
			return []PublicLoginOptionForProfile{}, err
		}

		loginOptionSlice[i] = PublicLoginOptionForProfile{
			LoginIdentityType: *identityType,
			ID:                v.LoginIdentityID,
			Email:             email,
			Phone:             phone,
			IsVerified:        v.PasswordVerifiedAt.Valid,
			OidcProvider:      v.OauthProviderName.String,
		}
	}

	return loginOptionSlice, nil
}
