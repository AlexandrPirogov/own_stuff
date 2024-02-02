package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Coffee struct represents a type of coffee available in the shop
type Coffee struct {
	ID    int    `json:"id" bson:"_id"`
	Name  string `json:"name" bson:"name"`
	Price int    `json:"price" bson:"price"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func main() {
	// Set up MongoDB client
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection = client.Database("coffee_shop").Collection("coffees")

	http.HandleFunc("/coffees", coffeeHandler)
	http.HandleFunc("/buy", buyHandler)

	fmt.Println("Server is listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func coffeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var coffees []Coffee
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var coffee Coffee
		if err := cur.Decode(&coffee); err != nil {
			log.Fatal(err)
		}
		coffees = append(coffees, coffee)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coffees)
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var coffeeID int
	err := json.NewDecoder(r.Body).Decode(&coffeeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var coffee Coffee
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = collection.FindOne(ctx, Coffee{ID: coffeeID}).Decode(&coffee)
	if err != nil {
		http.Error(w, "Coffee not found", http.StatusNotFound)
		return
	}

	response := fmt.Sprintf("You have successfully bought a %s for %d cents.", coffee.Name, coffee.Price)
	w.Write([]byte(response))
}
