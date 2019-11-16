package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
)

var dir = flag.String("dir", "tusd-files", "where the uploaded files should be stored")

func main() {
	flag.Parse()

	_, err := os.Stat(*dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(*dir, 0777)
	}

	store := filestore.New(*dir)

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/files/",
		StoreComposer:           composer,
		RespectForwardedHeaders: true,
	})
	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}

	go func() {
		for {
			event := <-handler.CompleteUploads
			fmt.Printf("Upload %s finished", event.Upload.ID)
		}
	}()

	http.Handle("/files/", http.StripPrefix("/files/", handler))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(fmt.Errorf("Unable to listen: %s", err))
	}
}
