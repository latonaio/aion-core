package my_mongo

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/avast/retry-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	instance = &MongoClient{}
)

const (
	connRetryCount = 10
)

type MongoClient struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func GetInstance() *MongoClient {
	return instance
}

func (c *MongoClient) CreatePool(ctx context.Context, addr string, db string, collection string) error {
	uri := fmt.Sprintf("mongodb://%s", addr)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("cant connect to mongo (Address: %s)", addr)
	}

	if err := retry.Do(
		func() error {
			if err := client.Ping(ctx, readpref.Primary()); err != nil {
				return fmt.Errorf("cant connect to mongo (Address: %s)", addr)
			}
			return nil
		},
		retry.DelayType(func(n uint, config *retry.Config) time.Duration {
			log.Printf("Retry connecting to mongo (Addr:%s)", addr)
			return time.Second
		}),
		retry.Attempts(connRetryCount),
	); err != nil {
		return err
	}

	c.client = client
	c.db = c.client.Database(db)
	c.collection = c.db.Collection(collection)
	return nil
}

func (c *MongoClient) InsertOne(ctx context.Context, document interface{}) error {
	_, err := c.collection.InsertOne(ctx, document)
	if err != nil {
		return fmt.Errorf("cant write to mongo, %v", err)
	}
	return nil
}
