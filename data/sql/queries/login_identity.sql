-- name: LoginIdentityCreateNewUserAndPasswordLoginIdentity :one
WITH new_user AS (
  INSERT INTO users (
    username,
    profile_image,
    first_name,
    last_name,
    role_name
  )
  VALUES (
    @user_username::text,
    sqlc.narg(user_profile_image)::text,
    @user_first_name::text,
    sqlc.narg(user_last_name)::text,
    sqlc.narg(user_role_name)::text
  )
  RETURNING *
),
new_identity AS (
  INSERT INTO login_identity (
    user_id,
    identity_type
  )
  VALUES (
    (SELECT id FROM new_user),
    @identity_type::text
  )
  RETURNING id
),
final_insert AS (
  INSERT INTO password_login_identity (
    login_identity_id,
    email,
    phone,
    hashed_pass,
    pass_salt,
    verified_at
  )
  VALUES (
    (SELECT id FROM new_identity),
    sqlc.narg(password_email)::text,
    sqlc.narg(password_phone)::text,
    @password_hashed_pass::text,
    @password_pass_salt::text,
    @password_verified_at::timestamptz
  )
)
SELECT * FROM new_user;

-- name: LoginIdentityCreateNewPasswordLoginIdentity :one
WITH new_identity AS (
 INSERT INTO login_identity (
    user_id,
    identity_type
 )
 VALUES (
    @identity_user_id::int,
    @identity_type::text
 )
 RETURNING id
)
INSERT INTO password_login_identity (
    login_identity_id,
    email,
    phone,
    hashed_pass,
    pass_salt,
    verified_at
)
VALUES (
    (SELECT id FROM new_identity),
    sqlc.narg(password_email)::text,
    sqlc.narg(password_phone)::text,
    @password_hashed_pass::text,
    @password_pass_salt::text,
    @password_verified_at::timestamptz
)
RETURNING id AS password_login_identity_id, (SELECT id AS login_identity_id FROM new_identity);


-- name: LoginIdentityGetOIDCDataBySub :one
SELECT
    u.id AS user_id,
    u.username AS user_username,
    u.profile_image AS user_profile_image,
    u.first_name AS user_first_name,
    u.middle_name AS user_middle_name,
    u.last_name AS user_last_name,
    u.blocked_at AS user_blocked_at,
    u.blocked_until AS user_blocked_until,
    u.created_at AS user_created_at,
    u.updated_at AS user_updated_at,
    u.role_name as user_role_name,
    li.id AS login_identity_id,
    od.provider_name AS oauth_provider_name,
    od.id AS oidc_data_id

from active_oidc_data AS od
JOIN active_oidc_login_identity AS oli
    ON od.id = oli.oidc_data_id
JOIN active_login_identity AS li
    ON oli.login_identity_id = li.id
JOIN not_deleted_users AS u
    ON li.user_id = u.id

WHERE od.sub = @oidc_sub::text
    AND li.identity_type = 'oidc'
    AND od.provider_name = @oidc_provider_name::text
LIMIT 1;


-- name: LoginIdentityGetPasswordLoginIdentity :one
SELECT
    li.id AS login_identity_id,
    li.user_id,
    li.identity_type,
    li.is_primary AS login_identity_is_primary,
    li.last_used_at AS login_identity_last_used_at,

    pli.id AS password_login_identity_id,
    pli.email ,
    pli.phone,
    pli.hashed_pass,
    pli.pass_salt,
    pli.verified_at
FROM active_login_identity AS li
    JOIN active_password_login_identity pli
        ON li.id = pli.login_identity_id
WHERE li.identity_type = @identity_type::text
    AND (
      (@identity_type::text = 'email' AND pli.email = @identity_value::text)
      OR
      (@identity_type::text = 'phone' AND pli.phone = @identity_value::text)
    )
LIMIT 1;


-- name: LoginIdentityGetPasswordLoginIdentityWithUser :one
SELECT
    li.id AS login_identity_id,
    li.user_id,
    li.identity_type,
    li.is_primary,
    li.last_used_at,

    pli.id AS password_login_identity_id,
    pli.email,
    pli.phone,
    pli.hashed_pass,
    pli.pass_salt,
    pli.verified_at,

    u.id as user_id,
    u.username as user_username,
    u.profile_image as user_profile_image,
    u.first_name as user_first_name,
    u.middle_name as user_middle_name,
    u.last_name as user_last_name,
    u.blocked_at as user_blocked_at,
    u.blocked_until as user_blocked_until,
    u.role_name as user_role_name
FROM not_deleted_users AS u
    JOIN active_login_identity AS li
        ON u.id = li.user_id
    JOIN active_password_login_identity pli
        ON li.id = pli.login_identity_id
