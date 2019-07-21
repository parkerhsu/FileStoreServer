package meta

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

// return FileMeta
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}