package v1

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"subscription-service/internal/repository/postgresRepository"
	pb "subscription-service/proto/pkg/proto"
	"testing"
)

type createSubscriptionTestCase struct {
	name      string
	req       *pb.CreateSubscriptionRequest
	want      *pb.CreateSubscriptionResponse
	stubError error
	wantCode  codes.Code
}

type getSubscriptionTestCase struct {
	name      string
	req       *pb.GetSubscriptionRequest
	want      *pb.GetSubscriptionResponse
	stubError error
	wantCode  codes.Code
}

type updateSubscriptionTestCase struct {
	name      string
	req       *pb.UpdateSubscriptionRequest
	want      *pb.UpdateSubscriptionResponse
	stubError error
	wantCode  codes.Code
}

type deleteSubscriptionTestCase struct {
	name      string
	req       *pb.DeleteSubscriptionRequest
	want      *pb.DeleteSubscriptionResponse
	stubError error
	wantCode  codes.Code
}

type getAnalyticsTestCase struct {
	name      string
	req       *pb.GetAnalyticsRequest
	want      *pb.GetAnalyticsResponse
	stubError error
	wantCode  codes.Code
}

type StubSubscriptionService struct {
	CreateFunc                 func(ctx context.Context, id, name, startedAt, expiration string, price int) (*pb.Subscription, error)
	GetFunc                    func(ctx context.Context, id, name string) (*pb.Subscription, error)
	UpdateFunc                 func(ctx context.Context, id, oldName, name, startedAt, expiration string, price int) (*pb.Subscription, error)
	DeleteFunc                 func(ctx context.Context, id, name string) (bool, error)
	GetAnalyticsFunc           func(ctx context.Context, id string, period *pb.Period) (*pb.Analytics, error)
	GetSubscriptionsByDateFunc func(ctx context.Context, date string) ([]*pb.Subscription, error)
}

func (scs *StubSubscriptionService) GetSubscriptionsByDate(ctx context.Context, date string) ([]*pb.Subscription, error) {
	return scs.GetSubscriptionsByDateFunc(ctx, date)
}

func (scs *StubSubscriptionService) CreateSubscription(ctx context.Context, id, name, startedAt, expiration string, price int) (*pb.Subscription, error) {
	return scs.CreateFunc(ctx, id, name, startedAt, expiration, price)
}

func (scs *StubSubscriptionService) GetSubscription(ctx context.Context, id, name string) (*pb.Subscription, error) {
	return scs.GetFunc(ctx, id, name)
}

func (scs *StubSubscriptionService) UpdateSubscription(ctx context.Context, id, oldName, name, startedAt, expiration string, price int) (*pb.Subscription, error) {
	return scs.UpdateFunc(ctx, id, oldName, name, startedAt, expiration, price)
}

func (scs *StubSubscriptionService) DeleteSubscription(ctx context.Context, id, name string) (bool, error) {
	return scs.DeleteFunc(ctx, id, name)
}

func (scs *StubSubscriptionService) GetAnalytics(ctx context.Context, id string, period *pb.Period) (*pb.Analytics, error) {
	return scs.GetAnalyticsFunc(ctx, id, period)
}

func TestSubscriptionServer_CreateSubscription(t *testing.T) {
	testCases := []createSubscriptionTestCase{
		{
			name: "success",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			want: &pb.CreateSubscriptionResponse{
				Sub: &pb.Subscription{
					Id:         "user-1",
					Name:       "twitch",
					StartedAt:  "2024-08-30",
					Expiration: "2025-12-30",
					Price:      int32(400),
				},
			},
			stubError: nil,
		},
		{
			name: "empty id",
			req: &pb.CreateSubscriptionRequest{
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			stubError: ErrEmptyID,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty name",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			stubError: ErrEmptyName,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty start date",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			stubError: ErrEmptyStartDate,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty expiration date",
			req: &pb.CreateSubscriptionRequest{
				Id:        "user-1",
				Name:      "twitch",
				StartedAt: "2024-08-30",
				Price:     int32(400),
			},
			stubError: ErrEmptyExpiration,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "not positive price",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
			},
			stubError: ErrNotPositivePrice,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "db unavailable",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrDBUnavailable,
			wantCode:  codes.Unavailable,
		},
		{
			name: "row was not created",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrNotFound,
			wantCode:  codes.NotFound,
		},
		{
			name: "row already exists",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrAlreadyExists,
			wantCode:  codes.AlreadyExists,
		},
		{
			name: "invalid date from db",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrInvalidDate,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "start date after end date",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2025-08-30",
				Expiration: "2024-12-30",
				Price:      int32(1000),
			},
			stubError: ErrPeriodNotValid,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "unknown error",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrUnknown,
			wantCode:  codes.Internal,
		},
		{
			name: "deadline exceeded",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: context.DeadlineExceeded,
			wantCode:  codes.DeadlineExceeded,
		},
		{
			name: "canceled",
			req: &pb.CreateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: context.Canceled,
			wantCode:  codes.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &StubSubscriptionService{
				CreateFunc: func(
					ctx context.Context,
					id, name, startedAt, expiration string,
					price int,
				) (*pb.Subscription, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}
					return &pb.Subscription{
						Id:         id,
						Name:       name,
						StartedAt:  startedAt,
						Expiration: expiration,
						Price:      int32(price),
					}, nil
				},
			}

			server := &SubscriptionServer{
				subService: stub,
			}

			resp, err := server.CreateSubscription(context.Background(), tc.req)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.want, resp)
				return
			}

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.wantCode, st.Code())
		})
	}
}

