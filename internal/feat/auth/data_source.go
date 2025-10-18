package auth

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database/database_queries"
	dbutils "github.com/Nidal-Bakir/go-todo-backend/internal/utils/db_utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	expirationForTempUser               = time.Minute * 30
	expirationForForgetPasswordTempData = time.Minute * 15
)

type DataSource interface {
	// Query ---

	GetUserById(ctx context.Context, id int32) (database_queries.UsersGetUserByIdRow, error)

	GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (database_queries.UsersGetUserAndSessionDataBySessionTokenRow, error)

	GetUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) (*TempPasswordUser, error)
	GetForgetPasswordDataFromTempCache(ctx context.Context, dataId uuid.UUID) (*ForgetPasswordTmpDataStore, error)

	GetInstallationUsingTokenAndWhereAttachTo(ctx context.Context, installationToken string, attachedToSession int32) (database_queries.Installation, error)
	GetInstallationUsingToken(ctx context.Context, installationToken string) (database_queries.Installation, error)

	GetPasswordLoginIdentityWithUser(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database_queries.LoginIdentityGetPasswordLoginIdentityWithUserRow, error)
	GetPasswordLoginIdentity(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database_queries.LoginIdentityGetPasswordLoginIdentityRow, error)
	GetAllPasswordLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database_queries.LoginIdentityGetAllPasswordLoginIdentitiesByUserIdRow, error)
	GetAllLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database_queries.LoginIdentityGetAllByUserIdRow, error)

	IsEmailUsedInPasswordLoginIdentity(ctx context.Context, email string) (bool, error)
	IsPhoneUsedInPasswordLoginIdentity(ctx context.Context, phone string) (bool, error)
	IsEmailUsedInOidcLoginIdentity(ctx context.Context, email string) (bool, error)

	// Create ---

	StoreUserInTempCache(ctx context.Context, tUser TempPasswordUser) error
	StoreForgetPasswordDataInTempCache(ctx context.Context, forgetPassData ForgetPasswordTmpDataStore) error
	CreatePasswordUser(ctx context.Context, userArgs CreatePasswordUserArgs) (user database_queries.User, err error)
	CreateNewSessionAndAttachUserToInstallation(ctx context.Context, loginIdentityId, installationId int32, token string, ipAddress netip.Addr, expiresAt time.Time) error
	CreateInstallation(ctx context.Context, data CreateInstallationData, installationToken string) error

	LoginOrCreateUserWithOidc(ctx context.Context, data LoginOrCreateUserWithOidcData, tokenGenerator func(userId int32) (string, time.Time, error)) (database_queries.User, error)

	// Update ---

	UpdateusernameForUser(ctx context.Context, userId int32, newUsername string) error

	UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error
	ExpTokenAndUnlinkFromInstallation(ctx context.Context, installationId, tokenId int) error
	ExpAllTokensAndUnlinkThemFromInstallation(ctx context.Context, userId int) error

	ChangePasswordLoginIdentityForUser(ctx context.Context, userId int32, HashedPass, PassSalt string) error

	// Delete ---
	DeleteUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) error
	DeleteForgetPasswordDataFromTempCache(ctx context.Context, dataId uuid.UUID) error
}

type dataSourceImpl struct {
	db    *database.Service
	redis *redis.Client
}

func NewDataSource(db *database.Service, redis *redis.Client) DataSource {
	return &dataSourceImpl{db: db, redis: redis}
}

func (ds dataSourceImpl) GetUserById(ctx context.Context, id int32) (database_queries.UsersGetUserByIdRow, error) {
	dbUser, err := ds.db.Queries.UsersGetUserById(ctx, id)
	if err != nil {
		if dbutils.IsErrPgxNoRows(err) {
			return dbUser, apperr.ErrNoResult
		}
		return dbUser, err
	}
	return dbUser, nil
}

