package util

import "errors"

const (
	DefaultFailedSqlTxInitMsg                    string = "An error occurred while initiating an SQL transaction."
	DefaultFailedSqlQueryMsg                     string = "An error occurred while executing a SQL query."
	DefaultFailedScanningToPostgresqlDataTypeMsg string = "An error occurred while scanning the value into a PostgreSQL data type."
	DefaultFailedFetchingXMRDaemonMsg            string = "An error occurred while fetching."
)

var (
	invalidProtoBufCoinTypeErr error = errors.New("invalid protoBuf coin type")
	invalidDbCoinTypeErr       error = errors.New("invalid db coin type")
	invalidDbStatusTypeErr     error = errors.New("invalid db status type")
)
