package database

import (
	"cmp"
	"context"
	"slices"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database/database_queries"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/perm/baseperm"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/settings/labels"
	dbutils "github.com/Nidal-Bakir/go-todo-backend/internal/utils/db_utils"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type seeder struct {
	version  int32
	seederFn func(ctx context.Context, dbTx database_queries.DBTX, queries *database_queries.Queries) error
}

// add new seeders here
var seeders = []seeder{
	v1_baseRollsAndPermission,
	v2_settingsClientApiToken,
}

func seed(ctx context.Context, db *Service) (err error) {
	zlog := zerolog.Ctx(ctx)

	tx, err := db.ConnPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	queries := db.Queries.WithTx(tx)

	slices.SortFunc(seeders, func(a, b seeder) int {
		return cmp.Compare(a.version, b.version)
	})

	memoryVersion := seeders[len(seeders)-1].version
	dbVersion, err := db.Queries.SeederVersionReadLatestAppliedVersion(ctx)
	if err != nil {
		if dbutils.IsErrPgxNoRows(err) {
			dbVersion = -1
		} else {
			return err
		}
	}
	if memoryVersion == dbVersion {
		zlog.Info().
			Int32("current_version", memoryVersion).
			Msg("No seeding required.")
		return nil
	}
	if memoryVersion < dbVersion {
		zlog.Fatal().
			Int32("db_version", dbVersion).
			Int32("memory_version", memoryVersion).
			Msg("Database version is ahead of the in-memory version. Downward migrations or unseeding are not supported.")
		return nil
	}

	for _, s := range seeders {
		if s.version > dbVersion {
			if err := s.seederFn(ctx, tx, queries); err != nil {
				tx.Rollback(ctx)
				return err
			}
		}
	}

	err = db.Queries.SeederVersionAddValue(ctx, int32(memoryVersion))
	if err != nil {
		return err
	}

	zlog.Info().
		Int32("current_version", memoryVersion).
		Msg("Seeding completed successfully.")

	return nil
}

var v1_baseRollsAndPermission = seeder{
	version: 1,
	seederFn: func(ctx context.Context, dbTx database_queries.DBTX, queries *database_queries.Queries) error {
		adminRole, err := queries.PermCreateNewRole(ctx, baseperm.BaseRollAdmin)
		if err != nil {
			return err
		}

		readAppSettingsPer, err := queries.PermCreateNewPermission(ctx, baseperm.BasePermReadAppSettings)
		if err != nil {
			return err
		}
		writeAppSettingsPer, err := queries.PermCreateNewPermission(ctx, baseperm.BasePermWriteAppSettings)
		if err != nil {
			return err
		}
		deleteAppSettingsPer, err := queries.PermCreateNewPermission(ctx, baseperm.BasePermDeleteAppSettings)
		if err != nil {
			return err
		}

		err = queries.PermAddPermissionToRole(
			ctx,
			database_queries.PermAddPermissionToRoleParams{
				RoleName:       adminRole.Name,
				PermissionName: readAppSettingsPer.Name,
			},
		)
		if err != nil {
			return err
		}

		err = queries.PermAddPermissionToRole(
			ctx,
			database_queries.PermAddPermissionToRoleParams{
				RoleName:       adminRole.Name,
				PermissionName: writeAppSettingsPer.Name,
			},
		)
		if err != nil {
			return err
		}

		err = queries.PermAddPermissionToRole(
			ctx,
			database_queries.PermAddPermissionToRoleParams{
				RoleName:       adminRole.Name,
				PermissionName: deleteAppSettingsPer.Name,
			},
		)
		if err != nil {
			return err
		}

		return nil
	},
}

var v2_settingsClientApiToken = seeder{
	version: 2,
	seederFn: func(ctx context.Context, dbTx database_queries.DBTX, queries *database_queries.Queries) error {
		err := queries.SettingsCreateLabel(
			ctx,
			labels.ClientApiTokenWeb,
		)
		if err != nil {
			return err
		}
		err = queries.SettingsCreateLabel(
			ctx,
			labels.ClientApiTokenMobile,
		)
		if err != nil {
			return err
		}
		return nil
	},
}
