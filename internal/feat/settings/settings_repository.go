package settings

import (
	"context"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database/database_queries"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/perm"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/perm/baseperm"
	dbutils "github.com/Nidal-Bakir/go-todo-backend/internal/utils/db_utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Repository interface {
	GetSetting(ctx context.Context, role, label string) (string, error)
	SetSetting(ctx context.Context, role, label, value string) error
	DeleteSetting(ctx context.Context, role, label string) error
}

func NewRepository(db *database.Service, redis *redis.Client, permRepo perm.Repository) Repository {
	return &repositoryImpl{db: db, redis: redis, permRepo: permRepo}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	db       *database.Service
	redis    *redis.Client
	permRepo perm.Repository
}

const redisKey = "app:settings"

func (r repositoryImpl) GetSetting(ctx context.Context, role, label string) (string, error) {
	zlog := zerolog.Ctx(ctx).With().Str("label", label).Logger()

	err := r.permRepo.HasPermissionErr(ctx, role, baseperm.BasePermReadAppSettings)
	if err != nil {
		return "", err
	}

	if val := r.readSettingFromCache(ctx, label, zlog); val != nil {
		return *val, nil
	}

	setting, err := r.db.Queries.SettingsGetByLable(ctx, label)
	if err != nil {
		if dbutils.IsErrPgxNoRows(err) {
			err = apperr.ErrNoResult
		} else {
			zlog.Err(err).Msg("can not get setting")
		}
		return "", err
	}

	if setting.Value.Valid {
		r.addSettingToCache(ctx, label, setting.Value.String, zlog)
	}

	return setting.Value.String, nil
}

func (r repositoryImpl) readSettingFromCache(ctx context.Context, label string, zlog zerolog.Logger) *string {
	val, err := r.redis.HGet(ctx, redisKey, label).Result()
	if err != nil {
		if !dbutils.IsErrRedisNilNoRows(err) {
			zlog.Err(err).Msg("could not read the settings value from redis")
		}
		return nil
	}
	return &val
}

func (r repositoryImpl) addSettingToCache(ctx context.Context, label, value string, zlog zerolog.Logger) {
	err := r.redis.HSetEXWithArgs(
		ctx,
		redisKey,
		&redis.HSetEXOptions{
			ExpirationType: redis.HSetEXExpirationEX,
			ExpirationVal:  int64(time.Hour.Seconds()),
		},
		label,
		value,
	).Err()
	if err != nil {
		zlog.Err(err).Msg("can not set the setting value in redis")
	}
}

func (r repositoryImpl) SetSetting(ctx context.Context, role, label, value string) error {
	zlog := zerolog.Ctx(ctx).With().Str("label", label).Str("value", value).Logger()

	err := r.permRepo.HasPermissionErr(ctx, role, baseperm.BasePermWriteAppSettings)
	if err != nil {
		return err
	}

	err = r.db.Queries.SettingsSetSetting(
		ctx,
		database_queries.SettingsSetSettingParams{
			Label: label,
			Value: dbutils.ToPgTypeText(value),
		},
	)
	if err != nil {
		zlog.Err(err).Msg("can not set setting")
		return err
	}

	r.addSettingToCache(ctx, label, value, zlog)

	return nil
}

func (r repositoryImpl) DeleteSetting(ctx context.Context, role, label string) error {
	zlog := zerolog.Ctx(ctx).With().Str("label", label).Logger()

	err := r.permRepo.HasPermissionErr(ctx, role, baseperm.BasePermDeleteAppSettings)
	if err != nil {
		return err
	}

	if err := r.redis.HDel(ctx, redisKey, label).Err(); err != nil {
		zlog.Err(err).Msg("could not delete the settign from cache")
	}

	err = r.db.Queries.SettingsDeleteByLable(ctx, label)
	if err != nil {
		zlog.Err(err).Msg("can not delete setting")
		return err
	}
	return nil
}
