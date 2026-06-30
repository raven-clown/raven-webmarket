package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/raven-clown/raven-webmarket/backend/internal/admin"
	"github.com/raven-clown/raven-webmarket/backend/internal/auth"
	"github.com/raven-clown/raven-webmarket/backend/internal/cart"
	"github.com/raven-clown/raven-webmarket/backend/internal/catalog"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/database"
	"github.com/raven-clown/raven-webmarket/backend/internal/delivery"
	"github.com/raven-clown/raven-webmarket/backend/internal/handlers"
	"github.com/raven-clown/raven-webmarket/backend/internal/metrics"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/milestone"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	"github.com/raven-clown/raven-webmarket/backend/internal/order"
	"github.com/raven-clown/raven-webmarket/backend/internal/payment"
	"github.com/raven-clown/raven-webmarket/backend/internal/redeem"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
	"github.com/raven-clown/raven-webmarket/backend/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	esxDB, err := database.ConnectESX(cfg)
	if err != nil {
		log.Printf("esx database warning: %v", err)
		esxDB = db
	} else {
		defer esxDB.Close()
	}
	redis, err := redisstore.Connect(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redis.Close()
	store, err := storage.New(cfg)
	if err != nil {
		log.Printf("storage warning: %v", err)
		store, _ = storage.New(cfg)
	}
	deliverySvc := delivery.New(cfg, db)
	deliverFn := func(ctx context.Context, payload models.DeliveryPayload) error {
		return deliverySvc.Deliver(ctx, payload)
	}
	authSvc := auth.New(cfg, db, esxDB, redis)
	catalogSvc := catalog.New(db, redis)
	cartSvc := cart.New(redis)
	orderSvc := order.New(db, redis, deliverFn)
	milestoneSvc := milestone.New(db, redis, deliverFn)
	redeemSvc := redeem.New(db, redis, deliverFn)
	paymentSvc := payment.New(cfg, db, redis, store)
	adminSvc := admin.New(db)
	api := handlers.NewAPI(cfg, authSvc, catalogSvc, cartSvc, orderSvc, milestoneSvc, redeemSvc, paymentSvc, deliverySvc, adminSvc)
	r := chi.NewRouter()
	r.Use(middleware.Metrics)
	r.Use(middleware.JSONContentType)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RateLimit(cfg, redis))
	r.Get("/metrics", metrics.Handler().ServeHTTP)
	authMw := middleware.AuthRequired(cfg)
	adminMw := middleware.AdminRequired
	api.Routes(r, authMw, adminMw)
	addr := fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort)
	srv := &http.Server{Addr: addr, Handler: r, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	go func() {
		log.Printf("api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
