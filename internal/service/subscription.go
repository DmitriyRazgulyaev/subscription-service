package service

import (
	"context"
	"fmt"
	pb "subscription-service/proto/pkg/proto"
	"time"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *pb.Subscription) (*pb.Subscription, error)
	SelectByName(ctx context.Context, id, name string) (*pb.Subscription, error)
	Update(ctx context.Context, subscription *pb.Subscription) (*pb.Subscription, error)
	Delete(ctx context.Context, id, name string) (bool, error)
	SelectAll(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error)
}

type SubscriptionService struct {
	subRepo SubscriptionRepository
}

func NewSubscriptionService(repository SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		subRepo: repository,
	}
}

func (ss *SubscriptionService) CreateSubscription(ctx context.Context, id, name, expiration string, price int) (*pb.Subscription, error) {
	sub := pb.Subscription{
		Id:         id,
		Name:       name,
		Expiration: expiration,
		Price:      int32(price),
	}

	newSub, err := ss.subRepo.Create(ctx, &sub)
	if err != nil {
		return nil, fmt.Errorf("createSubscription.Create: %w", err)
	}

	return newSub, nil
}

func (ss *SubscriptionService) GetSubscription(ctx context.Context, id, name string) (*pb.Subscription, error) {
	sub, err := ss.subRepo.SelectByName(ctx, id, name)
	if err != nil {
		return nil, fmt.Errorf("getSubscription: %w", err)
	}

	return sub, nil
}

func (ss *SubscriptionService) UpdateSubscription(ctx context.Context, id, name, expiration string, price int) (*pb.Subscription, error) {
	sub := pb.Subscription{
		Id:         id,
		Name:       name,
		Expiration: expiration,
		Price:      int32(price),
	}
	updatedSub, err := ss.subRepo.Update(ctx, &sub)
	if err != nil {
		return nil, fmt.Errorf("updateSubscription: %w", err)
	}

	return updatedSub, nil
}

func (ss *SubscriptionService) DeleteSubscription(ctx context.Context, id, name string) (bool, error) {
	ok, err := ss.subRepo.Delete(ctx, id, name)
	if err != nil {
		return false, fmt.Errorf("deleteSubscription: %w", err)
	}

	return ok, nil
}

// GetAnalytics выдает все подписки клиента по его id в определенный период:
func (ss *SubscriptionService) GetAnalytics(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, int, error) {
	var subs []*pb.Subscription
	var err error
	if period == nil {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		nextMonth := start.AddDate(0, 1, 0)
		end := nextMonth.Add(-time.Nanosecond)

		period = &pb.Period{
			Start: start.String(),
			End:   end.String(),
		}
		subs, err = ss.subRepo.SelectAll(ctx, id, period)
	} else {
		subs, err = ss.subRepo.SelectAll(ctx, id, period)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("getAnalytics: %w", err)
	}

	var totalPrice int
	for _, sub := range subs {
		totalPrice += int(sub.GetPrice())
	}

	return []*pb.Subscription{}, 0, nil
}
