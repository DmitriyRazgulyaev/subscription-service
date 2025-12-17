package cachedDbRepository

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"subscription-service/internal/repository/redisRepository"
	"subscription-service/internal/service"
	pb "subscription-service/proto/pkg/proto"
	"time"
)

const (
	defaultTTL = time.Second * 30
)

type CachedDbRepository struct {
	dbRepo    service.SubscriptionRepository
	redisRepo redisRepository.RedisRepository
}

func NewCachedPostgresRepository(
	dbRepo service.SubscriptionRepository,
	redisRepo *redisRepository.RedisRepository,
) *CachedDbRepository {
	return &CachedDbRepository{
		dbRepo:    dbRepo,
		redisRepo: *redisRepo,
	}
}

func (cdr *CachedDbRepository) Create(ctx context.Context, subscription *pb.Subscription) (*pb.Subscription, error) {
	sub, err := cdr.redisRepo.Create(ctx, subscription)
	if err != nil {
		return nil, err
	}

	key := redisRepository.NewKey(sub.GetId(), sub.GetName())
	err = cdr.redisRepo.SetTTL(ctx, key, defaultTTL)
	if err != nil {
		return nil, err
	}

	sub, err = cdr.dbRepo.Create(ctx, subscription)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (cdr *CachedDbRepository) SelectByName(ctx context.Context, id, name string) (*pb.Subscription, error) {
	selectedSub, err := cdr.redisRepo.SelectByName(ctx, id, name)
	if err == nil {
		return selectedSub, nil
	}
	if !errors.Is(err, redis.Nil) {
		log.Println(err)
		//TODO сделать нормальное логирование
	}

	selectedSub, err = cdr.dbRepo.SelectByName(ctx, id, name)
	if err != nil {
		return nil, err
	}
	_, err = cdr.redisRepo.Create(ctx, selectedSub)
	if err != nil {
		log.Println(err)
	}
	err = cdr.redisRepo.SetTTL(ctx, redisRepository.NewKey(selectedSub.GetId(), selectedSub.GetName()), defaultTTL)
	if err != nil {
		return nil, err
	}

	return selectedSub, nil
}

func (cdr *CachedDbRepository) Update(ctx context.Context, sub *pb.Subscription, oldName string) (*pb.Subscription, error) {
	sub, err := cdr.redisRepo.Create(ctx, sub)
	if err != nil {
		return nil, err
	}

	key := redisRepository.NewKey(sub.GetId(), sub.GetName())
	err = cdr.redisRepo.SetTTL(ctx, key, defaultTTL)
	if err != nil {
		return nil, err
	}

	updatedSub, err := cdr.dbRepo.Update(ctx, sub, oldName)
	if err != nil {
		return nil, err
	}

	return updatedSub, nil
}

func (cdr *CachedDbRepository) Delete(ctx context.Context, id, name string) (bool, error) {
	_, err := cdr.redisRepo.Delete(ctx, id, name)
	if err != nil {
		return false, err
	}

	ok, err := cdr.dbRepo.Delete(ctx, id, name)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (cdr *CachedDbRepository) SelectAll(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error) {
	subs, err := cdr.dbRepo.SelectAll(ctx, id, period)
	if err != nil {
		log.Println("selectAll: ", err)
		return nil, err
	}

	return subs, nil
}

func (cdr *CachedDbRepository) Close() {
	cdr.dbRepo.Close()
	cdr.redisRepo.Close()
}
