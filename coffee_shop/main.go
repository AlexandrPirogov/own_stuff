package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
		panic(err)
	}
	defer client.Disconnect(ctx)

	collection = client.Database("coffee_shop").Collection("coffees")

	// Initialize Chi router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)

	// Define routes
	router.Get("/coffees", coffeeHandler)
	router.Post("/buy", buyHandler)
	router.Post("/coffees", createCoffeeHandler) // New handler for creating a coffee instance

	// Start server
	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", router)
}

func coffeeHandler(w http.ResponseWriter, r *http.Request) {
	// Functionality to retrieve coffees from the database remains the same
	// Omitted for brevity
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	// Functionality to buy coffee remains the same
	// Omitted for brevity
}

func createCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	// Define a struct to receive JSON request
	type CoffeeRequest struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}

	// Parse JSON request body
	var coffeeReq CoffeeRequest
	err := json.NewDecoder(r.Body).Decode(&coffeeReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert new coffee instance into the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, Coffee{Name: coffeeReq.Name, Price: coffeeReq.Price})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	response := fmt.Sprintf("New coffee created: %s - Price: %d", coffeeReq.Name, coffeeReq.Price)
	w.Write([]byte(response))
}
