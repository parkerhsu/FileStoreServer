package handler

import (
	"net/http"
	"io/ioutil"
	"io"
	"log"
	"os"	
	"time"

	"FileStoreServer/meta"
	"FileStoreServer/util"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// return index.html
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			log.Println(err)
			io.WriteString(w, "internal server error")	
			return 
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// receive file stream and store into local directory
		file, head, err := r.FormFile("file")
		if err != nil {
			log.Printf("Failed to get data: %v\n", err)
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/"+head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			log.Printf("Failed to create file: %v\n", err)
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			log.Printf("Failed to save data into file: %v\n", err)
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMeta(fileMeta)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload succeed")
}