package main

import (
	"gitlab.com/codmill/customer-projects/guardian/lobby-attendant/helpers"
	"os"
	"path"
)

type FileEntry struct {
	FileName   string               `json:"fileName"`
	ParentPath string               `json:"parent"`
	LeafLevel  int                  `json:"leafLevel"`
	IsDir      bool                 `json:"isDir"`
	Size       int64                `json:"size"`
	ItemType   helpers.BulkItemType `json:"mimeType"`
}

func NewFileEntry(dir string, leafLevel int, from os.FileInfo) FileEntry {
	fullPath := path.Join(dir, from.Name())

	mt := helpers.ItemTypeForFilepath(fullPath)

	return FileEntry{
		FileName:   from.Name(),
		IsDir:      from.IsDir(),
		Size:       from.Size(),
		ParentPath: dir,
		LeafLevel:  leafLevel,
		ItemType:   mt,
	}
}
