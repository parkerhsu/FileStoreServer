package handler

import (
	"net/http"
	"io/ioutil"
	"io"
	"log"
	"os"	
	"time"
	"encoding/json"
	"strconv"

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

func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileHash := r.Form["filehash"][0]
	fMeta := meta.GetFileMeta(fileHash)
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	
	fileMetas := meta.GetLastFileMeta(limitCnt)
	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fMeta := meta.GetFileMeta(r.Form.Get("filehash"))

	f, err := os.Open(fMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\"" + fMeta.FileName + "\"")
	w.Write(data)
}

func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	
	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	fileMeta := meta.GetFileMeta(fSha1)
	fileMeta.FileName = newFileName
	meta.UpdateFileMeta(fileMeta)

	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fSha1 := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fSha1)
	os.Remove(fileMeta.Location)
	meta.DeleteFileMeta(fSha1)

	w.WriteHeader(http.StatusOK)
}