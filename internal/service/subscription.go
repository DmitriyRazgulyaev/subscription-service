package service

import (
	"context"
	"log"
	"strconv"
	pb "subscription-service/proto/pkg/proto"
	"time"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *pb.Subscription) (*pb.Subscription, error)
	SelectByName(ctx context.Context, id, name string) (*pb.Subscription, error)
	Update(ctx context.Context, subscription *pb.Subscription, oldName string) (*pb.Subscription, error)
	Delete(ctx context.Context, id, name string) (bool, error)
	SelectAll(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error)
	Close()
}

type SubscriptionService struct {
	subRepo SubscriptionRepository
}

func NewSubscriptionService(repository SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		subRepo: repository,
	}
}

func (ss *SubscriptionService) CreateSubscription(ctx context.Context, id, name, startedAt, expiration string, price int) (*pb.Subscription, error) {
	sub := pb.Subscription{
		Id:         id,
		Name:       name,
		StartedAt:  startedAt,
		Expiration: expiration,
		Price:      int32(price),
	}

	newSub, err := ss.subRepo.Create(ctx, &sub)
	if err != nil {
		return nil, err
	}

	return newSub, nil
}

func (ss *SubscriptionService) GetSubscription(ctx context.Context, id, name string) (*pb.Subscription, error) {
	sub, err := ss.subRepo.SelectByName(ctx, id, name)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (ss *SubscriptionService) UpdateSubscription(ctx context.Context, id, oldName, name, startedAt, expiration string, price int) (*pb.Subscription, error) {
	sub := pb.Subscription{
		Id:         id,
		Name:       name,
		StartedAt:  startedAt,
		Expiration: expiration,
		Price:      int32(price),
	}
	updatedSub, err := ss.subRepo.Update(ctx, &sub, oldName)
	if err != nil {
		return nil, err
	}

	return updatedSub, nil
}

func (ss *SubscriptionService) DeleteSubscription(ctx context.Context, id, name string) (bool, error) {
	ok, err := ss.subRepo.Delete(ctx, id, name)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// GetAnalytics выдает все подписки клиента по его id в определенный период: если period == nil,
// то аналитика за нынешний месяц, если period задан, то аналитика по заданному period
func (ss *SubscriptionService) GetAnalytics(ctx context.Context, id string, period *pb.Period) (*pb.Analytics, error) {
	subs, periodStart, periodEnd, err := ss.getDataForAnalytics(ctx, id, period)
	if err != nil {
		return nil, err
	}

	analytics := initAnalytics(periodStart, periodEnd)
	for date := periodStart; !date.After(periodEnd); date = date.AddDate(0, 1, 0) {

		for _, sub := range subs {
			subStart, err := time.Parse("2006-01-02", sub.StartedAt)
			if err != nil {
				return nil, err
			}
			subEnd, err := time.Parse("2006-01-02", sub.Expiration)
			if err != nil {
				return nil, err
			}
			log.Println(date, sub)
			if !date.Before(subStart) && !date.After(subEnd) {
				log.Println(sub)
				yearKey := strconv.Itoa(date.Year())
				analytics.TotalSpent += sub.Price
				analytics.Years[yearKey].TotalSpent += sub.Price
				analytics.Years[yearKey].Months[date.Month()-1].TotalMonthSpent += sub.Price
				analytics.Years[yearKey].Months[date.Month()-1].Subs = append(analytics.Years[yearKey].Months[date.Month()-1].Subs, sub)
			}
		}
	}

	return analytics, nil
}

func ResolvePeriod(ctx context.Context, period *pb.Period, now time.Time) (time.Time, time.Time, *pb.Period, error) {
	var periodStart, periodEnd time.Time
	var err error
	if period == nil {
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		period = &pb.Period{
			Start: periodStart.Format("2006-01-02"),
			End:   periodEnd.Format("2006-01-02"),
		}
		return periodStart, periodEnd, period, nil
	}

	periodStart, err = time.Parse("2006-01-02", period.Start)
	if err != nil {
		return time.Time{}, time.Time{}, &pb.Period{}, err
	}

	periodEnd, err = time.Parse("2006-01-02", period.End)
	if err != nil {
		return time.Time{}, time.Time{}, &pb.Period{}, err
	}

	return periodStart, periodEnd, period, nil
}

func (ss *SubscriptionService) getDataForAnalytics(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, time.Time, time.Time, error) {
	periodStart, periodEnd, period, err := ResolvePeriod(ctx, period, time.Now())
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	subs, err := ss.subRepo.SelectAll(ctx, id, period)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	return subs, periodStart, periodEnd, nil
}

func initAnalytics(start, end time.Time) *pb.Analytics {
	analytics := &pb.Analytics{
		Years:           make(map[string]*pb.Year),
		NextMonthSpends: &pb.MonthSpent{Subs: make([]*pb.Subscription, 0)},
	}

	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	end = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	for current := start; !current.After(end); current = current.AddDate(0, 1, 0) {
		yearKey := strconv.Itoa(current.Year())

		if _, exists := analytics.Years[yearKey]; !exists {
			year := &pb.Year{
				Year:   yearKey,
				Months: make([]*pb.MonthSpent, 12),
			}

			for i := 0; i < 12; i++ {
				year.Months[i] = &pb.MonthSpent{
					Month: time.Month(i + 1).String(),
					Subs:  make([]*pb.Subscription, 0),
				}
			}

			analytics.Years[yearKey] = year
		}
	}

	return analytics
}