func TestSubscriptionServer_GetSubscription(t *testing.T) {
	testCases := []getSubscriptionTestCase{
		{
			name: "success",
			req: &pb.GetSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			want: &pb.GetSubscriptionResponse{
				Sub: &pb.Subscription{
					Id:         "user-1",
					Name:       "twitch",
					StartedAt:  "2024-08-30",
					Expiration: "2025-12-30",
					Price:      int32(400),
				},
			},
			stubError: nil,
		},
		{
			name: "empty id",
			req: &pb.GetSubscriptionRequest{
				Name: "twitch",
			},
			stubError: ErrEmptyID,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty name",
			req: &pb.GetSubscriptionRequest{
				Id: "user-1",
			},
			stubError: ErrEmptyName,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "db unavailable",
			req: &pb.GetSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: postgresRepository.ErrDBUnavailable,
			wantCode:  codes.Unavailable,
		},
		{
			name: "unknown error",
			req: &pb.GetSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: postgresRepository.ErrUnknown,
			wantCode:  codes.Internal,
		},
		{
			name: "deadline exceeded",
			req: &pb.GetSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: context.DeadlineExceeded,
			wantCode:  codes.DeadlineExceeded,
		},
		{
			name: "canceled",
			req: &pb.GetSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: context.Canceled,
			wantCode:  codes.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &StubSubscriptionService{
				GetFunc: func(
					ctx context.Context,
					id, name string,
				) (*pb.Subscription, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}
					return &pb.Subscription{
						Id:         id,
						Name:       name,
						StartedAt:  "2024-08-30",
						Expiration: "2025-12-30",
						Price:      int32(400),
					}, nil
				},
			}

			server := &SubscriptionServer{
				subService: stub,
			}

			resp, err := server.GetSubscription(context.Background(), tc.req)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.want, resp)
				return
			}

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.wantCode, st.Code())
		})
	}
}

func TestSubscriptionServer_UpdateSubscription(t *testing.T) {
	testCases := []updateSubscriptionTestCase{
		{
			name: "success",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "youtube",
				OldName:    "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			want: &pb.UpdateSubscriptionResponse{
				Sub: &pb.Subscription{
					Id:         "user-1",
					Name:       "youtube",
					StartedAt:  "2024-08-30",
					Expiration: "2025-12-30",
					Price:      int32(400),
				},
			},
			stubError: nil,
		},
		{
			name: "empty id",
			req: &pb.UpdateSubscriptionRequest{
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			stubError: ErrEmptyID,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty name",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			stubError: ErrEmptyName,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty start date",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				Expiration: "2025-12-30",
				Price:      int32(400),
			},
			stubError: ErrEmptyStartDate,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty expiration date",
			req: &pb.UpdateSubscriptionRequest{
				Id:        "user-1",
				Name:      "twitch",
				StartedAt: "2024-08-30",
				Price:     int32(400),
			},
			stubError: ErrEmptyExpiration,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "not positive price",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
			},
			stubError: ErrNotPositivePrice,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "db unavailable",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrDBUnavailable,
			wantCode:  codes.Unavailable,
		},
		{
			name: "row was not created",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrNotFound,
			wantCode:  codes.NotFound,
		},
		{
			name: "row already exists",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrAlreadyExists,
			wantCode:  codes.AlreadyExists,
		},
		{
			name: "invalid argument",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "not date",
				Expiration: "2024-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrInvalidArgument,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "unknown error",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: postgresRepository.ErrUnknown,
			wantCode:  codes.Internal,
		},
		{
			name: "deadline exceeded",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: context.DeadlineExceeded,
			wantCode:  codes.DeadlineExceeded,
		},
		{
			name: "canceled",
			req: &pb.UpdateSubscriptionRequest{
				Id:         "user-1",
				Name:       "twitch",
				StartedAt:  "2024-08-30",
				Expiration: "2025-12-30",
				Price:      int32(1000),
			},
			stubError: context.Canceled,
			wantCode:  codes.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &StubSubscriptionService{
				UpdateFunc: func(
					ctx context.Context,
					id, oldName, name, startedAt, expiration string,
					price int,
				) (*pb.Subscription, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}
					return &pb.Subscription{
						Id:         id,
						Name:       name,
						StartedAt:  startedAt,
						Expiration: expiration,
						Price:      int32(price),
					}, nil
				},
			}

			server := &SubscriptionServer{
				subService: stub,
			}

			resp, err := server.UpdateSubscription(context.Background(), tc.req)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.want, resp)
				return
			}

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.wantCode, st.Code())
		})
	}
}

