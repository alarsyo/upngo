package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"

	_ "github.com/joho/godotenv/autoload"

	models "github.com/alarsyo/upngo/models"
	routes "github.com/alarsyo/upngo/routes"
)

var debug = flag.Bool("debug", false, "enable debugging output")
var dir = flag.String("dir", "tusd-files", "where the uploaded files should be stored")

// The following code to extract id from a path is from https://github.com/tus/tusd
var reExtractFileID = regexp.MustCompile(`([^/]+)\/?$`)

// extractIDFromPath pulls the last segment from the url provided
func extractIDFromPath(url string) (string, error) {
	result := reExtractFileID.FindStringSubmatch(url)
	if len(result) != 2 {
		return "", errors.New("Id not found")
	}
	return result[1], nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func init() {
	models.InitDb()
	flag.Parse()
	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	if *debug {
		_, tokenString, _ := models.TokenAuth.Encode(jwt.MapClaims{"user_id": 123})
		fmt.Fprintf(os.Stderr, "DEBUG: a sample jwt is %s\n\n", tokenString)
	}
}

func main() {
	_, err := os.Stat(*dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(*dir, 0777)
	}
	defer models.DB.Close()
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
		r.Use(jwtauth.Verifier(models.TokenAuth))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		r.Use(jwtauth.Authenticator)

		// /files/ routes
		r.Route("/files", func(r chi.Router) {
			r.Get("/", routes.GetFilesInfo)
			r.Get("/{id}", getHandler)
		})

		// /tus/ routes
		r.Route("/tus", func(r chi.Router) {
			filesHandler := http.StripPrefix("/tus/", tusdHandler())
			r.Post("/", postHandler)
			r.Method("HEAD", "/{id}", filesHandler)
			r.Method("PATCH", "/{id}", filesHandler)
			r.Delete("/{id}", deleteHandler)
		})
	})

	//Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", hello)
		r.Post("/login", routes.Login)
		r.Post("/signup", routes.Create)
	})

	return r
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := fmt.Sprintf("%v", claims["user_id"])
	User, ok := strconv.Atoi(userId)
	if ok == nil {
		FileId, err := extractIDFromPath(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = models.DeleteFile(uint(User), FileId)
		if err == nil {
			store := filestore.New(*dir)
			composer := tusd.NewStoreComposer()
			store.UseIn(composer)
			handler, err := tusd.NewUnroutedHandler(tusd.Config{
				BasePath:                "/tus/",
				StoreComposer:           composer,
				RespectForwardedHeaders: true,
				NotifyCompleteUploads:   true,
				NotifyCreatedUploads:    true,
			})
			if err != nil {
				panic(fmt.Errorf("Unable to create handler: %s", err))
			}
			handler.DelFile(w, r)
			return
		}
		if err == sql.ErrNoRows {
			http.Error(w, "File does not exists", http.StatusNotFound)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	} else {
		http.Error(w, ok.Error(), http.StatusInternalServerError)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate user access to the file
	store := filestore.New(*dir)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	handler, err := tusd.NewUnroutedHandler(tusd.Config{
		BasePath:                "/files/",
		StoreComposer:           composer,
		RespectForwardedHeaders: true,
		NotifyCompleteUploads:   true,
	})
	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}
	handler.GetFile(w, r)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := fmt.Sprintf("%v", claims["user_id"])
	value := base64.StdEncoding.EncodeToString([]byte(userId))
	meta := r.Header.Get("Upload-Metadata")
	meta += ",user_id " + value
	r.Header.Set("Upload-Metadata", meta)

	store := filestore.New(*dir)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	handler, err := tusd.NewUnroutedHandler(tusd.Config{
		BasePath:                "/tus/",
		StoreComposer:           composer,
		RespectForwardedHeaders: true,
		NotifyCompleteUploads:   true,
		NotifyCreatedUploads:    true,
	})

	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}

	go func() {
		for {
			select {
			case event := <-handler.CreatedUploads:
				filename, ok := event.Upload.MetaData["filename"]
				if ok {
					fmt.Printf("Adding upload %s of user %s to db\n", filename, event.Upload.MetaData["user_id"])
					owner, ok := strconv.Atoi(event.Upload.MetaData["user_id"])
					if ok == nil {
						file := models.File{FileId: event.Upload.ID, Owner: uint(owner), Filename: filename, Size: event.Upload.Size, Completed: false}
						file.Create()
					} else {
						fmt.Fprintf(os.Stderr, "Could not add upload %s of user %s to db\n", filename, event.Upload.MetaData["user_id"])
					}
				} else {
					fmt.Fprintf(os.Stderr, "Could not retrieve filename for user %s\n", userId)
				}
			}
		}
	}()
	handler.PostFile(w, r)
}

func tusdHandler() http.Handler {
	store := filestore.New(*dir)

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/tus/",
		StoreComposer:           composer,
		RespectForwardedHeaders: true,
		NotifyCompleteUploads:   true,
	})

	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}

	go func() {
		for {
			select {
			case event := <-handler.CompleteUploads:
				fmt.Printf("Upload %s finished\n", event.Upload.ID)
				err := models.SetFileCompleted(event.Upload.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not set file with id %s as completed", event.Upload.ID)
				}
			}
		}
	}()

	return handler
}
