package v1

import (
	"context"

	"github.com/chekist32/go-monero/utils"
	"github.com/chekist32/goipay/internal/db"
	pb_v1 "github.com/chekist32/goipay/internal/pb/v1"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGrpc struct {
	dbConnPool *pgxpool.Pool
	log        *zerolog.Logger
	pb_v1.UnimplementedUserServiceServer
}

func (u *UserGrpc) createUser(ctx context.Context, q *db.Queries, in *pb_v1.RegisterUserRequest) (*pgtype.UUID, error) {
	// With userId in the request
	if in.UserId == nil {
		userId, err := q.CreateUser(ctx)
		if err != nil {
			u.log.Err(err).Str("queryName", "CreateUser").Msg(util.DefaultFailedSqlQueryMsg)
			return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
		}

		return &userId, err
	}

	// Without userId in the request
	userIdReq, err := util.StringToPgUUID(*in.UserId)
	if err != nil {
		u.log.Err(err).Msg("invalid userId (uuid)")
		return nil, status.Error(codes.InvalidArgument, "invalid userId (uuid)")
	}

	userId, err := q.CreateUserWithId(ctx, *userIdReq)
	if err != nil {
		u.log.Err(err).Str("queryName", "CreateUserWithId").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	return &userId, err
}

func (u *UserGrpc) RegisterUser(ctx context.Context, in *pb_v1.RegisterUserRequest) (*pb_v1.RegisterUserResponse, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, u.dbConnPool)
	if err != nil {
		u.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlTxInitMsg)
	}

	userId, err := u.createUser(ctx, q, in)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{UserID: *userId})
	if err != nil {
		tx.Rollback(ctx)
		u.log.Err(err).Str("queryName", "CreateUser").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	tx.Commit(ctx)

	return &pb_v1.RegisterUserResponse{UserId: util.PgUUIDToString(*userId)}, nil
}

func (u *UserGrpc) handleXmrCryptoDataUpdate(ctx context.Context, q *db.Queries, in *pb_v1.XmrKeysUpdateRequest, cryptData *db.CryptoDatum) error {
	_, err := utils.NewPrivateKey(in.PrivViewKey)
	if err != nil {
		u.log.Err(err).Msg("An error occurred while creating the XMR private view key.")
		return status.Error(codes.InvalidArgument, "invalid private view key")
	}
	_, err = utils.NewPublicKey(in.PubSpendKey)
	if err != nil {
		u.log.Err(err).Msg("An error occurred while creating the XMR public spend key.")
		return status.Error(codes.InvalidArgument, "invalid public spend key")
	}

	_, err = q.DeleteAllCryptoAddressByUserIdAndCoin(ctx, db.DeleteAllCryptoAddressByUserIdAndCoinParams{Coin: db.CoinTypeXMR, UserID: cryptData.UserID})
	if err != nil {
		u.log.Err(err).Str("queryName", "DeleteAllCryptoAddressByUserIdAndCoin").Msg(util.DefaultFailedSqlQueryMsg)
		return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	if !cryptData.XmrID.Valid {
		xmrData, err := q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: in.PrivViewKey, PubSpendKey: in.PubSpendKey})
		if err != nil {
			u.log.Err(err).Str("queryName", "CreateXMRCryptoData").Msg(util.DefaultFailedSqlQueryMsg)
			return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
		}
		_, err = q.SetXMRCryptoDataByUserId(ctx, db.SetXMRCryptoDataByUserIdParams{UserID: cryptData.UserID, XmrID: xmrData.ID})
		if err != nil {
			u.log.Err(err).Str("queryName", "SetXMRCryptoDataByUserId").Msg(util.DefaultFailedSqlQueryMsg)
			return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
		}
		return nil
	}
	_, err = q.UpdateKeysXMRCryptoDataById(ctx, db.UpdateKeysXMRCryptoDataByIdParams{ID: cryptData.XmrID, PrivViewKey: in.PrivViewKey, PubSpendKey: in.PubSpendKey})
	if err != nil {
		return status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	return nil
}

func (u *UserGrpc) UpdateCryptoKeys(ctx context.Context, in *pb_v1.UpdateCryptoKeysRequest) (*pb_v1.UpdateCryptoKeysResponse, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, u.dbConnPool)
	if err != nil {
		u.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlTxInitMsg)
	}

	userId, err := util.StringToPgUUID(in.UserId)
	if err != nil {
		tx.Rollback(ctx)
		u.log.Err(err).Msg("An error occurred while converting the string to the PostgreSQL UUID data type.")
		return nil, status.Error(codes.InvalidArgument, "invalid userId")
	}

	if err := checkIfUserExistsUUID(ctx, u.log, q, *userId); err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	cryptData, err := q.FindCryptoDataByUserId(ctx, *userId)
	if err != nil {
		tx.Rollback(ctx)
		u.log.Err(err).Str("queryName", "FindCryptoDataByUserId").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	if in.XmrReq != nil {
		if err := u.handleXmrCryptoDataUpdate(ctx, q, in.XmrReq, &cryptData); err != nil {
			tx.Rollback(ctx)
			u.log.Err(err).Msg("")
			return nil, err
		}
	}

	tx.Commit(ctx)

	return &pb_v1.UpdateCryptoKeysResponse{}, nil
}

func (u *UserGrpc) GetCryptoKeys(ctx context.Context, in *pb_v1.GetCryptoKeysRequest) (*pb_v1.GetCryptoKeysResponse, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, u.dbConnPool)
	if err != nil {
		u.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlTxInitMsg)
	}

	userId, err := util.StringToPgUUID(in.UserId)
	if err != nil {
		tx.Rollback(ctx)
		u.log.Err(err).Msg("An error occurred while converting the string to the PostgreSQL UUID data type.")
		return nil, status.Error(codes.InvalidArgument, "invalid userId")
	}

	if err := checkIfUserExistsUUID(ctx, u.log, q, *userId); err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	cryptoKeys, err := q.FindCryptoKeysByUserId(ctx, *userId)
	if err != nil {
		tx.Rollback(ctx)
		u.log.Err(err).Str("queryName", "FindCryptoKeysByUserId").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, status.Error(codes.Internal, util.DefaultFailedSqlQueryMsg)
	}

	tx.Commit(ctx)

	return &pb_v1.GetCryptoKeysResponse{
		XmrKeys: &pb_v1.XmrKeys{
			PrivViewKey: cryptoKeys.PrivViewKey,
			PubSpendKey: cryptoKeys.PubSpendKey,
		},
	}, nil
}

func NewUserGrpc(dbConnPool *pgxpool.Pool, log *zerolog.Logger) *UserGrpc {
	return &UserGrpc{dbConnPool: dbConnPool, log: log}
}