WHERE li.identity_type = @identity_type::text
    AND (
      (@identity_type::text = 'email' AND pli.email = @identity_value::text)
      OR
      (@identity_type::text = 'phone' AND pli.phone = @identity_value::text)
    )
LIMIT 1;


-- name: LoginIdentityGetAllByUserId :many
SELECT
  li.id AS login_identity_id,
  li.user_id AS login_identity_user_id,
  li.identity_type AS login_identity_identity_type,
  li.is_primary AS login_identity_is_primary,
  li.last_used_at AS login_identity_last_used_at,

  -- Password-based
  pli.id AS password_id,
  pli.email  AS password_email,
  pli.phone  AS password_phone,
  pli.hashed_pass  AS password_hashed_pass,
  pli.pass_salt  AS password_pass_salt,
  pli.verified_at AS password_verified_at,

  -- Guest-based
  gli.id AS guest_id,
  gli.device_id AS guest_device_id,

  -- OIDC-based
  oidc_data.id AS oidc_data_id,
  oidc_data.sub AS oidc_data_sub,
  oidc_data.email AS oidc_data_email,
  oidc_data.iss AS oidc_data_issuer,
  oidc_data.aud AS oidc_data_audience,
  oidc_data.given_name AS oidc_data_given_name,
  oidc_data.family_name AS oidc_data_family_name,
  oidc_data.name AS oidc_data_name,
  oidc_data.picture AS oidc_data_picture,
  oidc_data.provider_name AS oauth_provider_name

FROM active_login_identity AS li
LEFT JOIN active_password_login_identity AS pli
  ON li.id = pli.login_identity_id
LEFT JOIN active_guest_login_identity AS gli
  ON li.id = gli.login_identity_id
LEFT JOIN active_oidc_login_identity AS oli
  ON li.id = oli.login_identity_id
LEFT JOIN active_oidc_data AS oidc_data
  ON oli.oidc_data_id = oidc_data.id

WHERE li.user_id = $1
ORDER BY li.is_primary DESC, li.last_used_at DESC;



-- name: LoginIdentityGetAllPasswordLoginIdentitiesByUserId :many
SELECT
  li.id AS login_identity_id,
  li.user_id AS login_identity_user_id,
  li.identity_type AS login_identity_identity_type,
  li.is_primary AS login_identity_is_primary,
  li.last_used_at AS login_identity_last_used_at,

  -- Password-based
  pli.id AS password_id,
  pli.email  AS password_email,
  pli.phone  AS password_phone,
  pli.hashed_pass  AS password_hashed_pass,
  pli.pass_salt  AS password_pass_salt,
  pli.verified_at AS password_verified_at
FROM active_login_identity AS li
LEFT JOIN active_password_login_identity AS pli
  ON li.id = pli.login_identity_id
WHERE li.user_id = $1
ORDER BY li.is_primary DESC, li.last_used_at DESC;


-- name: LoginIdentityChangePasswordLoginIdentityByUserId :exec
UPDATE password_login_identity pli
SET
    hashed_pass = $2,
    pass_salt = $3
FROM active_login_identity li
WHERE pli.login_identity_id = li.id
  AND li.user_id = $1;


-- name: LoginIdentityIsEmailUsed :one
SELECT COUNT(*) FROM active_password_login_identity WHERE email = $1;

-- name: LoginIdentityIsPhoneUsed :one
SELECT COUNT(*) FROM active_password_login_identity WHERE phone = $1;

-- name: LoginIdentityIsOidcEmailUsed :one
SELECT COUNT(*) FROM active_oidc_data WHERE email = $1;


-- name: LoginIdentityUpdateLastUsedAtToNow :exec
UPDATE login_identity SET
last_used_at = NOW()
WHERE id = @id;


-- WITH new_user AS (
--     INSERT INTO users (username, first_name)
--     VALUES (concat('guest_', gen_random_uuid()), 'Guest')
--     RETURNING id
-- ),
-- new_identity AS (
--     INSERT INTO login_identity (user_id)
--     SELECT id FROM new_user
--     RETURNING id, user_id
-- ),
-- new_guest AS (
--     INSERT INTO guest_login_option (login_identity_id, device_id)
--     SELECT id, $1 FROM new_identity
--     RETURNING login_identity_id
-- ),
-- new_session AS (
--     INSERT INTO session (token, originated_from, used_installation, expires_at)
--     VALUES (
--         $2,                        -- token
--         (SELECT login_identity_id FROM new_guest),
--         $3,                        -- installation_id
--         NOW() + INTERVAL '30 days' -- expires
--     )
--     RETURNING token
-- )
-- SELECT token FROM new_session;


