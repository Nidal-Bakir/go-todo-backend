package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

const (
	expirationForTempUser = time.Minute * 30
)

type DataSource interface {
	// Query ---

	GetUserById(ctx context.Context, id int32) (database.User, error)
	GetUserBySessionToken(ctx context.Context, sessionToken string) (database.User, error)

	GetUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) (*TempUser, error)

	GetInstallationUsingUUIdAndWhereAttachTo(ctx context.Context, InstallationId uuid.UUID, attachedToUser int32) (database.Installation, error)
	GetInstallationUsingUUID(ctx context.Context, InstallationId uuid.UUID) (database.Installation, error)

	GetActiveLoginOptionWithUser(ctx context.Context, accessKey string, loginMethod LoginMethod) (database.LoginOptionGetActiveLoginOptionWithUserRow, error)
	GetAllActiveLoginOptionByUserIdAndSupportPassword(ctx context.Context, userId int32) ([]database.LoginOption, error)
	IsAccessKeyUsedInAnyLoginOption(ctx context.Context, accessKey string) (bool, error)

	// Create ---

	StoreUserInTempCache(ctx context.Context, tUser *TempUser) (*TempUser, error)
	CreateUser(ctx context.Context, userArgs CreateUserArgs) (user database.User, err error)
	CreateNewSession(ctx context.Context, loginOptionId, installationId int32, token string, expiresAt time.Time) error

	// Update ---

	UpdateusernameForUser(ctx context.Context, userId int32, newUsername string) error

	// change the all the passwords of all the login options that support password usage
	ChangeAllPasswordsForLoginOptions(ctx context.Context, userId int32, HashedPass, PassSalt string) error

	// Delete ---
	DeleteUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) error
}

type dataSourceImpl struct {
	db    *database.Service
	redis *redis.Client
}

func NewDataSource(db *database.Service, redis *redis.Client) DataSource {
	return &dataSourceImpl{db: db, redis: redis}
}

func (ds dataSourceImpl) GetUserById(ctx context.Context, id int32) (database.User, error) {
	dbUser, err := ds.db.Queries.UsersGetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbUser, apperr.ErrNoResult
		}
		return dbUser, err
	}
	return dbUser, nil
}

func (ds dataSourceImpl) GetUserBySessionToken(ctx context.Context, sessionToken string) (database.User, error) {
	dbUser, err := ds.db.Queries.UsersGetUserBySessionToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbUser, apperr.ErrNoResult
		}
		return dbUser, err
	}

	return dbUser, err
}

func genTempUserId(id uuid.UUID) string {
	return fmt.Sprint("user:tmp:", id.String())
}

