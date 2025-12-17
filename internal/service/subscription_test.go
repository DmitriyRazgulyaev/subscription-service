package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	pb "subscription-service/proto/pkg/proto"
	"testing"
	"time"
)

type createSubscriptionTestCase struct {
	caseName                        string
	id, name, startedAt, expiration string
	price                           int
	want                            *pb.Subscription
	stubError                       error
}

type selectByNameSubscriptionTestCase struct {
	caseName  string
	id, name  string
	want      *pb.Subscription
	stubError error
}

type updateSubscriptionTestCase struct {
	caseName                                 string
	id, oldName, name, startedAt, expiration string
	price                                    int
	want                                     *pb.Subscription
	stubError                                error
}

type deleteSubscriptionTestCase struct {
	caseName  string
	id, name  string
	want      bool
	stubError error
}

type StubSubscriptionRepository struct {
	CreateFunc       func(ctx context.Context, subscription *pb.Subscription) (*pb.Subscription, error)
	SelectByNameFunc func(ctx context.Context, id, name string) (*pb.Subscription, error)
	UpdateFunc       func(ctx context.Context, subscription *pb.Subscription, oldName string) (*pb.Subscription, error)
	DeleteFunc       func(ctx context.Context, id, name string) (bool, error)
	SelectAllFunc    func(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error)
}

func (ssr *StubSubscriptionRepository) Create(ctx context.Context, subscription *pb.Subscription) (*pb.Subscription, error) {
	return ssr.CreateFunc(ctx, subscription)
}

func (ssr *StubSubscriptionRepository) SelectByName(ctx context.Context, id, name string) (*pb.Subscription, error) {
	return ssr.SelectByNameFunc(ctx, id, name)
}

func (ssr *StubSubscriptionRepository) Update(ctx context.Context, subscription *pb.Subscription, oldName string) (*pb.Subscription, error) {
	return ssr.UpdateFunc(ctx, subscription, oldName)
}

func (ssr *StubSubscriptionRepository) Delete(ctx context.Context, id, name string) (bool, error) {
	return ssr.DeleteFunc(ctx, id, name)
}

func (ssr *StubSubscriptionRepository) SelectAll(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, error) {
	return ssr.SelectAllFunc(ctx, id, period)
}

func (ssr *StubSubscriptionRepository) Close() {
}

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	testCases := []createSubscriptionTestCase{
		{
			caseName:   "success",
			id:         "user-1",
			name:       "twitch",
			startedAt:  "2024-01-01",
			expiration: "2025-01-01",
			price:      500,
			want: &pb.Subscription{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-01-01",
				Expiration: "2025-01-01",
				Price:      int32(500),
			},
			stubError: nil,
		},
		{
			caseName:   "error in repository",
			id:         "user-1",
			name:       "twitch",
			startedAt:  "2024-01-01",
			expiration: "2025-01-01",
			price:      500,
			want: &pb.Subscription{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-01-01",
				Expiration: "2025-01-01",
				Price:      int32(500),
			},
			stubError: errors.New("something goes wrong"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			stub := &StubSubscriptionRepository{
				CreateFunc: func(
					ctx context.Context,
					subscription *pb.Subscription,
				) (*pb.Subscription, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}
					return subscription, nil
				},
			}

			service := SubscriptionService{subRepo: stub}

			result, err := service.CreateSubscription(context.Background(), tc.id, tc.name,
				tc.startedAt, tc.expiration, tc.price)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.want, result)
				return
			}

			require.Error(t, err)
		})
	}
}

func TestSubscriptionService_GetSubscription(t *testing.T) {
	testCases := []selectByNameSubscriptionTestCase{
		{
			caseName: "success",
			id:       "user-1",
			name:     "twitch",
			want: &pb.Subscription{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-01-01",
				Expiration: "2025-01-01",
				Price:      int32(500),
			},
			stubError: nil,
		},
		{
			caseName: "error in repository",
			id:       "user-1",
			name:     "twitch",
			want: &pb.Subscription{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-01-01",
				Expiration: "2025-01-01",
				Price:      int32(500),
			},
			stubError: errors.New("something goes wrong"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			stub := &StubSubscriptionRepository{
				SelectByNameFunc: func(
					ctx context.Context,
					id, name string,
				) (*pb.Subscription, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}

					return &pb.Subscription{
						Id:         id,
						Name:       name,
						StartedAt:  "2024-01-01",
						Expiration: "2025-01-01",
						Price:      int32(500),
					}, nil
				},
			}

			service := SubscriptionService{subRepo: stub}

			result, err := service.GetSubscription(context.Background(), tc.id, tc.name)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.want, result)
				return
			}

			require.Error(t, err)
		})
	}
}

