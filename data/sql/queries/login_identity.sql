-- name: LoginIdentityCreateNewPasswordLoginIdentity :one
WITH new_identity AS (
 INSERT INTO login_identity (user_id, identity_type)
 VALUES ($1, $2)
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
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING id AS password_login_identity_id, (SELECT id AS login_identity_id FROM new_identity);


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
    u.role_id as user_role_id
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


-- name: LoginIdentityGetOIDCLoginIdentity :one
SELECT
    li.id AS login_identity_id,
    li.user_id,
    li.identity_type,
    li.is_primary,
    li.last_used_at,

    oli.id AS oidc_login_identity_id,

    oidc_data.id AS oidc_data_id,
    oidc_data.user_integration_id AS oidc_data_user_integration_id,
    oidc_data.sub AS oidc_data_sub,
    oidc_data.email AS oidc_data_email,
    oidc_data.iss AS oidc_data_iss,
    oidc_data.aud AS oidc_data_aud,
    oidc_data.given_name AS oidc_data_given_name,
    oidc_data.family_name AS oidc_data_family_name,
    oidc_data.name AS oidc_data_name,
    oidc_data.picture AS oidc_data_picture
FROM active_login_identity AS li
    JOIN active_oidc_login_identity oli
        ON li.id = oli.login_identity_id
    JOIN active_oidc_user_integration_data AS oidc_data
        ON oli.oidc_user_integration_data_id = oidc_data.id
WHERE li.identity_type = 'oidc'
    AND oidc_data.sub = @oidc_sub::text
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
  oud.id AS oidc_data_id,
  oud.sub AS oidc_data_sub,
  oud.email AS oidc_data_email,
  oud.iss AS oidc_data_issuer,
  oud.aud AS oidc_data_audience,
  oud.given_name AS oidc_data_given_name,
  oud.family_name AS oidc_data_family_name,
  oud.name AS oidc_data_name,
  oud.picture AS oidc_data_picture,
  
  -- oauth_provider
  op.name AS oauth_provider_name
  
FROM active_login_identity AS li
LEFT JOIN active_password_login_identity AS pli
  ON li.id = pli.login_identity_id
LEFT JOIN active_guest_login_identity AS gli
  ON li.id = gli.login_identity_id
LEFT JOIN active_oidc_login_identity AS oli
  ON li.id = oli.login_identity_id
LEFT JOIN active_oidc_user_integration_data AS oud
  ON oli.oidc_user_integration_data_id = oud.id
LEFT JOIN user_integration AS ui
  ON oud.user_integration_id = ui.id
LEFT JOIN oauth_integration AS oi
  ON ui.oauth_integration_id = oi.id
LEFT JOIN oauth_connection AS oc
  ON oi.oauth_connection_id = oc.id
LEFT JOIN oauth_provider AS op
  ON oc.provider_id = op.id
  
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
SELECT COUNT(*) FROM active_oidc_user_integration_data
WHERE email = $1;
