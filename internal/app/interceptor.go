package app

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type RequestLoggingInterceptor struct {
	log *zerolog.Logger
}

func (i *RequestLoggingInterceptor) Intercepte(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	i.log.Info().Msgf("PRE %s", info.FullMethod)

	res, err := handler(ctx, req)
	if err != nil {
		i.log.Info().Err(err).Str("status", "failure").Msgf("POST %s", info.FullMethod)
	} else {
		i.log.Info().Str("status", "success").Msgf("POST %s", info.FullMethod)
	}

	return res, err
}

func NewRequestLoggingInterceptor(log *zerolog.Logger) *RequestLoggingInterceptor {
	return &RequestLoggingInterceptor{log: log}
}
