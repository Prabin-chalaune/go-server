package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbConfig struct {
	User        string
	Pwd         string
	Host        string
	Port        uint16
	Name        string
	MinPoolSize uint16
	MaxPoolSize uint16
}

type Document[T any] interface {
	EnsureIndexes(Database)
	GetValue() *T
	Validate() error
}

type Database interface {
	GetInstance() *database
	Connect()
	Disconnect()
}

type database struct {
	*mongo.Database
	context context.Context
	config  DbConfig
}

func NewDatabase(ctx context.Context, config DbConfig) Database {
	db := database{
		context: ctx,
		config:  config,
	}
	return &db
}

func (db *database) GetInstance() *database {
	return db
}

func (db *database) Connect() {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s",
		db.config.User, db.config.Pwd, db.config.Host, db.config.Port, db.config.Name,
	)

	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(uint64(db.config.MaxPoolSize))
	clientOptions.SetMaxPoolSize(uint64(db.config.MinPoolSize))

	fmt.Println("Connecting Mongo...")
	client, err := mongo.Connect(db.context, clientOptions)
	if err != nil {
		log.Fatal("Connection to Mongo Failed!: ", err)
	}

	err = client.Ping(db.context, nil)
	if err != nil {
		log.Panic("Pinging to Mongo Failed!: ", err)
	}
	fmt.Println("Connected to Mongo!")

	db.Database = client.Database(db.config.Name)
}

func (db *database) Disconnect() {
	fmt.Println("Disconnecting Mongo...")
	err := db.Client().Disconnect(db.context)
	if err != nil {
		log.Panic(err)
	}
}

func NewObjectID(id string) (primitive.ObjectID, error) {
	i, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = errors.New(id + " is not a valid mongo id")
	}
	return i, err
}
