package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof" // registers pprof handlers on the default ServeMux
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"

	"gorm.io/gorm"

	"xaults-assignment/config"
	"xaults-assignment/internal/database"
	"xaults-assignment/router"
)

func Init(cfg *config.Config) *gorm.DB {
	db, err := database.InitDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: connect to database: %v\n", err)
		os.Exit(1)
	}
	return db
}

func main() {
	cfg := config.Load()

	e := echo.New()
	db := Init(cfg)

	//register routes
	router.RegisterRoutes(e, db)

	// ── Start with graceful shutdown ───────────────────────────────────────────
	sigCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if cfg.EnablePProf {
		startPProfServer(cfg.PProfPort, sigCtx)
	}

	sc := echo.StartConfig{
		Address:         ":" + cfg.Port,
		GracefulTimeout: 10 * time.Second,
	}

	fmt.Printf("server listening on :%s\n", cfg.Port)
	if err := sc.Start(sigCtx, e); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

func startPProfServer(port string, shutdownCtx context.Context) {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("pprof server failed: %v", err)
		}
	}()

	go func() {
		<-shutdownCtx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("pprof shutdown failed: %v", err)
		} else {
			log.Println("pprof server stopped gracefully")
		}
	}()
}