func TestSubscriptionService_UpdateSubscription(t *testing.T) {
	testCases := []updateSubscriptionTestCase{
		{
			caseName:   "success",
			id:         "user-1",
			name:       "twitch",
			startedAt:  "2024-01-01",
			expiration: "2025-01-01",
			price:      500,
			want: &pb.Subscription{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-01-01",
				Expiration: "2025-01-01",
				Price:      int32(500),
			},
			stubError: nil,
		},
		{
			caseName: "error in repository",
			id:       "user-1",
			name:     "twitch",
			want: &pb.Subscription{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-01-01",
				Expiration: "2025-01-01",
				Price:      int32(500),
			},
			stubError: errors.New("something goes wrong"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			stub := &StubSubscriptionRepository{
				UpdateFunc: func(
					ctx context.Context,
					subscription *pb.Subscription,
					oldName string,
				) (*pb.Subscription, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}

					return &pb.Subscription{
						Id:         subscription.Id,
						Name:       subscription.Name,
						StartedAt:  "2024-01-01",
						Expiration: "2025-01-01",
						Price:      int32(500),
					}, nil
				},
			}

			service := SubscriptionService{subRepo: stub}

			result, err := service.UpdateSubscription(context.Background(), tc.id, tc.oldName, tc.name, tc.startedAt, tc.expiration, tc.price)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.want, result)
				return
			}

			require.Error(t, err)
		})
	}
}

func TestSubscriptionService_DeleteSubscription(t *testing.T) {
	testCases := []deleteSubscriptionTestCase{
		{
			caseName:  "success",
			id:        "user-1",
			name:      "twitch",
			want:      true,
			stubError: nil,
		},
		{
			caseName:  "error in repository",
			id:        "user-1",
			name:      "twitch",
			want:      false,
			stubError: errors.New("something goes wrong"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			stub := &StubSubscriptionRepository{
				DeleteFunc: func(
					ctx context.Context,
					id, name string,
				) (bool, error) {
					if tc.stubError != nil {
						return false, tc.stubError
					}

					return true, nil
				},
			}

			service := SubscriptionService{subRepo: stub}

			result, err := service.DeleteSubscription(context.Background(), tc.id, tc.name)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.True(t, result)
				return
			}

			require.Error(t, err)
		})
	}
}

func TestResolvePeriod_NilPeriod(t *testing.T) {
	fixedNow := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)

	wantStartPeriod := time.Date(fixedNow.Year(), fixedNow.Month(), 1, 0, 0, 0, 0, fixedNow.Location())
	wantEndPeriod := wantStartPeriod.AddDate(0, 1, 0).Add(-time.Nanosecond)
	parsedWantStartPeriod := wantStartPeriod.Format("2006-01-02")
	parsedWantEndPeriod := wantEndPeriod.Format("2006-01-02")

	start, end, period, err := ResolvePeriod(context.Background(), nil, fixedNow)
	fmt.Println(parsedWantStartPeriod, parsedWantEndPeriod)
	require.NoError(t, err)
	require.Equal(t, parsedWantStartPeriod, period.Start, "start period")
	require.Equal(t, parsedWantEndPeriod, period.End, "end period")
	require.True(t, end.After(start))
}

func TestResolvePeriod_ValidPeriod(t *testing.T) {

	period := &pb.Period{Start: "2024-01-01", End: "2025-01-01"}
	parsedWantStartPeriod, err := time.Parse("2006-01-02", period.Start)
	if err != nil {
		t.Errorf("error with parsing start date: %v", err)
	}

	parsedWantEndPeriod, err := time.Parse("2006-01-02", period.End)
	if err != nil {
		t.Errorf("error with parsing end date: %v", err)
	}

	start, end, period, err := ResolvePeriod(context.Background(), period, time.Now())

	require.NoError(t, err)
	require.Equal(t, parsedWantStartPeriod.Format("2006-01-02"), period.Start, "start period")
	require.Equal(t, parsedWantEndPeriod.Format("2006-01-02"), period.End, "end period")
	require.True(t, end.After(start))
}
