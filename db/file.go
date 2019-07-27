package db

import (
	"database/sql"
	"FileStoreServer/db/mysql"
	"log"
)

func FileUpload(fileHash string, fileName string, fileSize int64, 
					fileAddr string) bool {
	stmt, err := mysql.DBConn().Prepare("insert ignore into tbl_file(`file_sha1`, `file_name`, " +
				 						"`file_size`, `file_addr`, `status`) values (?,?,?,?,?)")
	if err != nil {
		log.Println("Failed to prepare statement, " + err.Error())
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(fileHash, fileName, fileSize, fileAddr, 1)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if rf, err := res.RowsAffected(); err == nil {
		if rf <= 0 {
			log.Printf("file with hash: %s has been uploaded before\n", fileHash)
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

func GetFileMeta(fileHash string) (*TableFile, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, " + 
					"file_addr from tbl_file where file_sha1=? and status=1")
	if err != nil {
		log.Println("Failed to prepare statement, " + err.Error())
		return nil, err
	}
	defer stmt.Close()
	
	tFile := TableFile{}
	err = stmt.QueryRow(fileHash).Scan(&tFile.FileHash, &tFile.FileName, 
										&tFile.FileSize, &tFile.FileAddr)
	if err != nil {
		log.Println(err)
		return &tFile, err
	}
	return &tFile, nil
}