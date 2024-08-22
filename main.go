package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rowinf/blog-aggregator/internal"
	"github.com/rowinf/blog-aggregator/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

// addCorsHeaders is a middleware function that adds CORS headers to the response.
func addCorsHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// If it's a preflight request, respond with 200 OK
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

type UserParams struct {
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Id        uuid.UUID `json:"id"`
}

func main() {
	godotenv.Load()
	db, err := sql.Open("postgres", os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		panic("database error")
	}
	config := apiConfig{
		DB: database.New(db),
	}
	r := http.NewServeMux()
	port := os.Getenv("PORT")
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		internal.RespondWithJSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})
	r.HandleFunc("/v1/err", func(w http.ResponseWriter, _ *http.Request) {
		internal.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	})
	r.HandleFunc("POST /v1/users", func(w http.ResponseWriter, r *http.Request) {
		body := UserParams{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			payload := database.CreateUserParams{
				Name:      body.Name,
				ID:        uuid.New().String(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			user, err := config.DB.CreateUser(r.Context(), payload)
			if err != nil {
				internal.RespondWithError(w, http.StatusBadRequest, err.Error())
			} else {
				internal.RespondWithJSON(w, http.StatusCreated, UserParams{
					Id:        uuid.MustParse(user.ID),
					CreatedAt: user.CreatedAt.String(),
					UpdatedAt: user.UpdatedAt.String(),
					Name:      user.Name,
				})
			}
		}
	})
	corsMux := addCorsHeaders(r)
	// Create a new HTTP server with the corsMux as the handler
	server := &http.Server{
		Addr:    ":" + port, // Set the desired port
		Handler: corsMux,
	}

	// Start the server
	log.Printf("Serving files from %s on port: %s\n", ".", port)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
