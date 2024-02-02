package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Coffee struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
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
	router.Post("/import", importCoffeeHandler) // New handler for importing coffee instances

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

func importCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	// Read coffee instances from JSON file
	coffees, err := readCoffeeFile("coffees.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert coffee instances into the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, coffee := range coffees {
		_, err := collection.InsertOne(ctx, coffee)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Respond with success message
	response := fmt.Sprintf("Imported %d coffee instances into the database", len(coffees))
	w.Write([]byte(response))
}

func readCoffeeFile(filename string) ([]Coffee, error) {
	// Open JSON file
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON content into slice of Coffee structs
	var coffees []Coffee
	if err := json.Unmarshal(file, &coffees); err != nil {
		return nil, err
	}

	return coffees, nil
}
