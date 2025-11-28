package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Parse flags
	drop := flag.Bool("drop", false, "Drop all indexes before creating")
	list := flag.String("list", "", "List indexes for a specific collection")
	dbName := flag.String("db", "", "Database name (overrides env)")
	flag.Parse()

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get MongoDB connection string
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is required")
	}

	// Get database name
	database := *dbName
	if database == "" {
		database = os.Getenv("MONGODB_DATABASE")
		if database == "" {
			database = "leetgaming" // Default
		}
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal("Failed to disconnect from MongoDB:", err)
		}
	}()

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	fmt.Printf("Connected to MongoDB successfully (database: %s)\n\n", database)

	// Handle list command
	if *list != "" {
		if err := listIndexesForCollection(ctx, client, database, *list); err != nil {
			log.Fatal("Failed to list indexes:", err)
		}
		return
	}

	// Handle drop command
	if *drop {
		fmt.Println("🗑️  Dropping existing indexes...")
		if err := db.DropAllIndexes(ctx, client, database); err != nil {
			log.Fatal("Failed to drop indexes:", err)
		}
		fmt.Println()
	}

	// Create indexes
	fmt.Println("🔨 Creating indexes...")
	if err := db.CreateIndexes(ctx, client, database); err != nil {
		log.Fatal("Failed to create indexes:", err)
	}

	fmt.Println("\n✅ Index migration completed successfully!")
}

func listIndexesForCollection(ctx context.Context, client *mongo.Client, dbName, collectionName string) error {
	slog.Info("Listing indexes", "collection", collectionName)

	indexes, err := db.ListIndexes(ctx, client, dbName, collectionName)
	if err != nil {
		return err
	}

	fmt.Printf("Indexes for collection '%s':\n", collectionName)
	fmt.Println("=====================================")
	for i, idx := range indexes {
		fmt.Printf("\n%d. %v\n", i+1, idx["name"])
		fmt.Printf("   Keys: %v\n", idx["key"])
		if unique, ok := idx["unique"]; ok && unique == true {
			fmt.Println("   Unique: true")
		}
		if ttl, ok := idx["expireAfterSeconds"]; ok {
			fmt.Printf("   TTL: %v seconds\n", ttl)
		}
	}
	fmt.Println()

	return nil
}
