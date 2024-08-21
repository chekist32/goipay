-- name: CreateCryptoAddress :one
INSERT INTO crypto_addresses(address, coin, is_occupied, user_id) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin :one
UPDATE crypto_addresses SET is_occupied = true
WHERE address = (
    SELECT address FROM crypto_addresses AS ca
    WHERE ca.user_id = $1 AND ca.coin = $2 AND ca.is_occupied = false 
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: UpdateIsOccupiedByCryptoAddress :one
UPDATE crypto_addresses 
SET is_occupied = $2
WHERE address = $1
RETURNING *;

-- name: DeleteAllCryptoAddressByUserIdAndCoin :many
DELETE FROM crypto_addresses 
WHERE user_id = $1 AND coin = $2
RETURNING *;


