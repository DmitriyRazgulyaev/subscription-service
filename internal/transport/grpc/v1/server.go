package v1

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "subscription-service/proto/pkg/proto"
)

type SubscriptionServer struct {
	pb.UnimplementedSubscriptionServiceServer
	subService SubscriptionService
}

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, id, name, startedAt, expiration string, price int) (*pb.Subscription, error)
	GetSubscription(ctx context.Context, id, name string) (*pb.Subscription, error)
	UpdateSubscription(ctx context.Context, id, name, startedAt, expiration string, price int) (*pb.Subscription, error)
	DeleteSubscription(ctx context.Context, id, name string) (bool, error)
	GetAnalytics(ctx context.Context, id string, period *pb.Period) ([]*pb.Subscription, int, error)
}

func Register(grpcServer *grpc.Server, service SubscriptionService) {
	pb.RegisterSubscriptionServiceServer(grpcServer, &SubscriptionServer{subService: service})
}

func (ss *SubscriptionServer) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (*pb.CreateSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id is given")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty name is given")
	}

	if req.GetStartedAt() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty startedAt is given")
	}

	if req.GetExpiration() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty expiration date is given")
	}

	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}

	sub, err := ss.subService.CreateSubscription(ctx, req.GetId(), req.GetName(), req.GetStartedAt(), req.GetExpiration(), int(req.GetPrice()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "createSubscription: %v", err)
		//TODO сделать различные коды ошибок: уже существует, не удалось добавить
	}

	return &pb.CreateSubscriptionResponse{Sub: sub}, nil
}

func (ss *SubscriptionServer) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.GetSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id is given")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty name is given")
	}

	sub, err := ss.subService.GetSubscription(ctx, req.GetId(), req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getSubscription: %v", err)
		//TODO сделать различные коды ошибок: не существует такой подписки, внутренняя ошибка
	}

	return &pb.GetSubscriptionResponse{Sub: sub}, nil
}

func (ss *SubscriptionServer) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (*pb.UpdateSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id is given")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty name is given")
	}

	if req.GetStartedAt() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty startedAt is given")
	}

	if req.GetExpiration() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty expiration date is given")
	}

	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}

	sub, err := ss.subService.UpdateSubscription(ctx, req.GetId(), req.GetName(), req.GetStartedAt(), req.GetExpiration(), int(req.GetPrice()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "updateSubscription: %v", err)
		//TODO сделать разные коды ошибок: нет такой подписки, внутренняя ошибка
	}

	return &pb.UpdateSubscriptionResponse{Sub: sub}, nil
}

func (ss *SubscriptionServer) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (*pb.DeleteSubscriptionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id is given")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty name is given")
	}

	ok, err := ss.subService.DeleteSubscription(ctx, req.GetId(), req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deleteSubscription: %v", err)
		//TODO сделать разные коды ошибок: нет такой подписки, внутренняя ошибка
	}

	return &pb.DeleteSubscriptionResponse{Success: ok}, nil
}

func (ss *SubscriptionServer) GetAnalytics(ctx context.Context, req *pb.GetAnalyticsRequest) (*pb.GetAnalyticsResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id is given")
	}

	subs, totalPrice, err := ss.subService.GetAnalytics(ctx, req.GetId(), req.GetPeriod())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getAnalytics: %v", err)
		//TODO сделать разные коды ошибок: нет подписок у пользователя с таким id, внутренняя ошибка
	}

	return &pb.GetAnalyticsResponse{Subs: subs, SummarySpend: int32(totalPrice)}, nil
}
