package v1

import (
	"context"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkIfUserExistsString(ctx context.Context, log *zerolog.Logger, q *db.Queries, userId string) error {
	userIdUUID, err := util.StringToPgUUID(userId)
	if err != nil {
		log.Err(err).Msg("An error occurred while converting the string to the PostgreSQL UUID data type.")
		return status.Error(codes.InvalidArgument, "Invalid userId")
	}

	return checkIfUserExistsUUID(ctx, log, q, *userIdUUID)
}

func checkIfUserExistsUUID(ctx context.Context, log *zerolog.Logger, q *db.Queries, userId pgtype.UUID) error {
	res, err := q.UserExistsById(ctx, userId)
	if err != nil {
		log.Err(err).Str("queryName", "UserExistsById").Msg(util.DefaultFailedSqlQueryMsg)
		return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	if !res {
		return status.Error(codes.InvalidArgument, "Invalid userId")
	}

	return nil
}
