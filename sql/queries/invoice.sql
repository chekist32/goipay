-- name: CreateInvoice :one
INSERT INTO invoices(
    crypto_address,
    coin,
    required_amount, 
    confirmations_required,
    expires_at,
    user_id) 
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;


-- name: FindAllInvoicesByIds :many
SELECT * FROM invoices
WHERE id = ANY($1::uuid[]);
-- name: FindAllPendingInvoices :many
SELECT * FROM invoices
WHERE status IN ('PENDING', 'PENDING_MEMPOOL');


-- name: ConfirmInvoiceById :one
UPDATE invoices
SET status = 'CONFIRMED',
    confirmed_at = timezone('UTC', now())
WHERE id = $1
RETURNING *;

-- name: ConfirmInvoiceStatusMempoolById :one
UPDATE invoices
SET actual_amount = $2,
    status = 'PENDING_MEMPOOL',
    tx_id = $3
WHERE id = $1
RETURNING *;

-- name: ExpireInvoiceById :one
UPDATE invoices
SET status = 'EXPIRED'
WHERE id = $1
RETURNING *;

-- name: ShiftExpiresAtForNonConfirmedInvoices :many
UPDATE invoices
SET expires_at = timezone('UTC', now()) + INTERVAL '5 minute'
WHERE status IN ('PENDING', 'PENDING_MEMPOOL') AND (expires_at - timezone('UTC', now()) < INTERVAL '5 minutes')
RETURNING *;