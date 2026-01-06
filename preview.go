package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	maxPreviewSize = 1 * 1024 * 1024 // 1MB
)

// GeneratePreview はファイル/ディレクトリのプレビューを生成
func GeneratePreview(path string, maxLines int) []string {
	info, err := os.Stat(path)
	if err != nil {
		return []string{"Error: " + err.Error()}
	}

	if info.IsDir() {
		return previewDirectory(path, maxLines)
	}

	return previewFile(path, maxLines, info.Size())
}

// previewFile はファイル内容のプレビュー
func previewFile(path string, maxLines int, size int64) []string {
	// サイズチェック
	if size > maxPreviewSize {
		return []string{
			fmt.Sprintf("File too large: %.2f MB", float64(size)/(1024*1024)),
			"(Preview disabled for files > 1MB)",
		}
	}

	// バイナリファイルチェック
	if isBinaryFile(path) {
		return []string{"Binary file"}
	}

	file, err := os.Open(path)
	if err != nil {
		return []string{"Error: " + err.Error()}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() && lineNum < maxLines {
		line := scanner.Text()

		// タブを4スペースに変換
		line = strings.ReplaceAll(line, "\t", "    ")

		// 長すぎる行は切り詰め
		if len(line) > 80 {
			line = line[:80] + "..."
		}

		lines = append(lines, line)
		lineNum++
	}

	if scanner.Err() != nil {
		lines = append(lines, "Error reading file: "+scanner.Err().Error())
	}

	if lineNum == maxLines {
		lines = append(lines, "...")
	}

	if len(lines) == 0 {
		return []string{"(Empty file)"}
	}

	return lines
}

// previewDirectory はディレクトリ内容のプレビュー
func previewDirectory(path string, maxLines int) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return []string{"Error: " + err.Error()}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Directory: %d items", len(entries)))
	lines = append(lines, "")

	count := 0
	for _, entry := range entries {
		if count >= maxLines-2 {
			lines = append(lines, fmt.Sprintf("... and %d more", len(entries)-count))
			break
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}

		lines = append(lines, fmt.Sprintf("%s %s", icon, entry.Name()))
		count++
	}

	return lines
}

// isBinaryFile はバイナリファイルかどうか判定
func isBinaryFile(path string) bool {
	// 拡張子でチェック
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".zip": true, ".tar": true, ".gz": true, ".bz2": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true,
		".o": true, ".a": true, ".pyc": true,
	}

	if binaryExts[ext] {
		return true
	}

	// ファイル先頭をチェック
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	// UTF-8として有効かチェック
	return !utf8.Valid(buf[:n])
}
