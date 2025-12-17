package postgresRepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"subscription-service/internal/config"
	pb "subscription-service/proto/pkg/proto"
	"time"
)

var (
	ErrInvalidArgument = errors.New("invalid argument given")
	ErrAlreadyExists   = errors.New("subscription already exists")
	ErrInvalidDate     = errors.New("invalid date format")
	ErrNotFound        = errors.New("row not found")
	ErrDBUnavailable   = errors.New("database unavailable")
	ErrUnknown         = errors.New("unknown database error")
)

const (
	pingsAmount = 3
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func NewPool(config *config.Config) (*pgxpool.Pool, error) {
	ctx := context.Background()

	url := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		config.PostgresUser, config.PostgresPassword, config.PostgresHost, config.PostgresPort, config.PostgresDb)

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("newPool.New: %w", err)
	}

	err = tryToPing(ctx, pingsAmount, pool)
	if err != nil {
		return nil, fmt.Errorf("newPool.Ping: %w", err)
	}

	return pool, nil
}

func tryToPing(ctx context.Context, tries int, pool *pgxpool.Pool) error {
	var err error
	for i := 0; i < tries; i++ {
		err = pool.Ping(ctx)
		if err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return err
}

func (pr *PostgresRepository) Create(ctx context.Context, sub *pb.Subscription) (*pb.Subscription, error) {
	conn, err := pr.pool.Acquire(ctx)
	if err != nil {
		return nil, ErrDBUnavailable
	}
	defer conn.Release()

	row := conn.QueryRow(
		ctx,
		"insert into subscriptions (id, name, started, expire, price) values ($1, $2, $3, $4, $5) returning id, name, started, expire, price",
		sub.GetId(), sub.GetName(), sub.GetStartedAt(), sub.GetExpiration(), sub.GetPrice(),
	)
	var started, expire time.Time
	var id, name string
	var price int
	err = row.Scan(&id, &name, &started, &expire, &price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return nil, ErrAlreadyExists
			case pgerrcode.NotNullViolation:
				return nil, ErrInvalidDate
			default:
				return nil, ErrUnknown
			}
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}
		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}

		return nil, ErrUnknown
	}

	return &pb.Subscription{Id: id, Name: name, StartedAt: started.Format("2006-01-02"), Expiration: expire.Format("2006-01-02"), Price: int32(price)}, nil
}

func (pr *PostgresRepository) SelectByName(ctx context.Context, id, name string) (*pb.Subscription, error) {
	conn, err := pr.pool.Acquire(ctx)
	if err != nil {
		return nil, ErrDBUnavailable
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, "select * from subscriptions where id = $1 and name = $2", id, name)

	sub := pb.Subscription{}
	var startedAt, expiration time.Time
	err = row.Scan(&sub.Id, &sub.Name, &startedAt, &expiration, &sub.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, ErrUnknown
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}
		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}

		return nil, ErrUnknown
	}

	sub.StartedAt = startedAt.Format("2006-01-02")
	sub.Expiration = expiration.Format("2006-01-02")

	return &sub, nil
}

func (pr *PostgresRepository) Update(ctx context.Context, sub *pb.Subscription, oldName string) (*pb.Subscription, error) {
	conn, err := pr.pool.Acquire(ctx)
	if err != nil {
		return nil, ErrDBUnavailable
	}
	defer conn.Release()

	comm, err := conn.Exec(
		ctx,
		"update subscriptions set name = $1, started = $2, expire = $3, price = $4 where id = $5 and name = $6",
		sub.GetName(), sub.GetStartedAt(), sub.GetExpiration(), sub.GetPrice(), sub.GetId(), oldName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			//База данных не смогла преобразовать данные в нужный тип данных
			case pgerrcode.InvalidTextRepresentation:
				return nil, ErrInvalidArgument
			case pgerrcode.UniqueViolation:
				return nil, ErrAlreadyExists
			case pgerrcode.NotNullViolation:
				return nil, ErrInvalidDate
			default:
				return nil, ErrUnknown
			}
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}
		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}

		return nil, ErrUnknown
	}
	if comm.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return sub, nil

}

func (pr *PostgresRepository) Delete(ctx context.Context, id, name string) (bool, error) {
	conn, err := pr.pool.Acquire(ctx)
	if err != nil {
		return false, ErrDBUnavailable
	}
	defer conn.Release()

	comm, err := conn.Exec(ctx, "delete from subscriptions where id = $1 and name = $2", id, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			//База данных не смогла преобразовать данные в нужный тип данных
			case pgerrcode.InvalidTextRepresentation:
				return false, ErrInvalidArgument
			default:
				return false, ErrUnknown
			}
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return false, context.DeadlineExceeded
		}
		if errors.Is(err, context.Canceled) {
			return false, context.Canceled
		}
		return false, ErrUnknown
	}
	if comm.RowsAffected() == 0 {
		return false, ErrNotFound
	}

	return true, nil
}

func (pr *PostgresRepository) SelectAll(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error) {
	conn, err := pr.pool.Acquire(ctx)
	if err != nil {
		return nil, ErrDBUnavailable
	}
	defer conn.Release()

	rows, err := conn.Query(
		ctx,
		"SELECT * FROM subscriptions WHERE id = $1 AND started <= $3 AND expire >= $2",
		id, period.Start, period.End,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, ErrUnknown
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, ErrDBUnavailable
		}
	}
	defer rows.Close()

	var subs []*pb.Subscription
	for rows.Next() {
		var sub pb.Subscription
		var startedAt, expiration time.Time
		err = rows.Scan(&sub.Id, &sub.Name, &startedAt, &expiration, &sub.Price)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case pgerrcode.InvalidTextRepresentation:
					return nil, ErrInvalidArgument
				}
				return nil, ErrUnknown
			}

			if errors.Is(err, context.DeadlineExceeded) {
				return nil, context.DeadlineExceeded
			}
			if errors.Is(err, context.Canceled) {
				return nil, context.Canceled
			}

			return nil, ErrUnknown
		}

		sub.StartedAt = startedAt.Format("2006-01-02")
		sub.Expiration = expiration.Format("2006-01-02")

		subs = append(subs, &sub)
	}
	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, ErrUnknown
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}

		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}

		return nil, ErrUnknown
	}
	if len(subs) == 0 {
		return nil, ErrNotFound
	}

	return subs, nil
}

func (pr *PostgresRepository) Close() {
	pr.pool.Close()
}
