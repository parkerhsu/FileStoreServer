package handler

import (
	"FileStoreServer/db"
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
	"FileStoreServer/store/oss"
)

func IndexHanlder(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/user/signin", http.StatusFound)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// return index.html
		data, err := ioutil.ReadFile("./static/view/upload.html")
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

		// 写入OSS
		newFile.Seek(0, 0)
		ossPath := "oss/" + fileMeta.FileSha1
		err = oss.Bucket().PutObject(ossPath, newFile)
		if err != nil {
			log.Println(err)
			w.Write([]byte("Upload failed!"))
			return
		}
		//fileMeta.Location = ossPath


		meta.UpdateFileMeta(fileMeta)
		if ok := meta.UpdateFileMetaDB(fileMeta); !ok {
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}

		// 更新用户文件表
		r.ParseForm()
		userName := r.Form.Get("username")
		ok := db.AddUserFile(userName, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if !ok {
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}

		// http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload succeed")
}

func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileHash := r.Form["filehash"][0]
	fMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
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
	userName := r.Form.Get("username")
	//fileMetas := meta.GetLastFileMeta(limitCnt)
	userFiles, err := db.QueryUserFiles(userName, limitCnt)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	data, err := json.Marshal(userFiles)
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

func TryFastLoadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// 解析请求参数
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileName := r.Form.Get("filename")
	fileSize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 查询文件信息
	fileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 查询失败返回错误信息
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg: "FastLoad failed, please get to the normal upload url"}
		w.Write(resp.JSONBytes())
		return
	}

	// 将文件信息写入用户文件表
	suc := db.AddUserFile(userName, fileHash, fileName, int64(fileSize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")
	row, _ := db.GetFileMeta(fileHash)
	signedURL := oss.DownloadURL(row.FileAddr.String)
	w.Write([]byte(signedURL))
}