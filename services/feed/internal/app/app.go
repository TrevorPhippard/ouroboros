package app

import (
	"context"
	"net"
	"net/http"

	"feed/internal/config"
	"feed/internal/discovery"
	"feed/internal/feed"
	"feed/internal/infra/redisx"
	"feed/internal/messaging"
	"feed/internal/transport"

	pbFeed "ouroboros/proto/generated/feed"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type App struct {
	cfg      config.Config
	grpcSrv  *grpc.Server
	httpSrv  *http.Server
	consumer *messaging.Consumer
}

func New(cfg config.Config) (*App, error) {
	// Infra
	rdb := redisx.New(cfg.RedisAddr)

	// Store + Service
	store := feed.NewStore(rdb)
	social := feed.NewSocialGraph()
	fanout := feed.NewFanoutEngine()

	svc := feed.NewService(store, social, fanout)

	// gRPC
	grpcSrv := transport.NewServer()
	pbFeed.RegisterFeedServiceServer(grpcSrv, svc)

	// HTTP
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("OK"))
	})

	httpSrv := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: mux,
	}

	// Kafka
	consumer := messaging.New(cfg, svc)

	return &App{
		cfg:      cfg,
		grpcSrv:  grpcSrv,
		httpSrv:  httpSrv,
		consumer: consumer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	// HTTP
	go a.httpSrv.ListenAndServe()

	// Kafka
	go a.consumer.Run(ctx)

	// gRPC
	lis, err := net.Listen("tcp", a.cfg.GRPCPort)
	if err != nil {
		return err
	}

	discovery.Register("consul:8500", "feed-service", 50055)

	go a.grpcSrv.Serve(lis)

	<-ctx.Done()

	a.grpcSrv.GracefulStop()
	a.httpSrv.Shutdown(context.Background())

	return nil
}
