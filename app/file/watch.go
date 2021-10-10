package file

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"time"
)

func GetFileList(latestFileCreatedTime *time.Time, watchDirPath string) (Files, error) {
	var fileList Files
	err := filepath.Walk(watchDirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if latestFileCreatedTime.Before(info.ModTime()) {
				file := NewFile(info)
				fileList = append(fileList, file)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get file list in %v: %v", watchDirPath, err)
	}
	// 新しい順に並び替え
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].CreatedTime.Unix() > fileList[j].CreatedTime.Unix()
	})
	for _, file := range fileList {
		fmt.Printf("%s\n", file.Name)
	}
	return fileList, nil
}
