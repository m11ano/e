package e

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrCheckIsTxСoncurrentExec(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "25P02") {
		return true
	}
	return errors.Is(err, ErrTxСoncurrentExec)
}

func ErrConvertPgxToLogic(err error) (bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return true, NewErrorFrom(ErrStoreNoRows).Wrap(err)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "40001":
			return true, NewErrorFrom(ErrTxСoncurrentExec).Wrap(err)
		case pgErr.Code == "25P02":
			return true, NewErrorFrom(ErrTxСoncurrentExec).Wrap(err)
		case pgErr.Code == "23505":
			return true, NewErrorFrom(ErrStoreUniqueViolation).Wrap(err).SetData(pgErr.ColumnName)
		case pgErr.Code == "23503":
			return true, NewErrorFrom(ErrStoreForeignKeyViolation).Wrap(err).SetData(pgErr.ColumnName)
		case pgErr.Code == "23502":
			return true, NewErrorFrom(ErrStoreNotNullViolation).Wrap(err).SetData(pgErr.ColumnName)
		case pgErr.Code == "23514":
			return true, NewErrorFrom(ErrStoreCheckViolation).Wrap(err).SetData(pgErr.ConstraintName)
		case pgErr.Code == "23001":
			return true, NewErrorFrom(ErrStoreRestrictViolation).Wrap(err).SetData(pgErr.ConstraintName)
		case pgErr.Code == "23000":
			return true, NewErrorFrom(ErrStoreIntegrityViolation).Wrap(err).SetData(pgErr.ConstraintName)
		default:
			return false, NewErrorFrom(ErrInternal).Wrap(err)
		}
	}
	return false, err
}

func ErrConvertGRPCToLogic(err error) (bool, error) {

	if err == nil {
		return false, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return false, err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return true, NewErrorFrom(ErrBadRequest).Wrap(err).SetMessage(st.Message())
	case codes.Unauthenticated:
		return true, NewErrorFrom(ErrUnauthorized).Wrap(err).SetMessage(st.Message())
	case codes.PermissionDenied:
		return true, NewErrorFrom(ErrForbidden).Wrap(err).SetMessage(st.Message())
	case codes.NotFound:
		return true, NewErrorFrom(ErrNotFound).Wrap(err).SetMessage(st.Message())
	case codes.AlreadyExists, codes.Aborted:
		return true, NewErrorFrom(ErrConflict).Wrap(err).SetMessage(st.Message())
	case codes.FailedPrecondition, codes.OutOfRange:
		return true, NewErrorFrom(ErrUnprocessableEntity).Wrap(err).SetMessage(st.Message())
	case codes.Unimplemented, codes.DataLoss:
		return true, NewErrorFrom(ErrNotAcceptable).Wrap(err).SetMessage(st.Message())
	case codes.Unavailable:
		return true, NewErrorFrom(ErrServiceUnavailable).Wrap(err).SetMessage(st.Message())
	case codes.Internal, codes.Unknown:
		return true, NewErrorFrom(ErrInternal).Wrap(err).SetMessage(st.Message())
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Canceled:
		return true, NewErrorFrom(ErrGRPCCanceled).Wrap(err).SetMessage(st.Message())
	default:
		return true, NewErrorFrom(ErrInternal).Wrap(err).SetMessage("unknown internal error")
	}
}
