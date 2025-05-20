package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ApodCollection *mongo.Collection

func Connect() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DATABASE")
	collectionName := os.Getenv("MONGODB_COLLECTION")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ApodCollection = client.Database(dbName).Collection(collectionName)
	log.Println("MongoDB conectado com sucesso.")
}
