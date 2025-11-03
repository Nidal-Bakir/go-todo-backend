package database

import (
	"cmp"
	"context"
	"slices"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database/database_queries"
	dbutils "github.com/Nidal-Bakir/go-todo-backend/internal/utils/db_utils"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

const createSeederTableIfNotExistsSqlCMD = `
  CREATE TABLE IF NOT EXISTS seeder_version (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    version INTEGER UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
`

const insertSeederVersion = `
  INSERT INTO seeder_version (
    version
  )
  VALUES (
    $1
  );
`
const readLatestAppliedVersion = `
  SELECT
    version
  FROM
    seeder_version
  ORDER BY
    version DESC
  LIMIT 1;
`

type seeder struct {
	version  int
	seederFn func(ctx context.Context, dbTx database_queries.DBTX, queries *database_queries.Queries) error
}

var seeders = []seeder{
	v1_baseRollsAndPermission,
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

	_, err = tx.Exec(ctx, createSeederTableIfNotExistsSqlCMD)
	if err != nil {
		return err
	}

	slices.SortFunc(seeders, func(a, b seeder) int {
		return cmp.Compare(a.version, b.version)
	})

	memoryVersion := seeders[len(seeders)-1].version
	dbVersion, err := readLatestAppliedSeederVersion(ctx, tx)
	if err != nil {
		return err
	}
	if memoryVersion == dbVersion {
		zlog.Info().
			Int("current_version", memoryVersion).
			Msg("No seeding required.")
		return nil
	}
	if memoryVersion < dbVersion {
		zlog.Fatal().
			Int("db_version", dbVersion).
			Int("memory_version", memoryVersion).
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

	_, err = tx.Exec(ctx, insertSeederVersion, memoryVersion)
	if err != nil {
		return err
	}

	zlog.Info().
		Int("current_version", memoryVersion).
		Msg("Seeding completed successfully.")

	return nil
}

func readLatestAppliedSeederVersion(ctx context.Context, tx pgx.Tx) (int, error) {
	ver := -1
	row := tx.QueryRow(ctx, readLatestAppliedVersion)
	err := row.Scan(&ver)
	if dbutils.IsErrPgxNoRows(err) {
		err = nil
	}
	return ver, err
}

var v1_baseRollsAndPermission = seeder{
	version: 1,
	seederFn: func(ctx context.Context, dbTx database_queries.DBTX, queries *database_queries.Queries) error {
		adminRole, err := queries.PerRollCreateNewRole(ctx, "admin")
		if err != nil {
			return err
		}

		readAppSettingsPer, err := queries.PerRollCreateNewPermission(ctx, "read_app_settings")
		if err != nil {
			return err
		}
		writeAppSettingsPer, err := queries.PerRollCreateNewPermission(ctx, "write_app_settings")
		if err != nil {
			return err
		}
		deleteAppSettingsPer, err := queries.PerRollCreateNewPermission(ctx, "delete_app_settings")
		if err != nil {
			return err
		}

		err = queries.PerRollAddPermissionToRole(
			ctx,
			database_queries.PerRollAddPermissionToRoleParams{
				RoleID:       adminRole.ID,
				PermissionID: readAppSettingsPer.ID,
			},
		)
		if err != nil {
			return err
		}

		err = queries.PerRollAddPermissionToRole(
			ctx,
			database_queries.PerRollAddPermissionToRoleParams{
				RoleID:       adminRole.ID,
				PermissionID: writeAppSettingsPer.ID,
			},
		)
		if err != nil {
			return err
		}

		err = queries.PerRollAddPermissionToRole(
			ctx,
			database_queries.PerRollAddPermissionToRoleParams{
				RoleID:       adminRole.ID,
				PermissionID: deleteAppSettingsPer.ID,
			},
		)
		if err != nil {
			return err
		}

		return nil
	},
}
