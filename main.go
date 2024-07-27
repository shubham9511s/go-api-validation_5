package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// APIKeyValidationRequest represents the request structure for validation
type APIKeyValidationRequest struct {
	APIKey string `json:"apiKey"`
}

// APIKeyValidationResponse represents the response structure
type APIKeyValidationResponse struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
}

var client *mongo.Client

func validateAPIKeyInDB(apiKey string) (bool, error) {
	collection := client.Database("todo").Collection("api_keys")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func validateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req APIKeyValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	isValid, err := validateAPIKeyInDB(req.APIKey)
	if err != nil {
		http.Error(w, "Error validating API key", http.StatusInternalServerError)
		return
	}

	resp := APIKeyValidationResponse{
		IsValid: isValid,
		Message: "API Key validation successful",
	}
	if !isValid {
		resp.Message = "Invalid API Key"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://admin:password@52.66.247.95:27017/todo?directConnection=true&appName=mongosh+2.2.12"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	http.HandleFunc("/validate-api-key", validateAPIKeyHandler)

	log.Println("Server starting on port 9090...")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
