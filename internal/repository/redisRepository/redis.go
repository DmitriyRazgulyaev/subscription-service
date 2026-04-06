package redisRepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"strconv"
	"subscription-service/internal/config"
	pb "subscription-service/proto/pkg/proto"
	"time"
)

const (
	keyId         = "id"
	keyName       = "name"
	keyStartedAt  = "startedAt"
	keyExpiration = "expiration"
	keyPrice      = "price"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

func NewRedisClient(config config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr,
		MaxRetries:   config.RedisMaxRetries,
		DialTimeout:  config.RedisDialTimeout,
		ReadTimeout:  config.RedisTimeout,
		WriteTimeout: config.RedisTimeout,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil

}

func NewKey(id, subName string) string {
	return fmt.Sprintf("%s:%s", id, subName)
}

func (rr *RedisRepository) Create(ctx context.Context, sub *pb.Subscription) (*pb.Subscription, error) {
	key := NewKey(sub.GetId(), sub.GetName())

	_, err := rr.client.HSet(
		ctx, key,
		keyId, sub.GetId(),
		keyName, sub.GetName(),
		keyStartedAt, sub.GetStartedAt(),
		keyExpiration, sub.GetExpiration(),
		keyPrice, sub.GetPrice(),
	).Result()
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (rr *RedisRepository) SelectByName(ctx context.Context, id, name string) (*pb.Subscription, error) {
	key := NewKey(id, name)

	data, err := rr.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("hash does not exist")
	}

	price, err := strconv.Atoi(data[keyPrice])
	if err != nil {
		return nil, err
	}

	sub := &pb.Subscription{
		Id:         data[keyId],
		Name:       data[keyName],
		StartedAt:  data[keyStartedAt],
		Expiration: data[keyExpiration],
		Price:      int32(price),
	}

	return sub, nil
}

func (rr *RedisRepository) Update(ctx context.Context, subscription *pb.Subscription, oldName string) (*pb.Subscription, error) {
	updatedSub, err := rr.Create(ctx, subscription)
	if err != nil {
		return nil, err
	}

	return updatedSub, nil
}

func (rr *RedisRepository) Delete(ctx context.Context, id, name string) (bool, error) {
	key := NewKey(id, name)
	n, err := rr.client.Del(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}

	return true, nil
}

func (rr *RedisRepository) SelectAll(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error) {
	return nil, nil
}

func (rr *RedisRepository) SelectByExpiringDate(ctx context.Context, date string) ([]*pb.Subscription, error) {
	return nil, nil
}

func (rr *RedisRepository) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	if err := rr.client.Expire(ctx, key, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (rr *RedisRepository) Close() {
	rr.client.Close()
}
