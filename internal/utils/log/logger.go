package log

import (
	"context"
	"log/slog"
)

func DbQueryCtx(ctx context.Context, err error, query string, vals interface{}) {
	slog.ErrorContext(ctx, "db_query_err",
		slog.Any("error", err),
		slog.String("query", query),
		slog.Any("vals", vals))

}

func ErrWriteResp(n int, err error) {
	slog.Info("error_write_resp_body", err)
}

func ErrReadBody(err error) {
	slog.Info("read_req_body_err", slog.Any("err", err))
}

func UnmarshalBodyErr(err error) {
	slog.Info("unmarshal_err", slog.Any("err", err))
}
