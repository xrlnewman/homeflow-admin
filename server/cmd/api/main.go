package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xrlnewman/homeflow-admin/server/internal/config"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/cache"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/database"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/store"
	"github.com/xrlnewman/homeflow-admin/server/internal/transport/httpapi"
)

func main() {
	cfg := config.Load()
	st := store.NewMemoryStore()
	deps := httpapi.Dependencies{}
	if db, err := database.Open(cfg.DatabaseDSN); err != nil {
		slog.Warn("MySQL 未连接，使用内存演示模式", "error", err)
	} else {
		deps.DB = db
		defer db.Close()
	}
	deps.Redis = cache.NewRedisLocker(cfg.RedisAddr, cfg.RedisDB)
	server := &http.Server{Addr: cfg.Addr, Handler: httpapi.NewRouterWithDeps(cfg, st, deps), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		slog.Info("HomeFlow API 已启动", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP 服务异常", "error", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