func TestSubscriptionServer_DeleteSubscription(t *testing.T) {
	testCases := []deleteSubscriptionTestCase{
		{
			name: "success",
			req: &pb.DeleteSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			want: &pb.DeleteSubscriptionResponse{
				Success: true,
			},
			stubError: nil,
		},
		{
			name: "empty id",
			req: &pb.DeleteSubscriptionRequest{
				Name: "twitch",
			},
			stubError: ErrEmptyID,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty name",
			req: &pb.DeleteSubscriptionRequest{
				Id: "user-1",
			},
			stubError: ErrEmptyName,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "db unavailable",
			req: &pb.DeleteSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: postgresRepository.ErrDBUnavailable,
			wantCode:  codes.Unavailable,
		},
		{
			name: "invalid argument",
			req: &pb.DeleteSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: postgresRepository.ErrInvalidArgument,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "unknown error",
			req: &pb.DeleteSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: postgresRepository.ErrUnknown,
			wantCode:  codes.Internal,
		},
		{
			name: "deadline exceeded",
			req: &pb.DeleteSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: context.DeadlineExceeded,
			wantCode:  codes.DeadlineExceeded,
		},
		{
			name: "canceled",
			req: &pb.DeleteSubscriptionRequest{
				Id:   "user-1",
				Name: "twitch",
			},
			stubError: context.Canceled,
			wantCode:  codes.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &StubSubscriptionService{
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

			server := &SubscriptionServer{
				subService: stub,
			}

			resp, err := server.DeleteSubscription(context.Background(), tc.req)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.want, resp)
				return
			}

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.wantCode, st.Code())
		})
	}
}

func TestSubscriptionServer_GetAnalyticsSubscription(t *testing.T) {
	testCases := []getAnalyticsTestCase{
		{
			name: "success",
			req: &pb.GetAnalyticsRequest{
				Id: "user-1",
				Period: &pb.Period{
					Start: "2024-01-01",
					End:   "2025-01-01",
				},
			},
			want: &pb.GetAnalyticsResponse{
				Analytics: &pb.Analytics{},
			},
			stubError: nil,
		},
		{
			name:      "empty id",
			req:       &pb.GetAnalyticsRequest{},
			stubError: ErrEmptyID,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "db unavailable",
			req: &pb.GetAnalyticsRequest{
				Id: "user-1",
			},
			stubError: postgresRepository.ErrDBUnavailable,
			wantCode:  codes.Unavailable,
		},
		{
			name: "no rows found",
			req: &pb.GetAnalyticsRequest{
				Id: "user-1",
			},
			stubError: postgresRepository.ErrNotFound,
			wantCode:  codes.NotFound,
		},
		{
			name: "invalid argument",
			req: &pb.GetAnalyticsRequest{
				Id: "not exist id",
			},
			stubError: postgresRepository.ErrInvalidArgument,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "unknown error",
			req: &pb.GetAnalyticsRequest{
				Id: "user-1",
			},
			stubError: postgresRepository.ErrUnknown,
			wantCode:  codes.Internal,
		},
		{
			name: "deadline exceeded",
			req: &pb.GetAnalyticsRequest{
				Id: "user-1",
			},
			stubError: context.DeadlineExceeded,
			wantCode:  codes.DeadlineExceeded,
		},
		{
			name: "canceled",
			req: &pb.GetAnalyticsRequest{
				Id: "user-1",
			},
			stubError: context.Canceled,
			wantCode:  codes.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &StubSubscriptionService{
				GetAnalyticsFunc: func(
					ctx context.Context,
					id string,
					period *pb.Period,
				) (*pb.Analytics, error) {
					if tc.stubError != nil {
						return nil, tc.stubError
					}
					return &pb.Analytics{}, nil
				},
			}

			server := &SubscriptionServer{
				subService: stub,
			}

			resp, err := server.GetAnalytics(context.Background(), tc.req)

			if tc.stubError == nil {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.want, resp)
				return
			}

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.wantCode, st.Code())
		})
	}
}
