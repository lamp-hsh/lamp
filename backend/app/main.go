package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"example.com/m/auth"
	"example.com/m/storage"
	"example.com/m/media"
	"example.com/m/media/comic"
	"example.com/m/media/video"
	"example.com/m/db"
)

func main() {
	fmt.Printf("Server Start\n")
	db.TokenDBReset()
	router := mux.NewRouter()

	router.HandleFunc("/api/auth/sign", auth.Sign)
	router.HandleFunc("/api/auth/registration", auth.Registration)

	router.HandleFunc("/api/storage", storage.Storage)
	router.PathPrefix("/storage").Handler(http.HandlerFunc(storage.StorageResource))

	router.HandleFunc("/api/media", media.Media)
	router.HandleFunc("/api/media/comic", comic.Comic)
	router.PathPrefix("/media/comic").Handler(http.HandlerFunc(comic.ComicResource))
	router.HandleFunc("/api/media/video", video.Video)
	router.PathPrefix("/media/video").Handler(http.HandlerFunc(video.VideoResource))

	log.Fatal(http.ListenAndServe(":4000", router))
}