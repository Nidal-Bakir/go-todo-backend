package settings

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database/database_queries"
	dbutils "github.com/Nidal-Bakir/go-todo-backend/internal/utils/db_utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Repository interface {
	GetSetting(ctx context.Context, label string) (string, error)
	setSetting(ctx context.Context, label, value string) error
}

func NewRepository(db *database.Service, redis *redis.Client) Repository {
	return &repositoryImpl{db: db, redis: redis}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	db    *database.Service
	redis *redis.Client
}

func (r repositoryImpl) GetSetting(ctx context.Context, label string) (string, error) {
	zlog := zerolog.Ctx(ctx).With().Str("label", label).Logger()

	setting, err := r.db.Queries.SettingsGetByLable(ctx, label)
	if err != nil {
		if dbutils.IsErrPgxNoRows(err) {
			err = apperr.ErrNoResult
		} else {
			zlog.Err(err).Msg("can not get setting")
		}
		return "", err
	}

	return setting.Value.String, nil
}

func (r repositoryImpl) setSetting(ctx context.Context, label, value string) error {
	zlog := zerolog.Ctx(ctx).With().Str("label", label).Str("value", value).Logger()

	err := r.db.Queries.SettingsSetSetting(
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
	return nil
}