func (ds dataSourceImpl) GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (database_queries.UsersGetUserAndSessionDataBySessionTokenRow, error) {
	userWithSessionData, err := ds.db.Queries.UsersGetUserAndSessionDataBySessionToken(ctx, sessionToken)
	if err != nil {
		if dbutils.IsErrPgxNoRows(err) {
			return userWithSessionData, apperr.ErrNoResult
		}
		return userWithSessionData, err
	}

	return userWithSessionData, err
}

func genTempUserId(id uuid.UUID) string {
	return fmt.Sprint("user:tmp:", id.String())
}

func (ds dataSourceImpl) StoreUserInTempCache(ctx context.Context, tUser TempPasswordUser) error {
	key := genTempUserId(tUser.Id)

	pip := ds.redis.TxPipeline()
	pip.Del(ctx, key)
	pip.HSet(ctx, key, tUser.ToMap())
	pip.Expire(ctx, key, expirationForTempUser)
	resultArray, err := pip.Exec(ctx)

	if err != nil {
		return err
	}

	for _, cmdResult := range resultArray {
		if cmdResult.Err() != nil {
			return err
		}
	}

	return nil
}

func (ds dataSourceImpl) GetUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) (*TempPasswordUser, error) {
	key := genTempUserId(tempUserId)

	result, err := ds.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, apperr.ErrNoResult
	}

	tUser := new(TempPasswordUser).FromMap(result)

	return tUser, err
}

func (ds dataSourceImpl) DeleteUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) error {
	return ds.redis.Del(ctx, genTempUserId(tempUserId)).Err()
}

func (ds dataSourceImpl) CreatePasswordUser(ctx context.Context, userArgs CreatePasswordUserArgs) (user database_queries.User, err error) {
	userRow, err := ds.db.Queries.LoginIdentityCreateNewUserAndPasswordLoginIdentity(
		ctx,
		database_queries.LoginIdentityCreateNewUserAndPasswordLoginIdentityParams{

			UserFirstName:    userArgs.Fname,
			UserUsername:     userArgs.Username,
			UserLastName:     dbutils.ToPgTypeText(userArgs.Lname),
			UserProfileImage: dbutils.ToPgTypeText(userArgs.ProfileImagePath),
			UserRoleID:       dbutils.ToPgTypeInt4(userArgs.RoleID),

			IdentityType:       userArgs.LoginIdentityType.String(),
			PasswordEmail:      dbutils.ToPgTypeText(userArgs.Email),
			PasswordPhone:      dbutils.ToPgTypeText(userArgs.Phone),
			PasswordHashedPass: userArgs.HashedPass,
			PasswordPassSalt:   userArgs.PassSalt,
			PasswordVerifiedAt: pgtype.Timestamptz{Time: userArgs.VerifiedAt, Valid: !userArgs.VerifiedAt.IsZero()},
		},
	)

	user = database_queries.User{
		ID:           userRow.ID,
		Username:     userRow.Username,
		ProfileImage: userRow.ProfileImage,
		FirstName:    userRow.FirstName,
		MiddleName:   userRow.MiddleName,
		LastName:     userRow.LastName,
		CreatedAt:    userRow.CreatedAt,
		UpdatedAt:    userRow.UpdatedAt,
		BlockedAt:    userRow.BlockedAt,
		DeletedAt:    userRow.DeletedAt,
		RoleID:       userRow.RoleID,
	}

	return user, err
}

func (ds dataSourceImpl) UpdateusernameForUser(ctx context.Context, id int32, newUsername string) error {
	return ds.db.Queries.UsersUpdateUsernameForUser(ctx, database_queries.UsersUpdateUsernameForUserParams{ID: id, Username: newUsername})
}

func (ds dataSourceImpl) GetPasswordLoginIdentityWithUser(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database_queries.LoginIdentityGetPasswordLoginIdentityWithUserRow, error) {
	userWithLoginIdentity, err := ds.db.Queries.LoginIdentityGetPasswordLoginIdentityWithUser(
		ctx,
		database_queries.LoginIdentityGetPasswordLoginIdentityWithUserParams{
			IdentityType:  loginIdentityType.String(),
			IdentityValue: identityValue,
		},
	)

	if dbutils.IsErrPgxNoRows(err) {
		err = apperr.ErrNoResult
	}

	return userWithLoginIdentity, err
}

