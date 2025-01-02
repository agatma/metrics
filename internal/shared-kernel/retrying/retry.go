// Package retrying.
package retrying

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"metrics/internal/server/logger"
	"time"

	"github.com/avast/retry-go"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// addTime is the additional delay added to the initial delay for subsequent retries.
const addTime = 2

// Attempts specifies the maximum number of retry attempts.
const Attempts uint = 3

// DelayType calculates the duration for the next retry attempt.
func DelayType(n uint, _ error, config *retry.Config) time.Duration {
	switch n {
	case 0:
		return 1 * time.Second
	case 1:
		return (1 + addTime) * time.Second
	default:
		return (1 + addTime + addTime) * time.Second
	}
}

// OnRetry logs information about the retry attempt.
func OnRetry(n uint, err error) {
	logger.Log.Error(fmt.Sprintf(`%d %s`, n, err.Error()))
}

// Transaction represents a database transaction interface.
type Transaction interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// ExecContext executes a SQL query using the provided transaction and retry logic.
func ExecContext(ctx context.Context, tx Transaction, query string, args ...any) error {
	var originalErr error
	err := retry.Do(
		func() error {
			_, originalErr := tx.ExecContext(
				ctx,
				query,
				args...,
			)
			if originalErr != nil {
				return fmt.Errorf("%w", originalErr)
			}
			return nil
		},
		retry.RetryIf(func(err error) bool {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
				return true
			}
			return false
		}),
		retry.Attempts(Attempts),
		retry.DelayType(DelayType),
		retry.OnRetry(OnRetry),
	)
	if err != nil {
		logger.Log.Error("retryError", zap.Error(err), zap.Error(originalErr))
		return fmt.Errorf("%w", originalErr)
	}
	return originalErr
}
