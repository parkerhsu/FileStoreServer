package mq

import (
	cmn "FileStoreServer/common"
)


type TransferData struct {
	FileHash      string
	CurLocation   string
	DestLocation  string
	DestStoreType cmn.StoreType
}
