package main

import (
	"os"
	"path/filepath"
	"strings"
)

// FileEntry はファイル/ディレクトリ情報
type FileEntry struct {
	Path    string
	Name    string
	IsDir   bool
	DirPath string // 親ディレクトリパス
}

// ScanFiles は指定ディレクトリ配下を走査する♠
func ScanFiles(rootDir string, config Config) ([]FileEntry, error) {
	var entries []FileEntry
	fileCount := 0

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // エラーは無視して継続♧
		}

		// rootDir自体はスキップ
		if path == rootDir {
			return nil
		}

		// 相対パス計算♥
		relPath, _ := filepath.Rel(rootDir, path)

		// 深度チェック♠
		depth := strings.Count(relPath, string(os.PathSeparator)) + 1
		if depth > config.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 除外パターンチェック♧
		if shouldExclude(relPath, config.ExcludePatterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 隠しファイル/ディレクトリはスキップ♥
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// ファイル数上限チェック♠
		fileCount++
		if fileCount > config.MaxFiles {
			return filepath.SkipAll // これ以上スキャンしない♧
		}

		dirPath := filepath.Dir(relPath)

		entries = append(entries, FileEntry{
			Path:    relPath,
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			DirPath: dirPath,
		})

		return nil
	})

	return entries, err
}
