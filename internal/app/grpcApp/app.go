package grpcApp

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
	"subscription-service/internal/resilience"
	"subscription-service/internal/service"
	v1 "subscription-service/internal/transport/grpc/v1"
	"subscription-service/pkg/logger"
	"subscription-service/proto/pkg/proto"
	"time"
)

const (
	timeoutValue = time.Second * 10
	maxRetries   = 3
	baseDelay    = time.Second
	maxDelay     = time.Second * 10
)

type App struct {
	grpcServer *grpc.Server
	port       int
	srv        http.Server
	logger     *logger.Logger
	repo       service.SubscriptionRepository
}

func LoggingUnaryServerInterceptor(logger *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.Info(fmt.Sprintf("method: %s", info.FullMethod), zap.Any("request", req))
		resp, err = handler(ctx, req)

		if err != nil {
			logger.Error(
				fmt.Sprintf("method: %s", info.FullMethod),
				zap.Error(err),
			)
		} else {
			logger.Info(
				fmt.Sprintf("method: %s", info.FullMethod),
				zap.Any("response", resp),
			)
		}

		return resp, err
	}
}

func DeadLetterQueueInterceptor(dlq *resilience.DeadLetterQueue) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			dlq.Add(req)
		}

		return resp, err
	}
}

func RetryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = resilience.RetryOperation(ctx, func(ctx context.Context) (interface{}, error) {
			return handler(ctx, req)
		},
			maxRetries,
			baseDelay,
			maxDelay,
		)
		return resp, err
	}

}

func TimeoutInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		resp, err = resilience.Timeout(ctx, func(ctx context.Context) (interface{}, error) {
			return handler(ctx, req)
		}, timeoutValue)
		return resp, err
	}
}

func NewApp(service v1.SubscriptionService, port int, repo service.SubscriptionRepository, logger logger.Logger) *App {
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		LoggingUnaryServerInterceptor(&logger),
		RetryInterceptor(),
		TimeoutInterceptor(),
	))
	v1.Register(server, service)

	return &App{
		grpcServer: server,
		port:       port,
		repo:       repo,
		logger:     &logger,
	}
}

func (a *App) RunRest(ctx context.Context, grpcPort, gatewayPort int) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := proto.RegisterSubscriptionServiceHandlerFromEndpoint(ctxWithCancel, mux, fmt.Sprintf("localhost:%d", grpcPort), opts)
	if err != nil {
		panic(err)
	}
	a.srv = http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", gatewayPort),
	}
	a.logger.Info(fmt.Sprintf("starting grpc-gateway at %d", gatewayPort))
	if err := a.srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		return fmt.Errorf("grpc.Run: %w", err)
	}

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("grpc.Serve: %w", err)
	}

	return nil
}

func (a *App) GracefulStop(ctx context.Context) {
	err := a.srv.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
	a.grpcServer.GracefulStop()
	a.repo.Close()
}
