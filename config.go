package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config はアプリケーション設定
type Config struct {
	ExcludePatterns []string `json:"exclude_patterns"` // 除外パターン
	MaxDepth        int      `json:"max_depth"`        // 最大探索深度
	MaxFiles        int      `json:"max_files"`        // 最大ファイル数
	EnablePreview   bool     `json:"enable_preview"`   // プレビュー有効化
	PreviewLines    int      `json:"preview_lines"`    // プレビュー行数
}

// DefaultConfig はデフォルト設定
func DefaultConfig() Config {
	return Config{
		ExcludePatterns: []string{
			"node_modules",
			".git",
			".svn",
			".hg",
			"target", // Rust/Java
			"dist",
			"build",
			"*.log",
			".DS_Store",
			"__pycache__",
			".pytest_cache",
			".venv",
			"venv",
		},
		MaxDepth:      10,     // 10階層まで
		MaxFiles:      100000, // 10万ファイルまで
		EnablePreview: true,
		PreviewLines:  20,
	}
}

// LoadConfig は設定ファイルを読み込む
func LoadConfig() Config {
	configPath := getConfigPath()

	// 設定ファイルが存在しない場合はデフォルト値を使用
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig()
	}

	return config
}

// SaveConfig は設定をファイルに保存
func SaveConfig(config Config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// ディレクトリ作成
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath は設定ファイルパスを取得
func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".fuzzy-filer.json" // フォールバック
	}
	return filepath.Join(home, ".config", "fuzzy-filer", "config.json")
}

// shouldExclude はパスが除外対象かチェック
func shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// ワイルドカード対応（簡易版）
		if strings.HasPrefix(pattern, "*.") {
			// 拡張子マッチ
			ext := pattern[1:] // "*.log" -> ".log"
			if strings.HasSuffix(path, ext) {
				return true
			}
		} else {
			// 部分文字列マッチ
			if strings.Contains(path, pattern) {
				return true
			}
		}
	}
	return false
}
