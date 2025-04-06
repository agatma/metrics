package servergrpc

import (
	"context"
	"fmt"
	"log"
	"metrics/internal/server/core/domain"
	"metrics/internal/server/logger"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	pb "metrics/internal/proto"
	"metrics/internal/server/config"
)

// MetricService defines the interface for metric operations.
type MetricService interface {
	// SetMetric creates or updates a metric.
	SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error)
}

type GRPCServer struct {
	pb.UnimplementedMetricServiceServer
	metricService MetricService
	cfg           *config.Config
}

// NewGRPC creates a new instance of the GRPC.
func NewGRPC(metricService MetricService, cfg *config.Config) *GRPCServer {
	return &GRPCServer{metricService: metricService, cfg: cfg}
}

func (s *GRPCServer) Update(ctx context.Context, metric *pb.Metric) (*pb.MetricResponse, error) {
	m := domain.Metric{ID: metric.Id}
	if metric.Type == pb.Metric_GAUGE {
		m.MType = domain.Gauge
		m.Value = &metric.Value
	} else {
		m.MType = domain.Counter
		m.Delta = &metric.Delta
	}
	if _, err := s.metricService.SetMetric(ctx, &m); err != nil {
		return &pb.MetricResponse{Status: 13}, nil
	}
	logger.Log.Info("successfully updated metric")
	return &pb.MetricResponse{Status: 0}, nil
}

func (s *GRPCServer) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	grpcServer := grpc.NewServer()

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigint
		grpcServer.GracefulStop()
		idleConnsClosed <- struct{}{}
		close(idleConnsClosed)
	}()
	pb.RegisterMetricServiceServer(grpcServer, s)
	logger.Log.Info("Started gRPC server")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatal(err)
	}
	<-idleConnsClosed
	return nil
}
