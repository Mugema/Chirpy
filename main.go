package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"github.com/Mugema/Chirpy/internal/auth"
	"github.com/Mugema/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

import _ "github.com/lib/pq"

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	secret         string
	polkaKey       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, _ := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	router := http.NewServeMux()

	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = router

	apiCfg := apiConfig{db: dbQueries, secret: os.Getenv("secret"), polkaKey: os.Getenv("POLKA_KEY")}

	router.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	router.HandleFunc("GET /api/healthz/", handleHealth)
	router.HandleFunc("GET /admin/metrics", apiCfg.handlerNumberRequests)
	router.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	router.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)
	router.HandleFunc("/admin/reset", apiCfg.handlerReset)
	router.HandleFunc("POST /api/users", apiCfg.handlerUsers)
	router.HandleFunc("POST /api/chirps", apiCfg.handlerChirp)
	router.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	router.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	router.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	router.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	router.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	router.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUpgradeUser)

	err := server.ListenAndServe()
	if err != nil {
		return
	}
}

type bodyResp struct {
	Body string `json:"body"`
}

type errorResp struct {
	ErrorResp string `json:"error"`
}

type validResp struct {
	Valid string `json:"cleaned_body"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreateAt  time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	ChirpyRed bool      `json:"is_chirpy_red"`
}

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	CreateAt  time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(writer, req)
	})
}

func handleHealth(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerNumberRequests(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Add("content-type", "text/html")
	value := fmt.Sprintf("<html><body>"+
		"\n\t<h1>Welcome, Chirpy Admin</h1>"+
		"\n\t<p>Chirpy has been visited %d times!</p>"+
		"\n\t</body>"+
		"\n\t</html>", cfg.fileServerHits.Load())
	_, err := writer.Write([]byte(value))
	if err != nil {
		return
	}
}

func (cfg *apiConfig) handlerReset(writer http.ResponseWriter, req *http.Request) {
	cfg.db.Reset(req.Context())

	writer.Write([]byte("Database reset"))
}

func (cfg *apiConfig) handlerUsers(writer http.ResponseWriter, req *http.Request) {
	type resp struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	user1 := resp{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user1)
	if err != nil {
		fmt.Println("Error decoding")
		return
	}

	hashString, err := auth.HashPassword(user1.Password)
	if err != nil {
		fmt.Println("Error hashing password")
		return
	}
	fmt.Println(hashString)

	user, err := cfg.db.CreateUser(
		req.Context(),
		database.CreateUserParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().Local(),
			UpdatedAt:   time.Now().Local(),
			Email:       user1.Email,
			Password:    hashString,
			IsChirpyRed: false})

	if err != nil {
		fmt.Println("Error creating user")
		return
	}

	createdUser := jsonUserMapper(user)

	data, err := json.Marshal(createdUser)
	if err != nil {
		fmt.Println("Error marshaling the created user")
		return
	}

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(201)

	writer.Write(data)

}

func jsonUserMapper(u database.User) User {
	return User{
		u.ID,
		u.CreatedAt,
		u.UpdatedAt,
		u.Email,
		u.IsChirpyRed}
}

func (cfg *apiConfig) handlerGetChirps(writer http.ResponseWriter, req *http.Request) {
	var c []database.Chirp
	var err error

	queryParameter := req.URL.Query().Get("author_id")
	sortOrder := req.URL.Query().Get("sort")

	if queryParameter == "" {
		c, err = cfg.db.GetChirps(req.Context())
	} else {
		id, err := uuid.Parse(queryParameter)
		if err != nil {
			return
		}
		c, err = cfg.db.GetChirpByUserId(req.Context(), id)
	}
	if err != nil {
		fmt.Println("Error retrieving Chirps")
		return
	}

	chirps := make([]Chirp, 0)
	for _, chirp := range c {
		chirps = append(chirps, chirpMapper(chirp))
	}
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreateAt.After(chirps[j].CreateAt) })
		fmt.Printf("DESC ordering: %v\n", chirps)
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		fmt.Println("Error Marshaling the data")
		return
	}

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(http.StatusOK)

	writer.Write(data)
	return

}

func (cfg *apiConfig) handlerChirp(writer http.ResponseWriter, req *http.Request) {
	type request struct {
		Body string `json:"body"`
	}
	reqChirp := request{}

	token, err := auth.GetBearerToken(req.Header)

	if err != nil {
		fmt.Printf("No token provided in header. Error: %v", err)
		writer.WriteHeader(401)
		return
	} else if token == "" {
		fmt.Printf("No token provided %v", err)
		writer.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		fmt.Printf("token:%v \n error: %v\n", token, err)
		writer.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&reqChirp)
	if err != nil {
		fmt.Println("Error decoding")
		return
	}

	if len(reqChirp.Body) > 140 {
		errorResponse := errorResp{"Chirp too long"}

		data, _ := json.Marshal(errorResponse)

		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(400)
		writer.Write(data)

		return
	}

	createChirp, err := cfg.db.CreateChirp(req.Context(),
		database.CreateChirpParams{
			ID:        uuid.New(),
			UserID:    userID,
			CreatedAt: time.Now().Local(),
			UpdatedAt: time.Now().Local(),
			Body:      reqChirp.Body,
		})
	if err != nil {
		fmt.Printf("Error creating chirp %v", err)
		return
	}

	data, _ := json.Marshal(chirpMapper(createChirp))

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(201)

	writer.Write(data)

	return
}

func chirpMapper(chirp database.Chirp) Chirp {
	return Chirp{
		chirp.ID,
		chirp.UserID,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
	}
}

func (cfg *apiConfig) handlerGetChirpByID(writer http.ResponseWriter, req *http.Request) {
	chirpId, _ := uuid.Parse(req.PathValue("chirpID"))

	chirp, err := cfg.db.GetChirpById(req.Context(), chirpId)
	if err != nil {
		fmt.Println("Error getting chirp from the database")
		writer.WriteHeader(404)
		return
	}

	fmt.Println(chirp)

	c := chirpMapper(chirp)

	data, err := json.Marshal(c)
	if err != nil {
		fmt.Println("Error marshaling")
		return
	}

	writer.Header().Set("content-type", "application")
	writer.WriteHeader(http.StatusOK)

	writer.Write(data)
}
