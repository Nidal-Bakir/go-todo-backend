-- name: SeederVersionAddValue :exec
INSERT INTO seeder_version (
  version
)
VALUES (
  $1
);

-- name: SeederVersionReadLatestAppliedVersion :one
SELECT
  version
FROM
  seeder_version
ORDER BY
  version DESC
LIMIT 1;
