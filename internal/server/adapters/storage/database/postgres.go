// Package database provides implementations of the MetricStorage interface
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"metrics/internal/server/core/domain"
	"metrics/internal/server/logger"
	"metrics/internal/shared-kernel/retrying"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type MetricStorage struct {
	db *sqlx.DB
}

func NewStorage(cfg *Config) (*MetricStorage, error) {
	db, err := sqlx.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database %w", err)
	}
	return &MetricStorage{db: db}, migrate(db)
}

func (s *MetricStorage) GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error) {
	var (
		delta sql.NullInt64
		value sql.NullFloat64
	)
	row := s.db.QueryRowContext(
		ctx,
		`select delta, value from metrics where name=$1 and type=$2 ORDER BY created_at DESC LIMIT 1;`,
		mName,
		mType,
	)
	if row.Err() != nil {
		return nil, fmt.Errorf("%w", row.Err())
	}
	if err := row.Scan(&delta, &value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("%w", err)
	}
	switch mType {
	case domain.Gauge:
		return &domain.Metric{ID: mName, MType: mType, Value: &value.Float64}, nil
	case domain.Counter:
		return &domain.Metric{ID: mName, MType: mType, Delta: &delta.Int64}, nil
	default:
		return nil, domain.ErrIncorrectMetricType
	}
}

func (s *MetricStorage) SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error) {
	switch m.MType {
	case domain.Gauge:
		err := retrying.ExecContext(
			ctx,
			s.db,
			`INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3)`,
			m.ID, m.MType, *m.Value,
		)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	case domain.Counter:
		current, err := s.GetMetric(ctx, m.MType, m.ID)
		if err != nil {
			if !errors.Is(err, domain.ErrItemNotFound) {
				return nil, fmt.Errorf("%w", err)
			}
		} else {
			*m.Delta += *current.Delta
		}
		err = retrying.ExecContext(
			ctx,
			s.db,
			`INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)`,
			m.ID, m.MType, *m.Delta,
		)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	default:
		return nil, domain.ErrIncorrectMetricType
	}
	return m, nil
}

func (s *MetricStorage) SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	for _, m := range metrics {
		switch m.MType {
		case domain.Gauge:
			err = retrying.ExecContext(
				ctx,
				s.db,
				`INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3)`,
				m.ID, m.MType, *m.Value,
			)
			if err != nil {
				if txErr := tx.Rollback(); txErr != nil {
					if !errors.Is(txErr, sql.ErrTxDone) {
						logger.Log.Error("failed to rollback the transaction", zap.Error(txErr))
					}
				}
				return nil, fmt.Errorf("%w", err)
			}

		case domain.Counter:
			current, err := s.GetMetric(ctx, m.MType, m.ID)
			if err != nil {
				if !errors.Is(err, domain.ErrItemNotFound) {
					return nil, fmt.Errorf("%w", err)
				}
			} else {
				*m.Delta += *current.Delta
			}
			err = retrying.ExecContext(
				ctx,
				s.db,
				`INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)`,
				m.ID, m.MType, *m.Delta,
			)
			if err != nil {
				if txErr := tx.Rollback(); txErr != nil {
					if !errors.Is(txErr, sql.ErrTxDone) {
						logger.Log.Error("failed to rollback the transaction", zap.Error(txErr))
					}
				}
				return nil, fmt.Errorf("%w", err)
			}
		default:
			return nil, domain.ErrIncorrectMetricType
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction %w", err)
	}
	return metrics, nil
}

func (s *MetricStorage) GetAllMetrics(ctx context.Context) (domain.MetricsList, error) {
	metrics := make(domain.MetricsList, 0)
	rows, err := s.db.QueryContext(ctx,
		`SELECT t1.name, t1.type, m.delta, m.value 
		    FROM (select name, type, MAX(created_at) as created_at from metrics group by name, type) AS t1
			LEFT JOIN metrics AS m ON t1.name = m.name AND t1.type=m.type AND t1.created_at = m.created_at;`,
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	for rows.Next() {
		var (
			m     domain.Metric
			delta sql.NullInt64
			value sql.NullFloat64
		)

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		switch m.MType {
		case domain.Gauge:
			m.Value = &value.Float64
		case domain.Counter:
			m.Delta = &delta.Int64
		default:
			return nil, domain.ErrIncorrectMetricType
		}

		metrics = append(metrics, m)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("error occurred during closing rows", zap.Error(err))
		}
	}()
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return metrics, nil
}

func (s *MetricStorage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database %w", err)
	}
	return nil
}