func (ds dataSourceImpl) GetPasswordLoginIdentity(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database_queries.LoginIdentityGetPasswordLoginIdentityRow, error) {
	loginIdentity, err := ds.db.Queries.LoginIdentityGetPasswordLoginIdentity(
		ctx,
		database_queries.LoginIdentityGetPasswordLoginIdentityParams{
			IdentityType:  loginIdentityType.String(),
			IdentityValue: identityValue,
		},
	)

	if dbutils.IsErrPgxNoRows(err) {
		err = apperr.ErrNoResult
	}

	return loginIdentity, err
}

func (ds dataSourceImpl) CreateNewSessionAndAttachUserToInstallation(
	ctx context.Context,
	loginIdentityId,
	installationId int32,
	token string,
	ipAddress netip.Addr,
	expiresAt time.Time,
) (err error) {
	return ds.usingTransaction(
		ctx,
		func(queries *database_queries.Queries) error {
			return ds.createNewSessionAndAttachUserToInstallation(
				ctx,
				loginIdentityId,
				installationId,
				token,
				ipAddress,
				expiresAt,
				queries,
			)
		},
	)
}

func (ds dataSourceImpl) createNewSessionAndAttachUserToInstallation(
	ctx context.Context,
	loginIdentityId,
	installationId int32,
	token string,
	ipAddress netip.Addr,
	expiresAt time.Time,
	queries *database_queries.Queries,
) (err error) {
	sessionId, err := queries.SessionCreateNewSession(
		ctx,
		database_queries.SessionCreateNewSessionParams{
			Token:            token,
			OriginatedFrom:   loginIdentityId,
			UsedInstallation: installationId,
			ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
			IpAddress:        ipAddress,
		},
	)
	if err != nil {
		return err
	}

	affectedRows, err := queries.InstallationAttachSessionToInstallationById(
		ctx,
		database_queries.InstallationAttachSessionToInstallationByIdParams{
			ID:       installationId,
			AttachTo: pgtype.Int4{Int32: sessionId, Valid: true},
		},
	)
	if err != nil {
		return err
	}

	if affectedRows == 0 {
		err = apperr.ErrInstallationTokenInUse
		return err
	}

	err = queries.LoginIdentityUpdateLastUsedAtToNow(ctx, loginIdentityId)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Int32("login_identity_id", loginIdentityId).Msg("can not update the last used at for login identitiy")
		return err
	}

	return nil
}

func (ds dataSourceImpl) GetInstallationUsingTokenAndWhereAttachTo(ctx context.Context, installationToken string, attachedToSessionId int32) (database_queries.Installation, error) {
	installation, err := ds.db.Queries.InstallationGetInstallationUsingTokenAndWhereAttachTo(
		ctx,
		database_queries.InstallationGetInstallationUsingTokenAndWhereAttachToParams{
			InstallationToken: installationToken,
			AttachTo:          pgtype.Int4{Int32: attachedToSessionId, Valid: true},
		},
	)

	if dbutils.IsErrPgxNoRows(err) {
		err = apperr.ErrNoResult
	}

	return installation, err
}

func (ds dataSourceImpl) GetInstallationUsingToken(ctx context.Context, installationToken string) (database_queries.Installation, error) {
	installation, err := ds.db.Queries.InstallationGetInstallationUsingToken(
		ctx,
		installationToken,
	)

	if dbutils.IsErrPgxNoRows(err) {
		err = apperr.ErrNoResult
	}

	return installation, err
}

