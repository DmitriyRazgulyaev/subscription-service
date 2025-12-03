package postgresRepository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"subscription-service/internal/config"
	"time"
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

func (pr *PostgresRepository) Create() {

}

func (pr *PostgresRepository) SelectByName() {

}

func (pr *PostgresRepository) Update() {

}

func (pr *PostgresRepository) Delete() {

}
