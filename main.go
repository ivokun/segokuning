package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/jwx/v2/jwt"
	_ "github.com/lib/pq" // PostgreSQL driver
	"log"
	"net/http"
	"time"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	Message    string `json:"message"` // user-level status message
	StatusText string `json:"status"`  // user-level status message
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request",
		Message:        err.Error(),
	}
}

func ErrServer(err error, statusCode int) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: statusCode,
		StatusText:     "Server error",
		Message:        err.Error(),
	}
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func init() {
	jwtSecret := GenerateJWTSecret()
	TokenAuth = jwtauth.New("HS256", []byte(jwtSecret), nil, jwt.WithAcceptableSkew(30*time.Minute))
}

func main() {
	// Connection string for a remote PostgreSQL database
	connStr := GenerateConnectionString()

	// Open a connection to the remote database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec(`SELECT 1`)
	if err != nil {
		log.Fatalln(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Mount("/v1", AppRouter(db))
	http.ListenAndServe(":8000", r)
}

func AppRouter(db *sqlx.DB) chi.Router {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok gas, ok gas"))
	})
	r.Mount("/user", UserRouter(db))
	return r
}
