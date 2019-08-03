package db

import (
	"FileStoreServer/db/mysql"

	"time"
	"log"
)

type UserFile struct {
	Username string
	FileHash string
	FileName string
	FileSize int64
	UploadAt string
	LastUpdate string
}

func AddUserFile(userName, fileHash, fileName string, fileSize int64) bool {
	stmt, err := mysql.DBConn().Prepare("insert into tbl_user_file(user_name, " + 
					"file_sha1, file_name, file_size, upload_at) values (?, ?, ?, ?, ?)")
	if err != nil {
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(userName, fileHash, fileName, fileSize, time.Now())
	if err != nil {
		return false
	}
	return true
}

func QueryUserFiles(userName string, limit int) ([]UserFile, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, " + 
					"upload_at, last_update from tbl_user_file where user_name=? limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userName, limit)
	if err != nil {
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		uFile := UserFile{}
		err = rows.Scan(&uFile.FileHash, &uFile.FileName, &uFile.FileSize, &uFile.UploadAt, 
					&uFile.LastUpdate)
		if err != nil {
			log.Printf("Scan rows file when query user files: %v\n", err)
			break
		}
		userFiles = append(userFiles, uFile)
	}
	return userFiles, nil
}