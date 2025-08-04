package auth

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

const (
	expirationForTempUser               = time.Minute * 30
	expirationForForgetPasswordTempData = time.Minute * 15
)

type DataSource interface {
	// Query ---

	GetUserById(ctx context.Context, id int32) (database.UsersGetUserByIdRow, error)

	GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (database.UsersGetUserAndSessionDataBySessionTokenRow, error)

	GetUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) (*TempPasswordUser, error)
	GetForgetPasswordDataFromTempCache(ctx context.Context, dataId uuid.UUID) (*ForgetPasswordTmpDataStore, error)

	GetInstallationUsingTokenAndWhereAttachTo(ctx context.Context, installationToken string, attachedToSession int32) (database.Installation, error)
	GetInstallationUsingToken(ctx context.Context, installationToken string) (database.Installation, error)

	GetPasswordLoginIdentityWithUser(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database.LoginIdentityGetPasswordLoginIdentityWithUserRow, error)
	GetPasswordLoginIdentity(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database.LoginIdentityGetPasswordLoginIdentityRow, error)
	GetAllPasswordLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database.LoginIdentityGetAllPasswordLoginIdentitiesByUserIdRow, error)
	GetAllLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database.LoginIdentityGetAllByUserIdRow, error)

	IsEmailUsedInPasswordLoginIdentity(ctx context.Context, email string) (bool, error)
	IsPhoneUsedInPasswordLoginIdentity(ctx context.Context, phone string) (bool, error)
	IsEmailUsedInOidcLoginIdentity(ctx context.Context, email string) (bool, error)

	// Create ---

	StoreUserInTempCache(ctx context.Context, tUser TempPasswordUser) error
	StoreForgetPasswordDataInTempCache(ctx context.Context, forgetPassData ForgetPasswordTmpDataStore) error
	CreatePasswordUser(ctx context.Context, userArgs CreatePasswordUserArgs) (user database.User, err error)
	CreateNewSessionAndAttachUserToInstallation(ctx context.Context, loginOptionId, installationId int32, token string, ipAddress netip.Addr, expiresAt time.Time) error
	CreateInstallation(ctx context.Context, data CreateInstallationData, installationToken string) error

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

func (ds dataSourceImpl) GetUserById(ctx context.Context, id int32) (database.UsersGetUserByIdRow, error) {
	dbUser, err := ds.db.Queries.UsersGetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbUser, apperr.ErrNoResult
		}
		return dbUser, err
	}
	return dbUser, nil
}

