-- name: FindCryptoCacheByCoin :one
SELECT * FROM crypto_cache
WHERE coin = $1;


-- name: UpdateCryptoCacheByCoin :one
UPDATE crypto_cache 
SET last_synced_block_height = $2,
    synced_timestamp = timezone('UTC', now())
WHERE coin = $1
RETURNING *;