-- name: LoginIdentityCreateNewUserAndOIDCLoginIdentity :one
WITH new_user AS (
  INSERT INTO users (
    username,
    profile_image,
    first_name,
    last_name,
    role_name
  )
  VALUES (
    @user_username::text,
    sqlc.narg(user_profile_image)::text,
    @user_first_name::text,
    sqlc.narg(user_last_name)::text,
    sqlc.narg(user_role_name)::text
  )
  RETURNING id AS user_id, username, profile_image, first_name, middle_name, last_name, created_at, updated_at, blocked_at, blocked_until, deleted_at, role_name
),
new_identity AS (
  INSERT INTO login_identity (
    user_id,
    identity_type
  )
  VALUES (
    (SELECT user_id FROM new_user),
    'oidc'
  )
  RETURNING id
),
oauth_provider_record AS (
    SELECT
        @oauth_provider_name::text AS provider_name,
        @oauth_provider_is_oidc_capable::bool AS is_oidc_capable
),
oauth_provider_record_merge_op AS (
    MERGE INTO oauth_provider AS target
    USING oauth_provider_record AS r
    ON target.name = r.provider_name AND target.is_oidc_capable = r.is_oidc_capable
    WHEN NOT MATCHED THEN
        INSERT (name, is_oidc_capable)
        VALUES (r.provider_name, r.is_oidc_capable)
),
oauth_connection_record AS (
    SELECT
        (SELECT provider_name from oauth_provider_record) AS provider_name,
        @oauth_scopes::text[] AS scopes
),
oauth_connection_record_merge_op AS (
    MERGE INTO oauth_connection AS target
    USING oauth_connection_record AS r
    ON target.provider_name = r.provider_name AND target.scopes = r.scopes
    WHEN NOT MATCHED THEN
        INSERT (provider_name, scopes)
        VALUES (r.provider_name, r.scopes)
    RETURNING target.*
),
oauth_connection_row AS (
    SELECT id, provider_name, scopes, created_at, updated_at, deleted_at FROM oauth_connection_record_merge_op
    UNION ALL
    SELECT id, provider_name, scopes, created_at, updated_at, deleted_at from oauth_connection
        WHERE provider_name = (SELECT provider_name from oauth_provider_record)
            AND scopes = (SELECT scopes from oauth_connection_record)
),
new_oauth_integration AS (
    INSERT INTO oauth_integration (
        oauth_connection_id,
        integration_type
    )
    VALUES (
        (SELECT id FROM oauth_connection_row),
        'user'
    )
    RETURNING id
),
new_oauth_token AS (
	INSERT INTO oauth_token (
	    oauth_integration_id,
	    access_token,
	    refresh_token,
	    token_type,
	    expires_at,
	    issued_at
	)
	SELECT
	    (SELECT id FROM new_oauth_integration),
	    sqlc.narg(oauth_access_token)::text,
	    sqlc.narg(oauth_refresh_token)::text,
	    sqlc.narg(oauth_token_type)::text,
	    @oauth_token_expires_at::timestamp,
	    @oauth_token_issued_at::timestamp
	WHERE (
	    sqlc.narg(oauth_access_token)::text IS NOT NULL AND sqlc.narg(oauth_access_token)::text <> ''
	) OR (
	    sqlc.narg(oauth_refresh_token)::text IS NOT NULL AND sqlc.narg(oauth_refresh_token)::text <> ''
	)
),
new_user_integration AS (
    INSERT INTO user_integration (
        oauth_integration_id,
        user_id
    )
    VALUES (
        (SELECT id FROM new_oauth_integration),
        (SELECT user_id FROM new_user)
    )
    RETURNING id
),
new_oidc_data AS (
    INSERT INTO oidc_data (
        provider_name,
        sub,
        email,
        iss,
        aud,
        given_name,
        family_name,
        name,
        picture
    )
    VALUES (
        (SELECT provider_name from oauth_provider_record),
        @oidc_sub::text,
        sqlc.narg(oidc_email)::text,
        @oidc_iss::text,
        @oidc_aud::text,
        sqlc.narg(oidc_given_name)::text,
        sqlc.narg(oidc_family_name)::text,
        sqlc.narg(oidc_name)::text,
        sqlc.narg(oidc_picture)::text
    )
    RETURNING id
),
new_oidc_login_identity AS (
    INSERT INTO oidc_login_identity (
        login_identity_id,
        oidc_data_id
    )
    VALUES (
        (SELECT id FROM new_identity),
        (SELECT id FROM new_oidc_data)
    )
)
SELECT u.user_id, u.username, u.profile_image, u.first_name, u.middle_name, u.last_name, u.created_at, u.updated_at, u.blocked_at, u.blocked_until, u.deleted_at, u.role_name, i.id AS new_login_identity_id FROM new_user AS u, new_identity AS i;
