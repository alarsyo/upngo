package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"

	_ "github.com/joho/godotenv/autoload"
)

var dir = flag.String("dir", "tusd-files", "where the uploaded files should be stored")

var tokenAuth *jwtauth.JWTAuth

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func init() {
	secret := os.Getenv("JWTAuthSecret")
	tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	// Don't forget to import jwt "github.com/dgrijalva/jwt-go"
	// _, tokenString, _ := tokenAuth.Encode(jwt.MapClaims{"user_id": 123})
	// fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString) */
}

func main() {
	flag.Parse()

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

		// /files/ routes
		filesHandler := http.StripPrefix("/files/", tusdHandler())
		r.Method("GET", "/files/", filesHandler)
		r.Method("POST", "/files/", filesHandler)
		r.Method("HEAD", "/files/", filesHandler)
		r.Method("PATCH", "/files/", filesHandler)
		r.Method("DELETE", "/files/", filesHandler)
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
