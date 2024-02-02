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

type RequestLog struct {
	Path        string    `json:"path"`
	Method      string    `json:"method"`
	Host        string    `json:"host"`
	RemoteAddr  string    `json:"remote_addr"`
	RequestTime time.Time `json:"request_time"`
}

type Coffee struct {
	ID    string `json:"_id" bson:"_id"`
	Name  string `json:"name" bson:"name"`
	Price int    `json:"price" bson:"price"`
}

var (
	client               *mongo.Client
	coffeesCollection    *mongo.Collection
	requestsCollection   *mongo.Collection
)

func main() {
	// Set up MongoDB client
	clientOptions := options.Client().ApplyURI("mongodb://admin:admin@coffee_db:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	
	coffeesCollection = client.Database("coffee_shop").Collection("coffees")
	requestsCollection = client.Database("coffee_shop").Collection("requests")


	// Initialize Chi router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)
	router.Use(logRequestsToMongo)

	// Define routes
	router.Get("/coffees", coffeeHandler)
	router.Post("/buy", buyHandler)
	router.Post("/import", importCoffeeHandler) // New handler for importing coffee instances

	// Start server
	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", router)
}

func coffeeHandler(w http.ResponseWriter, r *http.Request) {
	var coffees []Coffee
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var coffee Coffee
		if err := cur.Decode(&coffee); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		coffees = append(coffees, coffee)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coffees)
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	var coffeeID string
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

func logRequestsToMongo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLog := RequestLog{
			Path:        r.URL.Path,
			Method:      r.Method,
			Host:        r.Host,
			RemoteAddr:  r.RemoteAddr,
			RequestTime: time.Now(),
		}

		// Convert to JSON
		requestLogJSON, err := json.Marshal(requestLog)
		if err != nil {
			log.Println("Error marshalling request log:", err)
		}

		// Insert request log into MongoDB collection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = requestsCollection.InsertOne(ctx, requestLog)
		if err != nil {
			log.Println("Error inserting request log into MongoDB:", err)
		}

		next.ServeHTTP(w, r)
	})
}
