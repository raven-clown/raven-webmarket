package database

import (
	"context"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func ConnectMongo(cfg *config.Config) (*MongoStore, error) {
	if !cfg.MongoEnabled {
		return nil, nil
	}
	uri := cfg.MongoURI
	if uri == "" {
		auth := ""
		if cfg.MongoUser != "" {
			auth = cfg.MongoUser
			if cfg.MongoPassword != "" {
				auth += ":" + cfg.MongoPassword
			}
			auth += "@"
		}
		uri = fmt.Sprintf("mongodb://%s%s:%s/%s", auth, cfg.MongoHost, cfg.MongoPort, cfg.MongoDBName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return &MongoStore{
		Client: client,
		DB:     client.Database(cfg.MongoDBName),
	}, nil
}

func (m *MongoStore) Close(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return nil
	}
	return m.Client.Disconnect(ctx)
}

func (m *MongoStore) Ping(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return fmt.Errorf("mongo not connected")
	}
	return m.Client.Ping(ctx, nil)
}

func PingMongoOptional(ctx context.Context, cfg *config.Config) error {
	m, err := ConnectMongo(cfg)
	if err != nil {
		return err
	}
	defer m.Close(ctx)
	return m.Ping(ctx)
}

func EnsureMongoIndexes(ctx context.Context, m *MongoStore) error {
	if m == nil || m.DB == nil {
		return nil
	}
	_, err := m.DB.Collection("activity_events").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"created_at": -1},
	})
	return err
}
