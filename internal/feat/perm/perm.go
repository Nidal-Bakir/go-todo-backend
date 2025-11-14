package perm

import (
	"context"
	"errors"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Repository interface {
	HasPermission(ctx context.Context, role string, requestedPermissions ...string) (bool, error)
	HasPermissionErr(ctx context.Context, role string, requestedPermissions ...string) error
}

func NewRepository(db *database.Service, redis *redis.Client) Repository {
	return &repositoryImpl{db: db, redis: redis}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	db    *database.Service
	redis *redis.Client
}

func (r *repositoryImpl) HasPermission(ctx context.Context, role string, requestedPermissions ...string) (bool, error) {
	err := r.HasPermissionErr(ctx, role, requestedPermissions...)
	if err != nil {
		if errors.Is(err, apperr.ErrPermissionDenied) {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func (r *repositoryImpl) HasPermissionErr(ctx context.Context, role string, requestedPermissions ...string) error {
	if role == "" {
		return apperr.ErrPermissionDenied
	}
	zlog := zerolog.Ctx(ctx)

	dbResult, err := r.db.Queries.PermGetRoleWithItsPermissions(ctx, role)
	if err != nil {
		zlog.Err(err).
			Str("role", role).
			Msg("failed to load permissions for role")
		return err
	}

	rolePerms := make(map[string]struct{}, len(dbResult))
	for _, res := range dbResult {
		rolePerms[res.PermissionName] = struct{}{}
	}

	for _, requested := range requestedPermissions {
		if _, ok := rolePerms[requested]; !ok {
			return apperr.ErrPermissionDenied
		}
	}

	return nil
}