func (ds dataSourceImpl) IsEmailUsedInPasswordLoginIdentity(ctx context.Context, email string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginIdentityIsEmailUsed(ctx, dbutils.ToPgTypeText(email))
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) IsPhoneUsedInPasswordLoginIdentity(ctx context.Context, phone string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginIdentityIsPhoneUsed(ctx, dbutils.ToPgTypeText(phone))
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) IsEmailUsedInOidcLoginIdentity(ctx context.Context, email string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginIdentityIsOidcEmailUsed(ctx, dbutils.ToPgTypeText(email))
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) ChangePasswordLoginIdentityForUser(ctx context.Context, userId int32, HashedPass, PassSalt string) error {
	return ds.db.Queries.LoginIdentityChangePasswordLoginIdentityByUserId(
		ctx, database_queries.LoginIdentityChangePasswordLoginIdentityByUserIdParams{
			UserID:     userId,
			HashedPass: HashedPass,
			PassSalt:   PassSalt,
		},
	)
}

func (ds dataSourceImpl) GetAllPasswordLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database_queries.LoginIdentityGetAllPasswordLoginIdentitiesByUserIdRow, error) {
	result, err := ds.db.Queries.LoginIdentityGetAllPasswordLoginIdentitiesByUserId(ctx, userId)

	if dbutils.IsErrPgxNoRows(err) {
		err = apperr.ErrNoResult
	}

	return result, err
}

func (ds dataSourceImpl) GetAllLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database_queries.LoginIdentityGetAllByUserIdRow, error) {
	result, err := ds.db.Queries.LoginIdentityGetAllByUserId(ctx, userId)

	if dbutils.IsErrPgxNoRows(err) {
		err = apperr.ErrNoResult
	}

	return result, err
}

func (ds dataSourceImpl) CreateInstallation(ctx context.Context, data CreateInstallationData, installationToken string) error {
	return ds.db.Queries.InstallationCreateNewInstallation(
		ctx,
		database_queries.InstallationCreateNewInstallationParams{
			InstallationToken:       installationToken,
			NotificationToken:       dbutils.ToPgTypeText(data.NotificationToken),
			AppVersion:              data.AppVersion,
			Locale:                  data.Locale,
			DeviceOsVersion:         dbutils.ToPgTypeText(data.DeviceOSVersion),
			DeviceOs:                data.DeviceOS.String(),
			ClientType:              data.ClientType.String(),
			DeviceManufacturer:      dbutils.ToPgTypeText(data.DeviceManufacturer),
			TimezoneOffsetInMinutes: int32(data.TimezoneOffsetInMinutes),
		},
	)
}

func (ds dataSourceImpl) UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error {
	return ds.db.Queries.InstallationUpdateInstallation(
		ctx,
		database_queries.InstallationUpdateInstallationParams{
			InstallationToken:       installationToken,
			NotificationToken:       dbutils.ToPgTypeText(data.NotificationToken),
			Locale:                  data.Locale,
			TimezoneOffsetInMinutes: int32(data.TimezoneOffsetInMinutes),
			AppVersion:              data.AppVersion,
		},
	)
}

func genTempForgetPasswordTmpDataStorId(id uuid.UUID) string {
	return fmt.Sprint("user:forget:password:", id.String())
}

func (ds dataSourceImpl) StoreForgetPasswordDataInTempCache(ctx context.Context, forgetPassData ForgetPasswordTmpDataStore) error {
	key := genTempForgetPasswordTmpDataStorId(forgetPassData.Id)

	pip := ds.redis.TxPipeline()
	pip.Del(ctx, key)
	pip.HSet(ctx, key, forgetPassData.ToMap())
	pip.Expire(ctx, key, expirationForForgetPasswordTempData)
	resultArray, err := pip.Exec(ctx)

	if err != nil {
		return err
	}

	for _, cmdResult := range resultArray {
		if cmdResult.Err() != nil {
			return err
		}
	}

	return nil
}

func (ds dataSourceImpl) GetForgetPasswordDataFromTempCache(ctx context.Context, dataId uuid.UUID) (*ForgetPasswordTmpDataStore, error) {
	key := genTempForgetPasswordTmpDataStorId(dataId)

	result, err := ds.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, apperr.ErrNoResult
	}

	data := new(ForgetPasswordTmpDataStore).FromMap(result)
	return data, err
}

