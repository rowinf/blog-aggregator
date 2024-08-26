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

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

type ApiConfig struct {
	DB *database.Queries
}

type UserParams struct {
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Id        uuid.UUID `json:"id"`
	ApiKey    string    `json:"apikey"`
}

type FeedParams struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Id        string `json:"id"`
	Url       string `json:"url"`
	UserId    string `json:"user_id"`
}

type FeedFollowsParams struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	FeedId    string `json:"feed_id"`
	UserId    string `json:"user_id"`
}

type FeedCreationParams struct {
	Feed       *FeedParams        `json:"feed"`
	FeedFollow *FeedFollowsParams `json:"feed_follow"`
}

func (params *FeedCreationParams) asJSON(feed database.Feed, feedFollow database.FeedFollow) *FeedCreationParams {
	feedParams := FeedParams{}
	feedFollowParams := FeedFollowsParams{}
	params.Feed = feedParams.asJSON(feed)
	params.FeedFollow = feedFollowParams.asJSON(feedFollow)
	return params
}

func (params *FeedParams) asJSON(feed database.Feed) *FeedParams {
	params.Id = feed.ID
	params.CreatedAt = feed.CreatedAt.Format(time.RFC3339)
	params.UpdatedAt = feed.UpdatedAt.Format(time.RFC3339)
	params.Name = feed.Name
	params.Url = feed.Url
	params.UserId = feed.UserID
	return params
}

func (params *FeedFollowsParams) asJSON(feedFollow database.FeedFollow) *FeedFollowsParams {
	params.Id = feedFollow.ID
	params.CreatedAt = feedFollow.CreatedAt.Format(time.RFC3339)
	params.UpdatedAt = feedFollow.UpdatedAt.Format(time.RFC3339)
	params.UserId = feedFollow.UserID
	params.FeedId = feedFollow.FeedID
	return params
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

func (cfg *ApiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := internal.GetHeaderApiKey(w, r)
		if err != nil {
			internal.RespondWithError(w, http.StatusBadRequest, "no api key")
		} else {
			user, uerr := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
			if uerr != nil {
				internal.RespondWithError(w, http.StatusBadRequest, "invalid api key")
			} else {
				handler(w, r, user)
			}
		}
	}
}

func (cfg *ApiConfig) handleUsersGet(w http.ResponseWriter, r *http.Request, user database.User) {
	internal.RespondWithJSON(w, http.StatusOK, UserParams{
		Id:        uuid.MustParse(user.ID),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		ApiKey:    user.Apikey,
		Name:      user.Name,
	})
}

func (cfg *ApiConfig) handleFeedsPost(w http.ResponseWriter, r *http.Request, user database.User) {
	body := FeedParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	feed, err := cfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      body.Name,
		Url:       body.Url,
		UserID:    user.ID,
	})
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    feed.UserID,
	})
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload := FeedCreationParams{}
	internal.RespondWithJSON(w, http.StatusCreated, payload.asJSON(feed, feedFollow))
}

func (cfg *ApiConfig) handleFeedsGet(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.GetAllFeeds(r.Context())
	if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload := make([]FeedParams, len(feeds))
	for i := range payload {
		payload[i].asJSON(feeds[i])
	}
	internal.RespondWithJSON(w, http.StatusOK, payload)
}

func (cfg *ApiConfig) handleFeedFollowsPost(w http.ResponseWriter, r *http.Request, user database.User) {
	body := FeedFollowsParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
	} else {
		feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
			ID:        uuid.NewString(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			FeedID:    body.FeedId,
			UserID:    user.ID,
		})
		if err != nil {
			internal.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		paylod := FeedFollowsParams{}
		internal.RespondWithJSON(w, http.StatusOK, paylod.asJSON(feedFollow))
	}
}

func (cfg *ApiConfig) handleFeedFollowsDelete(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowID := r.PathValue("feedFollowID")
	feedFollow, err := cfg.DB.DeleteFeedFollow(r.Context(), feedFollowID)
	if err != nil {
		internal.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	payload := FeedFollowsParams{}
	internal.RespondWithJSON(w, http.StatusNoContent, payload.asJSON(feedFollow))
}

func (cfg *ApiConfig) handleFeedFollowsGet(w http.ResponseWriter, r *http.Request, user database.User) {
	feeds, err := cfg.DB.GetFeedFollowsByUserId(r.Context(), user.ID)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload := make([]FeedFollowsParams, len(feeds))
	for index, feed := range feeds {
		payload[index].asJSON(feed)
	}
	internal.RespondWithJSON(w, http.StatusOK, payload)
}

func main() {
	godotenv.Load()
	db, err := sql.Open("postgres", os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		panic("database error")
	}
	apiConfig := ApiConfig{
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
	r.HandleFunc("GET /v1/users", apiConfig.middlewareAuth(apiConfig.handleUsersGet))
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
			user, err := apiConfig.DB.CreateUser(r.Context(), payload)
			if err != nil {
				internal.RespondWithError(w, http.StatusBadRequest, err.Error())
			} else {
				internal.RespondWithJSON(w, http.StatusCreated, UserParams{
					Id:        uuid.MustParse(user.ID),
					CreatedAt: user.CreatedAt.String(),
					UpdatedAt: user.UpdatedAt.String(),
					Name:      user.Name,
					ApiKey:    user.Apikey,
				})
			}
		}
	})
	r.HandleFunc("POST /v1/feeds", apiConfig.middlewareAuth(apiConfig.handleFeedsPost))
	r.HandleFunc("GET /v1/feeds", apiConfig.handleFeedsGet)
	r.HandleFunc("GET /v1/feed_follows", apiConfig.middlewareAuth(apiConfig.handleFeedFollowsGet))
	r.HandleFunc("POST /v1/feed_follows", apiConfig.middlewareAuth(apiConfig.handleFeedFollowsPost))
	r.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", apiConfig.middlewareAuth(apiConfig.handleFeedFollowsDelete))

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
