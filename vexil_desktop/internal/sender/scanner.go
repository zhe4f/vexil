package sender

import (
	"os"
	"path/filepath"
	"sort"
)

type FileInfo struct {
	Path    string
	RelPath string
	Size    int64
}

type ScanResult struct {
	Files []FileInfo
}

func ScanFiles(paths []string) (*ScanResult, error) {
	var allFiles []FileInfo

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			baseDir := absPath
			filepath.Walk(absPath, func(walkPath string, walkInfo os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if !walkInfo.IsDir() {
					relPath, _ := filepath.Rel(baseDir, walkPath)
					allFiles = append(allFiles, FileInfo{
						Path:    walkPath,
						RelPath: relPath,
						Size:    walkInfo.Size(),
					})
				}
				return nil
			})
		} else {
			relPath := filepath.Base(absPath)
			allFiles = append(allFiles, FileInfo{
				Path:    absPath,
				RelPath: relPath,
				Size:    info.Size(),
			})
		}
	}

	if len(allFiles) == 0 {
		return nil, os.ErrNotExist
	}

	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].RelPath < allFiles[j].RelPath
	})

	return &ScanResult{Files: allFiles}, nil
}

func (s *ScanResult) TotalSize() int64 {
	var total int64
	for _, f := range s.Files {
		total += f.Size
	}
	return total
}

func (s *ScanResult) TotalUnits() int {
	return len(s.Files)
}