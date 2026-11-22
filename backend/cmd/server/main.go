package main

import (
	"context"
	demoassets "github.com/EthanLe-dev/blue-sky-group-website/backend/demo-assets"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/platform"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		client := http.Client{Timeout: 2 * time.Second}
		resp, e := client.Get("http://127.0.0.1:" + port + "/readyz")
		if e != nil {
			os.Exit(1)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			os.Exit(1)
		}
		return
	}

	c, e := platform.LoadConfig()
	if e != nil {
		slog.Error("configuration rejected", "reason", e.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p, e := platform.Connect(ctx, c.DatabaseURL)
	if e != nil {
		slog.Error("database unavailable")
		os.Exit(1)
	}
	defer p.Close()
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if e = platform.Migrate(ctx, p); e != nil {
			slog.Error("migration failed")
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "bootstrap" {
		if e = platform.Bootstrap(ctx, p, os.Getenv("BOOTSTRAP_ADMIN_EMAIL"), os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")); e != nil {
			slog.Error("bootstrap failed", "reason", e.Error())
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "seed-demo" {
		if e = catalog.Seed(context.Background(), p, catalog.NewMediaStore(c.CSRFSecret), demoassets.Files); e != nil {
			slog.Error("demo seed failed", "reason", e.Error())
			os.Exit(1)
		}
		if e = platform.New(c, p).BookingServer().SeedDemoStates(context.Background()); e != nil {
			slog.Error("demo scenarios seed failed", "reason", e.Error())
			os.Exit(1)
		}
		return
	}
	app := platform.New(c, p)
	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()
	go app.BookingServer().RunWorker(workerCtx)
	if c.Env != "production" && c.MockNotification {
		go app.BookingServer().RunNotifications(workerCtx)
	}
	srv := &http.Server{Addr: c.Address, Handler: app.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			slog.Error("server stopped")
			os.Exit(1)
		}
	}()
	<-done
	shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	srv.Shutdown(shutdown)
}
