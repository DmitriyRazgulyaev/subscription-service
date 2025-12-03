package grpcApp

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	v1 "subscription-service/internal/transport/grpc/v1"
)

type App struct {
	grpcServer *grpc.Server
	port       int
}

func NewApp(service v1.SubscriptionService, port int) *App {
	server := grpc.NewServer()
	v1.Register(server, service)

	return &App{
		grpcServer: server,
		port:       port,
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

func (a *App) GracefulStop() {
	a.grpcServer.GracefulStop()
}
