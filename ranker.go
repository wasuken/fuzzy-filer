package main

import (
	"path/filepath"
	"sort"
	"strings"
)

// ScoredEntry はスコア付きファイルエントリ
type ScoredEntry struct {
	Entry FileEntry
	Score int
}

// RankEntries はクエリに基づいてエントリをランク付け♠
func RankEntries(entries []FileEntry, query string) []FileEntry {
	if query == "" {
		return entries[:min(10, len(entries))]
	}

	query = strings.ToLower(query)
	var scored []ScoredEntry

	for _, entry := range entries {
		score := calculateScore(entry, query)
		if score > 0 {
			scored = append(scored, ScoredEntry{
				Entry: entry,
				Score: score,
			})
		}
	}

	// スコア降順でソート♥
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		// 同点ならディレクトリ優先、その後名前順♧
		if scored[i].Entry.IsDir != scored[j].Entry.IsDir {
			return scored[i].Entry.IsDir
		}
		return scored[i].Entry.Name < scored[j].Entry.Name
	})

	// 上位10件のみ返す♠
	result := make([]FileEntry, 0, min(10, len(scored)))
	for i := 0; i < min(10, len(scored)); i++ {
		result = append(result, scored[i].Entry)
	}

	return result
}

// calculateScore はマッチスコアを計算♥
func calculateScore(entry FileEntry, query string) int {
	nameLower := strings.ToLower(entry.Name)
	score := 0

	// ディレクトリ名完全マッチ: 最優先♠
	if entry.IsDir && nameLower == query {
		return 10000
	}

	// ベースファイル名の前方一致: 高得点♥
	if strings.HasPrefix(nameLower, query) {
		score += 1000
		if entry.IsDir {
			score += 500 // ディレクトリならさらにボーナス♧
		}
		return score
	}

	// ベースファイル名の部分一致♠
	if idx := strings.Index(nameLower, query); idx >= 0 {
		score += 500 - idx*10 // 前方に近いほど高得点♥
		if entry.IsDir {
			score += 200
		}
		return score
	}

	// 親ディレクトリ名マッチ（下層から）♧
	if entry.DirPath != "." {
		dirs := strings.Split(entry.DirPath, string(filepath.Separator))
		for i := len(dirs) - 1; i >= 0; i-- {
			dirLower := strings.ToLower(dirs[i])
			if strings.Contains(dirLower, query) {
				score += 100 - (len(dirs)-1-i)*20 // 下層ほど高得点♠
				break
			}
		}
	}

	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
