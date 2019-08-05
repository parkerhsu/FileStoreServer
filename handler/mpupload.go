package handler

import (
	"net/http"
	"strconv"
	"fmt"
	"time"
	"math"
	"os"
	"strings"
	"path"

	"FileStoreServer/db"
	"github.com/garyburd/redigo/redis"
	"FileStoreServer/util"
	rPool "FileStoreServer/cache/redis"
)

type MpUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
}

// 返回初始化信息
func InitialMpUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileSize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// get a redis connection
	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// make initial info
	upInfo := MpUploadInfo{
		FileHash: fileHash,
		FileSize: fileSize,
		UploadID: userName + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: 5 * 1024 *1024,
		ChunkCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
	}

	// store initial info into redis
	conn.Do("HSET", "MP_" + upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	conn.Do("HSET", "MP_" + upInfo.UploadID, "filehash", upInfo.FileHash)
	conn.Do("HSET", "MP_" + upInfo.UploadID, "filesize", upInfo.FileSize)

	// return initial info
	w.Write(util.NewRespMsg(0, "ok", upInfo).JSONBytes())
}

func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// parse form
	r.ParseForm()
	//userName := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	fPath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fPath), 0744)
	fd, err := os.Create(fPath)

	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	conn.Do("HSET", "MP_" + uploadID, "chkinx_" + chunkIndex, 1)

	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	upId := r.Form.Get("uploadid")
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileSize := r.Form.Get("filesize")
	fileName := r.Form.Get("filename")

	conn := rPool.RedisPool().Get()
	defer conn.Close()

	// 判断分块数是否相同
	data, err := redis.Values(conn.Do("HSET", "MP_" + upId))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _  = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount += 1
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	fSize, _ := strconv.Atoi(fileSize)
	// 更新到文件表
	db.FileUpload(fileHash, fileName, int64(fSize), "")
	// 更新到用户文件表
	db.AddUserFile(userName, fileHash, fileName, int64(fSize))

	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}