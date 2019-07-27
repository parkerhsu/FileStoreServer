package meta

import(
	"sort"
	"FileStoreServer/db"
)

// file information
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// add or update FileMeta
func UpdateFileMeta(fMeta FileMeta) {
	fileMetas[fMeta.FileSha1] = fMeta
}

func UpdateFileMetaDB(fMeta FileMeta) bool {
	return db.FileUpload(fMeta.FileSha1, fMeta.FileName, fMeta.FileSize, fMeta.Location)
}

// return FileMeta
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

func GetFileMetaDB(filesha1 string) (FileMeta, error) {
	tFile, err := db.GetFileMeta(filesha1)
	if err != nil {
		return FileMeta{}, err
	}
	
	fMeta := FileMeta{
		FileSha1: tFile.FileHash,
		FileName: tFile.FileName.String,
		FileSize: tFile.FileSize.Int64,
		Location: tFile.FileAddr.String,
	}
	return fMeta, nil
}

func GetLastFileMeta(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	sort.Sort(ByUploadTime(fMetaArray))

	return fMetaArray[0:count]
}

func DeleteFileMeta(fSha1 string) {
	delete(fileMetas, fSha1)
}