func (ds dataSourceImpl) DeleteForgetPasswordDataFromTempCache(ctx context.Context, dataId uuid.UUID) error {
	return ds.redis.Del(ctx, genTempForgetPasswordTmpDataStorId(dataId)).Err()
}

func (ds dataSourceImpl) ExpTokenAndUnlinkFromInstallation(ctx context.Context, installationId, tokenId int) (err error) {
	return ds.usingTransaction(
		ctx,
		func(queries *database_queries.Queries) error {
			err = queries.SessionSoftDeleteSession(ctx, int32(tokenId))
			if err != nil {
				return err
			}

			err = queries.InstallationDetachSessionFromInstallationById(
				ctx,
				database_queries.InstallationDetachSessionFromInstallationByIdParams{
					ID:           int32(installationId),
					LastAttachTo: pgtype.Int4{Int32: int32(tokenId), Valid: true},
				},
			)
			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (ds dataSourceImpl) ExpAllTokensAndUnlinkThemFromInstallation(ctx context.Context, userId int) error {
	return ds.usingTransaction(
		ctx,
		func(queries *database_queries.Queries) error {
			err := queries.InstallationDetachSessionFromInstallationByUserId(ctx, int32(userId))
			if err != nil {
				return err
			}

			err = queries.SessionSoftDeleteAllActiveSessionsForUser(ctx, int32(userId))
			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (ds dataSourceImpl) LoginOrCreateUserWithOidc(
	ctx context.Context,
	oidcParamData LoginOrCreateUserWithOidcData,
	tokenGenerator func(userId int32) (string, time.Time, error),
) (database_queries.User, error) {

	var user database_queries.User

	fn := func(queries *database_queries.Queries) error {
		var loginIdentityId int32 = -1
		var userId int32 = -1

		oidcUser, err := queries.LoginIdentityGetOIDCDataBySub(
			ctx,
			database_queries.LoginIdentityGetOIDCDataBySubParams{
				OidcSub:          oidcParamData.OidcSub,
				OidcProviderName: oidcParamData.oauthProvider.String(),
			},
		)
		if err != nil {
			if dbutils.IsErrPgxNoRows(err) {
				loginIdentityId, user, err = oidcCreateAccountAndLogin(ctx, queries, oidcParamData)
				if err != nil {
					return err
				}
				userId = user.ID
			} else {
				return err
			}
		} else {
			loginIdentityId = oidcUser.LoginIdentityID
			userId = oidcUser.UserID

			user, err = oidcLoginOnly(ctx, queries, oidcParamData, oidcUser)
			if err != nil {
				return err
			}
		}

		{
			token, expiresAt, err := tokenGenerator(userId)
			if err != nil {
				return err
			}
			err = ds.createNewSessionAndAttachUserToInstallation(
				ctx,
				loginIdentityId,
				oidcParamData.InstallationId,
				token,
				oidcParamData.IpAddress,
				expiresAt,
				queries,
			)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := ds.usingTransaction(ctx, fn)
	if err != nil {
		return database_queries.User{}, err
	}

	return user, nil
}

func oidcCreateAccountAndLogin(ctx context.Context, queries *database_queries.Queries, oidcParamData LoginOrCreateUserWithOidcData) (loginIdentityId int32, user database_queries.User, err error) {
	result, err := queries.LoginIdentityCreateNewUserAndOIDCLoginIdentity(
		ctx,
		database_queries.LoginIdentityCreateNewUserAndOIDCLoginIdentityParams{
			
			UserUsername:               oidcParamData.UserUsername,
			UserProfileImage:           oidcParamData.UserProfileImage,
			UserFirstName:              oidcParamData.UserFirstName,
			UserLastName:               oidcParamData.UserLastName,
			UserRoleID:                 oidcParamData.UserRoleID,
			OauthProviderName:          oidcParamData.oauthProvider.String(),
			OauthProviderIsOidcCapable: true,
			OauthScopes:                oidcParamData.OauthScopes.Array(),
			OauthAccessToken:           oidcParamData.OauthAccessToken,
			OauthRefreshToken:          oidcParamData.OauthRefreshToken,
			OauthTokenType:             oidcParamData.OauthTokenType,
			OauthTokenExpiresAt:        oidcParamData.OauthTokenExpiresAt,
			OauthTokenIssuedAt:         oidcParamData.OauthTokenIssuedAt,
			OidcSub:                    oidcParamData.OidcSub,
			OidcEmail:                  oidcParamData.OidcEmail,
			OidcIss:                    oidcParamData.OidcIss,
			OidcAud:                    oidcParamData.OidcAud,
			OidcGivenName:              oidcParamData.OidcGivenName,
			OidcFamilyName:             oidcParamData.OidcFamilyName,
			OidcName:                   oidcParamData.OidcName,
			OidcPicture:                oidcParamData.OidcPicture,
		},
	)
	if err != nil {
		return -1, database_queries.User{}, err
	}

	user = database_queries.User{
		ID:           result.UserID,
		Username:     result.Username,
		FirstName:    result.FirstName,
		ProfileImage: result.ProfileImage,
		MiddleName:   result.MiddleName,
		LastName:     result.LastName,
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
		BlockedAt:    result.BlockedAt,
		BlockedUntil: result.BlockedUntil,
		RoleID:       result.RoleID,
	}

	return result.NewLoginIdentityID, user, nil
}

func oidcLoginOnly(
	ctx context.Context,
	queries *database_queries.Queries,
	oidcParamData LoginOrCreateUserWithOidcData,
	oidcUser database_queries.LoginIdentityGetOIDCDataBySubRow,
) (database_queries.User, error) {
	// 1. Check if the user already has an OAuth integration associated with the
	//    provided scopes (`data.OauthScopes`).
	//
	//    - If yes:
	//      a. If an access/refresh token is present in `oidcParamData`:
	//         • Update the existing OAuth token in the database.
	//         • Or create a new token record if none exists for this integration.
	//      b. If no access/refresh token is present in `oidcParamData`:
	//         • Do nothing and keep the current state.
	//         • If an old token exists for this scope, keep it until it either:
	//           - Expires, or
	//           - Is explicitly rejected by the OAuth provider.
	//           Once that happens, the token should be deleted.
	//
	//    - If no existing integration is found:
	//      a. Create a new connection if one is not found with a new `oauth_integration`
	//         and `user_integration` record.
	//      b. If an access/refresh token is present in `oidcParamData`:
	//         • Create a new `oauth_token` record and link it to the
	//           `oauth_integration`.
	//      c. If no access/refresh token is provided:
	//         • No further action is required.
	//
	// 2. Update the user's `oidc_data` with the values from the `oidcParamData` param.

	integration, err := queries.OauthIntegrationGetByUserAndScopes(
		ctx,
		database_queries.OauthIntegrationGetByUserAndScopesParams{
			UserID:       oidcUser.UserID,
			OauthScopes:  oidcParamData.OauthScopes.Array(),
			ProviderName: oidcParamData.oauthProvider.String(),
		},
	)
	if err != nil {
		if !dbutils.IsErrPgxNoRows(err) {
			return database_queries.User{}, err
		}
		// No existing connection found for this user with the given scopes and provider.
		// Create a new connection.
		err = queries.OauthCreateConnectionWithIntegrationDataAndTokens(
			ctx,
			database_queries.OauthCreateConnectionWithIntegrationDataAndTokensParams{
				UserID:       oidcUser.UserID,
				ProviderName: oidcUser.OauthProviderName.String,
				Scopes:       oidcParamData.OauthScopes.Array(),
				AccessToken:  oidcParamData.OauthAccessToken,
				RefreshToken: oidcParamData.OauthRefreshToken,
				TokenType:    oidcParamData.OauthTokenType,
				ExpiresAt:    oidcParamData.OauthTokenExpiresAt,
				IssuedAt:     oidcParamData.OauthTokenIssuedAt,
			},
		)
		if err != nil {
			return database_queries.User{}, err
		}

	} else {
		// A connection exists for the user with the same scopes and provider.
		// Update the access tokens if present.
		if oidcParamData.OauthAccessToken.Valid || oidcParamData.OauthRefreshToken.Valid {
			if integration.OauthTokenID.Valid { // there is a token data record just update its data
				err := queries.OauthTokenUpdate(
					ctx,
					database_queries.OauthTokenUpdateParams{
						ID:           integration.OauthTokenID.Int32,
						AccessToken:  oidcParamData.OauthAccessToken,
						RefreshToken: oidcParamData.OauthRefreshToken,
						TokenType:    oidcParamData.OauthTokenType,
						ExpiresAt:    oidcParamData.OauthTokenExpiresAt,
						IssuedAt:     oidcParamData.OauthTokenIssuedAt,
					},
				)
				if err != nil {
					return database_queries.User{}, err
				}
			} else { // there is no record for oauth token create one for this integration
				err := queries.OauthTokenCreate(
					ctx,
					database_queries.OauthTokenCreateParams{
						OauthIntegrationID: integration.OauthIntegrationID,
						AccessToken:        oidcParamData.OauthAccessToken,
						RefreshToken:       oidcParamData.OauthRefreshToken,
						TokenType:          oidcParamData.OauthTokenType,
						ExpiresAt:          oidcParamData.OauthTokenExpiresAt,
						IssuedAt:           oidcParamData.OauthTokenIssuedAt,
					},
				)
				if err != nil {
					return database_queries.User{}, err
				}
			}
		}
	}

	err = queries.OidcDataUpdateRecored(
		ctx,
		database_queries.OidcDataUpdateRecoredParams{
			ID:         oidcUser.OidcDataID,
			Email:      oidcParamData.OidcEmail,
			GivenName:  oidcParamData.OidcGivenName,
			FamilyName: oidcParamData.OidcFamilyName,
			Name:       oidcParamData.OidcName,
			Picture:    oidcParamData.OidcPicture,
		},
	)
	if err != nil {
		return database_queries.User{}, err
	}

	err = queries.LoginIdentityUpdateLastUsedAtToNow(ctx, oidcUser.LoginIdentityID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Int32("login_identity_id", oidcUser.LoginIdentityID).Msg("can not update the last used at for login identitiy")
		return database_queries.User{}, err
	}

	user := database_queries.User{
		ID:           oidcUser.UserID,
		Username:     oidcUser.UserUsername,
		FirstName:    oidcUser.UserFirstName,
		ProfileImage: oidcUser.UserProfileImage,
		MiddleName:   oidcUser.UserMiddleName,
		LastName:     oidcUser.UserLastName,
		CreatedAt:    oidcUser.UserCreatedAt,
		UpdatedAt:    oidcUser.UserUpdatedAt,
		BlockedAt:    oidcUser.UserBlockedAt,
		BlockedUntil: oidcUser.UserBlockedUntil,
		RoleID:       oidcUser.UserRoleID,
	}
	return user, nil
}

func (ds dataSourceImpl) usingTransaction(ctx context.Context, fn func(queries *database_queries.Queries) error) error {
	tx, err := ds.db.ConnPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		rolbackFn := func() {
			rollBackErr := tx.Rollback(ctx)
			err = errors.Join(rollBackErr, ctx.Err(), err)
		}
		commitFn := func() {
			commitErr := tx.Commit(ctx)
			err = errors.Join(commitErr, err)
		}

		select {
		case <-ctx.Done():
			rolbackFn()
		default:
			if err != nil {
				rolbackFn()
			} else {
				commitFn()
			}
		}
	}()

	queries := ds.db.Queries.WithTx(tx)
	err = fn(queries)
	return err
}
