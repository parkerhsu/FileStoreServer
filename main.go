package main

import(
	"net/http"
	"FileStoreServer/handler"
	"log"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}