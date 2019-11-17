package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"

	_ "github.com/joho/godotenv/autoload"
)

var debug = flag.Bool("debug", false, "enable debugging output")
var dir = flag.String("dir", "tusd-files", "where the uploaded files should be stored")
var tokenAuth *jwtauth.JWTAuth

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func init() {
	flag.Parse()
	secret := os.Getenv("JWTAuthSecret")
	if secret == "" {
		fmt.Fprintf(os.Stderr, "Please set up the JWTAuthSecret env variable\n")
		os.Exit(1)
	}
	tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	if *debug {
		_, tokenString, _ := tokenAuth.Encode(jwt.MapClaims{"user_id": 123})
		fmt.Fprintf(os.Stderr, "DEBUG: a sample jwt is %s\n\n", tokenString)
	}
}

func main() {
	_, err := os.Stat(*dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(*dir, 0777)
	}
	err = http.ListenAndServe(":8000", router())
	if err != nil {
		panic(fmt.Errorf("Unable to listen: %s", err))
	}
}

func router() http.Handler {
	r := chi.NewRouter()

	//Protected routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(tokenAuth))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		r.Use(jwtauth.Authenticator)

		// /files/ route
		r.Route("/files", func(r chi.Router) {
			filesHandler := http.StripPrefix("/files/", tusdHandler())
			r.Method("GET", "/{id}", filesHandler)
			r.Method("POST", "/", filesHandler)
			r.Method("HEAD", "/{id}", filesHandler)
			r.Method("PATCH", "/{id}", filesHandler)
			r.Method("DELETE", "/{id}", filesHandler)
		})
	})

	//Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", hello)
	})

	return r
}

func tusdHandler() http.Handler {
	store := filestore.New(*dir)

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/files/",
		StoreComposer:           composer,
		RespectForwardedHeaders: true,
		NotifyCompleteUploads:   true,
	})
	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}

	go func() {
		for {
			event := <-handler.CompleteUploads
			fmt.Printf("Upload %s finished\n", event.Upload.ID)
		}
	}()

	return handler
}
