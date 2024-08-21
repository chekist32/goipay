-- name: CreateCryptoData :one
INSERT INTO crypto_data(xmr_id, user_id) VALUES ($1, $2)
RETURNING *;

-- name: FindCryptoDataByUserId :one
SELECT * FROM crypto_data 
WHERE user_id = $1;

-- name: SetXMRCryptoDataByUserId :one
UPDATE crypto_data
SET xmr_id = $2 
WHERE user_id = $1
RETURNING *;


-- XMR
-- name: CreateXMRCryptoData :one
INSERT INTO xmr_crypto_data(priv_view_key, pub_spend_key) VALUES ($1, $2)
RETURNING *;

-- name: FindKeysAndLockXMRCryptoDataById :one
SELECT priv_view_key, pub_spend_key
FROM xmr_crypto_data
WHERE id = $1
FOR SHARE;

-- name: UpdateKeysXMRCryptoDataById :one
UPDATE xmr_crypto_data
SET priv_view_key = $2,
    pub_spend_key = $3,
    last_major_index = 0,
    last_minor_index = 0
WHERE id = $1
RETURNING *;

-- name: FindIndicesAndLockXMRCryptoDataById :one
SELECT last_major_index, last_minor_index 
FROM xmr_crypto_data
WHERE id = $1
FOR UPDATE;

-- name: UpdateIndicesXMRCryptoDataById :one
UPDATE xmr_crypto_data
SET last_major_index = $2,
    last_minor_index = $3
WHERE id = $1
RETURNING *;