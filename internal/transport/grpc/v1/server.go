package v1

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"subscription-service/internal/repository/postgresRepository"
	pb "subscription-service/proto/pkg/proto"
	"time"
)

var (
	ErrEmptyID          = errors.New("empty id is given")
	ErrEmptyName        = errors.New("empty name is given")
	ErrEmptyStartDate   = errors.New("empty startedAt is given")
	ErrEmptyExpiration  = errors.New("empty expiration date is given")
	ErrNotPositivePrice = errors.New("price must be positive")
	ErrPeriodNotValid   = errors.New("start date must be before end date")
)

type SubscriptionServer struct {
	pb.UnimplementedSubscriptionServiceServer
	subService SubscriptionService
}

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, id, name, startedAt, expiration string, price int) (*pb.Subscription, error)
	GetSubscription(ctx context.Context, id, name string) (*pb.Subscription, error)
	UpdateSubscription(ctx context.Context, id, oldName, name, startedAt, expiration string, price int) (*pb.Subscription, error)
	DeleteSubscription(ctx context.Context, id, name string) (bool, error)
	GetAnalytics(ctx context.Context, id string, period *pb.Period) (*pb.Analytics, error)
	GetSubscriptionsByDate(ctx context.Context, date string) ([]*pb.Subscription, error)
}

func Register(grpcServer *grpc.Server, service SubscriptionService) {
	pb.RegisterSubscriptionServiceServer(grpcServer, &SubscriptionServer{subService: service})
}

func (ss *SubscriptionServer) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (*pb.CreateSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyID.Error())
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyName.Error())
	}

	if req.GetStartedAt() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyStartDate.Error())
	}

	if req.GetExpiration() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyExpiration.Error())
	}

	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, ErrNotPositivePrice.Error())
	}
	startDate, err := time.Parse("2006-01-02", req.StartedAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	endDate, err := time.Parse("2006-01-02", req.Expiration)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !startDate.Before(endDate) {
		return nil, status.Error(codes.InvalidArgument, ErrPeriodNotValid.Error())
	}

	sub, err := ss.subService.CreateSubscription(ctx, req.GetId(), req.GetName(), req.GetStartedAt(), req.GetExpiration(), int(req.GetPrice()))
	if err != nil {
		switch {
		case errors.Is(err, postgresRepository.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, postgresRepository.ErrInvalidDate):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, postgresRepository.ErrDBUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		case errors.Is(err, postgresRepository.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.CreateSubscriptionResponse{Sub: sub}, nil
}

func (ss *SubscriptionServer) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.GetSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyID.Error())
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyName.Error())
	}

	sub, err := ss.subService.GetSubscription(ctx, req.GetId(), req.GetName())
	if err != nil {
		switch {
		case errors.Is(err, postgresRepository.ErrDBUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		case errors.Is(err, postgresRepository.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.GetSubscriptionResponse{Sub: sub}, nil
}

func (ss *SubscriptionServer) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (*pb.UpdateSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyID.Error())
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyName.Error())
	}

	if req.GetStartedAt() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyStartDate.Error())
	}

	if req.GetExpiration() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyExpiration.Error())
	}

	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, ErrNotPositivePrice.Error())
	}

	startDate, err := time.Parse("2006-01-02", req.StartedAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	endDate, err := time.Parse("2006-01-02", req.Expiration)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !startDate.Before(endDate) {
		return nil, status.Error(codes.InvalidArgument, ErrPeriodNotValid.Error())
	}

	sub, err := ss.subService.UpdateSubscription(ctx, req.GetId(), req.GetOldName(), req.GetName(), req.GetStartedAt(), req.GetExpiration(), int(req.GetPrice()))
	if err != nil {
		switch {
		case errors.Is(err, postgresRepository.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, postgresRepository.ErrDBUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		case errors.Is(err, postgresRepository.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, postgresRepository.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.UpdateSubscriptionResponse{Sub: sub}, nil
}

func (ss *SubscriptionServer) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (*pb.DeleteSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyID.Error())
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyName.Error())
	}

	ok, err := ss.subService.DeleteSubscription(ctx, req.GetId(), req.GetName())
	if err != nil {
		switch {
		case errors.Is(err, postgresRepository.ErrDBUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		case errors.Is(err, postgresRepository.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, postgresRepository.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.DeleteSubscriptionResponse{Success: ok}, nil
}

func (ss *SubscriptionServer) GetAnalytics(ctx context.Context, req *pb.GetAnalyticsRequest) (*pb.GetAnalyticsResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyID.Error())
	}

	analytics, err := ss.subService.GetAnalytics(ctx, req.GetId(), req.GetPeriod())
	if err != nil {
		switch {
		case errors.Is(err, postgresRepository.ErrDBUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		case errors.Is(err, postgresRepository.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, postgresRepository.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.GetAnalyticsResponse{Analytics: analytics}, nil
}

func (ss *SubscriptionServer) GetExpiringSubscriptions(ctx context.Context, req *pb.GetExpiringSubscriptionsRequest) (*pb.GetExpiringSubscriptionsResponse, error) {
	if req.GetDate() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyExpiration.Error())
	}

	subs, err := ss.subService.GetSubscriptionsByDate(ctx, req.GetDate())
	if err != nil {
		switch {
		case errors.Is(err, postgresRepository.ErrDBUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		case errors.Is(err, postgresRepository.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, postgresRepository.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.GetExpiringSubscriptionsResponse{Subs: subs}, nil
}