func (ds dataSourceImpl) GetUserAndSessionDataBySessionToken(ctx context.Context, sessionToken string) (database.UsersGetUserAndSessionDataBySessionTokenRow, error) {
	userWithSessionData, err := ds.db.Queries.UsersGetUserAndSessionDataBySessionToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (ds dataSourceImpl) CreatePasswordUser(ctx context.Context, userArgs CreatePasswordUserArgs) (user database.User, err error) {
	userRow, err := ds.db.Queries.LoginIdentityCreateNewUserAndPasswordLoginIdentity(
		ctx,
		database.LoginIdentityCreateNewUserAndPasswordLoginIdentityParams{

			UserFirstName:    userArgs.Fname,
			UserUsername:     userArgs.Username,
			UserLastName:     ds.toPgTypeText(userArgs.Lname),
			UserProfileImage: ds.toPgTypeText(userArgs.ProfileImagePath),
			UserRoleID:       ds.toPgTypeInt4(userArgs.RoleID),

			IdentityType:       userArgs.LoginIdentityType.String(),
			PasswordEmail:      ds.toPgTypeText(userArgs.Email),
			PasswordPhone:      ds.toPgTypeText(userArgs.Phone),
			PasswordHashedPass: userArgs.HashedPass,
			PasswordPassSalt:   userArgs.PassSalt,
			PasswordVerifiedAt: pgtype.Timestamptz{Time: userArgs.VerifiedAt, Valid: !userArgs.VerifiedAt.IsZero()},
		},
	)

	user = database.User{
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
	return ds.db.Queries.UsersUpdateUsernameForUser(ctx, database.UsersUpdateUsernameForUserParams{ID: id, Username: newUsername})
}

func (ds dataSourceImpl) GetPasswordLoginIdentityWithUser(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database.LoginIdentityGetPasswordLoginIdentityWithUserRow, error) {
	userWithLoginIdentity, err := ds.db.Queries.LoginIdentityGetPasswordLoginIdentityWithUser(
		ctx,
		database.LoginIdentityGetPasswordLoginIdentityWithUserParams{
			IdentityType:  loginIdentityType.String(),
			IdentityValue: identityValue,
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return userWithLoginIdentity, err
}

func (ds dataSourceImpl) GetPasswordLoginIdentity(ctx context.Context, identityValue string, loginIdentityType LoginIdentityType) (database.LoginIdentityGetPasswordLoginIdentityRow, error) {
	loginIdentity, err := ds.db.Queries.LoginIdentityGetPasswordLoginIdentity(
		ctx,
		database.LoginIdentityGetPasswordLoginIdentityParams{
			IdentityType:  loginIdentityType.String(),
			IdentityValue: identityValue,
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return loginIdentity, err
}

func (ds dataSourceImpl) CreateNewSessionAndAttachUserToInstallation(
	ctx context.Context,
	loginOptionId,
	installationId int32,
	token string,
	ipAddress netip.Addr,
	expiresAt time.Time,
) (err error) {
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

	sessionId, err := queries.SessionCreateNewSession(
		ctx,
		database.SessionCreateNewSessionParams{
			Token:            token,
			OriginatedFrom:   loginOptionId,
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
		database.InstallationAttachSessionToInstallationByIdParams{
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

	return nil
}

func (ds dataSourceImpl) GetInstallationUsingTokenAndWhereAttachTo(ctx context.Context, installationToken string, attachedToSessionId int32) (database.Installation, error) {
	installation, err := ds.db.Queries.InstallationGetInstallationUsingTokenAndWhereAttachTo(
		ctx,
		database.InstallationGetInstallationUsingTokenAndWhereAttachToParams{
			InstallationToken: installationToken,
			AttachTo:          pgtype.Int4{Int32: attachedToSessionId, Valid: true},
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return installation, err
}

func (ds dataSourceImpl) GetInstallationUsingToken(ctx context.Context, installationToken string) (database.Installation, error) {
	installation, err := ds.db.Queries.InstallationGetInstallationUsingToken(
		ctx,
		installationToken,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return installation, err
}

func (ds dataSourceImpl) IsEmailUsedInPasswordLoginIdentity(ctx context.Context, email string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginIdentityIsEmailUsed(ctx, ds.toPgTypeText(email))
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) IsPhoneUsedInPasswordLoginIdentity(ctx context.Context, phone string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginIdentityIsPhoneUsed(ctx, ds.toPgTypeText(phone))
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) IsEmailUsedInOidcLoginIdentity(ctx context.Context, email string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginIdentityIsOidcEmailUsed(ctx, ds.toPgTypeText(email))
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) ChangePasswordLoginIdentityForUser(ctx context.Context, userId int32, HashedPass, PassSalt string) error {
	return ds.db.Queries.LoginIdentityChangePasswordLoginIdentityByUserId(
		ctx, database.LoginIdentityChangePasswordLoginIdentityByUserIdParams{
			UserID:     userId,
			HashedPass: HashedPass,
			PassSalt:   PassSalt,
		},
	)
}

func (ds dataSourceImpl) GetAllPasswordLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database.LoginIdentityGetAllPasswordLoginIdentitiesByUserIdRow, error) {
	result, err := ds.db.Queries.LoginIdentityGetAllPasswordLoginIdentitiesByUserId(ctx, userId)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return result, err
}

func (ds dataSourceImpl) GetAllLoginIdentitiesForUser(ctx context.Context, userId int32) ([]database.LoginIdentityGetAllByUserIdRow, error) {
	result, err := ds.db.Queries.LoginIdentityGetAllByUserId(ctx, userId)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return result, err
}

func (ds dataSourceImpl) CreateInstallation(ctx context.Context, data CreateInstallationData, installationToken string) error {
	return ds.db.Queries.InstallationCreateNewInstallation(
		ctx,
		database.InstallationCreateNewInstallationParams{
			InstallationToken:       installationToken,
			NotificationToken:       ds.toPgTypeText(data.NotificationToken),
			AppVersion:              data.AppVersion,
			Locale:                  data.Locale,
			DeviceOsVersion:         ds.toPgTypeText(data.DeviceOSVersion),
			DeviceOs:                ds.toPgTypeText(data.DeviceOS),
			DeviceManufacturer:      ds.toPgTypeText(data.DeviceManufacturer),
			TimezoneOffsetInMinutes: int32(data.TimezoneOffsetInMinutes),
		},
	)
}

func (ds dataSourceImpl) UpdateInstallation(ctx context.Context, installationToken string, data UpdateInstallationData) error {
	return ds.db.Queries.InstallationUpdateInstallation(
		ctx, database.InstallationUpdateInstallationParams{
			InstallationToken:       installationToken,
			NotificationToken:       ds.toPgTypeText(data.NotificationToken),
			Locale:                  data.Locale,
			TimezoneOffsetInMinutes: int32(data.TimezoneOffsetInMinutes),
			AppVersion:              data.AppVersion,
		},
	)
}

func (dataSourceImpl) toPgTypeText(str string) pgtype.Text {
	return pgtype.Text{String: str, Valid: len(str) != 0}
}

func (dataSourceImpl) toPgTypeInt4(num *int32) pgtype.Int4 {
	if num == nil {
		return pgtype.Int4{Int32: -1, Valid: false}
	}
	return pgtype.Int4{Int32: int32(*num), Valid: true}
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

	err = queries.SessionSoftDeleteSession(ctx, int32(tokenId))
	if err != nil {
		return err
	}

	err = queries.InstallationDetachSessionFromInstallationById(
		ctx,
		database.InstallationDetachSessionFromInstallationByIdParams{
			ID:           int32(installationId),
			LastAttachTo: pgtype.Int4{Int32: int32(tokenId), Valid: true},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (ds dataSourceImpl) ExpAllTokensAndUnlinkThemFromInstallation(ctx context.Context, userId int) (err error) {
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

	err = queries.InstallationDetachSessionFromInstallationByUserId(ctx, int32(userId))
	if err != nil {
		return err
	}

	err = queries.SessionSoftDeleteAllActiveSessionsForUser(ctx, int32(userId))
	if err != nil {
		return err
	}

	return nil
}