func (ds dataSourceImpl) StoreUserInTempCache(ctx context.Context, tUser *TempUser) (*TempUser, error) {
	key := genTempUserId(tUser.Id)

	pip := ds.redis.TxPipeline()
	pip.Del(ctx, key)
	pip.HSet(ctx, key, tUser.ToMap())
	pip.Expire(ctx, key, expirationForTempUser)
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

func (ds dataSourceImpl) GetUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) (*TempUser, error) {
	key := genTempUserId(tempUserId)

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

func (ds dataSourceImpl) DeleteUserFromTempCache(ctx context.Context, tempUserId uuid.UUID) error {
	return ds.redis.Del(ctx, genTempUserId(tempUserId)).Err()
}

func (ds dataSourceImpl) CreateUser(ctx context.Context, userArgs CreateUserArgs) (user database.User, err error) {
	tx, err := ds.db.ConnPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return database.User{}, err
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

	user, err = queries.UsersCreateNewUser(
		ctx,
		database.UsersCreateNewUserParams{
			FirstName:    userArgs.Fname,
			Username:     userArgs.Username,
			LastName:     pgtype.Text{String: userArgs.Lname, Valid: userArgs.Lname != ""},
			ProfileImage: pgtype.Text{String: userArgs.ProfileImagePath, Valid: userArgs.ProfileImagePath != ""},
			RoleID:       pgtype.Int4{Int32: userArgs.RoleID, Valid: userArgs.RoleID != 0},
		})

	_, err = queries.LoginOptionCreateNewLoginOption(
		ctx,
		database.LoginOptionCreateNewLoginOptionParams{
			UserID:      user.ID,
			LoginMethod: userArgs.LoginMethod.String(),
			AccessKey:   userArgs.AccessKey,
			HashedPass:  pgtype.Text{String: userArgs.HashedPass, Valid: userArgs.HashedPass != ""},
			PassSalt:    pgtype.Text{String: userArgs.PassSalt, Valid: userArgs.PassSalt != ""},
			VerifiedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	)

	return user, err
}

func (ds dataSourceImpl) UpdateusernameForUser(ctx context.Context, id int32, newUsername string) error {
	return ds.db.Queries.UsersUpdateUsernameForUser(ctx, database.UsersUpdateUsernameForUserParams{ID: id, Username: newUsername})
}

func (ds dataSourceImpl) GetActiveLoginOptionWithUser(ctx context.Context, accessKey string, loginMethod LoginMethod) (database.LoginOptionGetActiveLoginOptionWithUserRow, error) {
	userWithLoginOption, err := ds.db.Queries.LoginOptionGetActiveLoginOptionWithUser(
		ctx,
		database.LoginOptionGetActiveLoginOptionWithUserParams{
			LoginMethod: loginMethod.String(),
			AccessKey:   accessKey,
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return userWithLoginOption, err
}

func (ds dataSourceImpl) CreateNewSession(ctx context.Context, loginOptionId, installationId int32, token string, expiresAt time.Time) error {
	return ds.db.Queries.SessionCreateNewSession(
		ctx,
		database.SessionCreateNewSessionParams{
			Token:            token,
			OriginatedFrom:   loginOptionId,
			UsedInstallation: installationId,
			ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
		},
	)
}

func (ds dataSourceImpl) GetInstallationUsingUUIdAndWhereAttachTo(ctx context.Context, InstallationId uuid.UUID, attachedToUser int32) (database.Installation, error) {
	installation, err := ds.db.Queries.InstallationGetInstallationUsingUUIdAndWhereAttachTo(
		ctx,
		database.InstallationGetInstallationUsingUUIdAndWhereAttachToParams{
			InstallationID: InstallationId,
			AttachTo:       pgtype.Int4{Int32: attachedToUser, Valid: true},
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return installation, err
}

func (ds dataSourceImpl) GetInstallationUsingUUID(ctx context.Context, InstallationId uuid.UUID) (database.Installation, error) {
	installation, err := ds.db.Queries.InstallationGetInstallationUsingUUID(
		ctx,
		InstallationId,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return installation, err
}

func (ds dataSourceImpl) IsAccessKeyUsedInAnyLoginOption(ctx context.Context, accessKey string) (isUsed bool, err error) {
	count, err := ds.db.Queries.LoginOptionIsAccessKeyUsed(ctx, accessKey)
	if count > 0 {
		isUsed = true
	}
	return isUsed, err
}

func (ds dataSourceImpl) ChangeAllPasswordsForLoginOptions(ctx context.Context, userId int32, HashedPass, PassSalt string) error {
	return ds.db.Queries.LoginOptionChangeAllPasswordsByUserId(
		ctx, database.LoginOptionChangeAllPasswordsByUserIdParams{
			UserID:     userId,
			HashedPass: pgtype.Text{String: HashedPass, Valid: true},
			PassSalt:   pgtype.Text{String: PassSalt, Valid: true},
		},
	)
}

func (ds dataSourceImpl) GetAllActiveLoginOptionByUserIdAndSupportPassword(ctx context.Context, userId int32) ([]database.LoginOption, error) {
	result, err := ds.db.Queries.LoginOptionGetAllActiveByUserIdAndSupportPassword(ctx, userId)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	}

	return result, err
}
