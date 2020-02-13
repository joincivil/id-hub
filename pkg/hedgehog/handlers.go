package hedgehog

import (
	"encoding/json"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"net/http"
	"strings"
)

// Dependencies contains objects needed to add hedgehog routes
type Dependencies struct {
	Router chi.Router
	Db     *gorm.DB
}

// AddRoutes configures the router to serve routes for Hedgehog storage
func AddRoutes(deps Dependencies) {

	// Set some rate limiters for the invoice handlers
	limiter := tollbooth.NewLimiter(4, nil) // 4 req/sec max
	limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	service := Service{
		db: deps.Db,
	}

	handler := Handler{service: service}

	deps.Router.Route(fmt.Sprintf("/%v/hedgehog", "v1"), func(r chi.Router) {

		// rate limited routes
		r.Group(func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(limiter))
			r.Post("/users", handler.PostUserHandler)

			r.Post("/authentication", handler.PostAuthenticationHandler)
			r.Get("/authentication/{lookupKey}", handler.GetAuthenticationHandler)
		})

		r.Post("/users/{username}/store/{key}", handler.PostUserData)
		r.Get("/users/{username}/store/{key}", handler.GetUserData)

	})

}

// Handler contains the http handler methods for hedgehog
type Handler struct {
	service Service
}

// GetAuthenticationHandler serves GET /authentication/{lookupKey}
func (s *Handler) GetAuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	lookupKey := chi.URLParam(r, "lookupKey")
	fmt.Printf("lookupKey: %s\n", lookupKey)

	data, err := s.service.GetAuthData(lookupKey)
	if err == ErrorNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)

}

// PostAuthenticationHandler serves the POST /authentication route
func (s *Handler) PostAuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	var p = struct {
		EncryptedData
		LookupKey string `json:"lookupKey"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.service.SetAuthData(p.LookupKey, p.EncryptedData)
	if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "ok")
}

// PostUserHandler services the POST /users route
func (s *Handler) PostUserHandler(w http.ResponseWriter, r *http.Request) {
	var p = struct {
		Username string `json:"username"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("PostUserHandler: %s \n", p.Username)

	err = s.service.ReserveUsername(p.Username)
	if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, "user exists", http.StatusBadRequest)
		// TODO(dankins): return proper status code if already exists
		return
	}

	_, _ = fmt.Fprintf(w, "ok")

}

// PostUserData serves the POST /users/{username}/store/{key} route
func (s *Handler) PostUserData(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(chi.URLParam(r, "username"))
	key := chi.URLParam(r, "key")

	fmt.Printf("PostUserData: %s %s \n", username, key)

	var data = EncryptedData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.service.StoreItem(username, key, data)
	if err != nil {
		fmt.Printf("error %v", err)
		http.Error(w, "key exists", http.StatusBadRequest)
		// TODO(dankins): return proper status code if already exists
		return
	}

	_, _ = fmt.Fprintf(w, "ok")

}

// GetUserData serves the GET /users/{username}/store/{key} route
func (s *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(chi.URLParam(r, "username"))
	key := chi.URLParam(r, "key")

	fmt.Printf("GetUserData: %s %s \n", username, key)
	data, err := s.service.GetItem(username, key)
	if err != nil {
		// TODO(dankins): handle not found vs actual error
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}
