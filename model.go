package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Model はアプリケーション状態♠
type Model struct {
	currentDir      string
	allEntries      []FileEntry
	filteredEntries []FileEntry
	query           string
	cursor          int
	keymap          KeyMap
	config          Config
	width           int
	height          int
}

// NewModel は新しいモデルを作成♥
func NewModel(startDir string) (*Model, error) {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	// 設定読み込み♠
	config := LoadConfig()

	entries, err := ScanFiles(absDir, config)
	if err != nil {
		return nil, err
	}

	width, height := getTerminalSize()

	m := &Model{
		currentDir:      absDir,
		allEntries:      entries,
		filteredEntries: RankEntries(entries, ""),
		query:           "",
		cursor:          0,
		keymap:          DefaultKeyMap(),
		config:          config,
		width:           width,
		height:          height,
	}

	return m, nil
}

// updateFilter はクエリに基づいてフィルタ更新♧
func (m *Model) updateFilter() {
	m.filteredEntries = RankEntries(m.allEntries, m.query)
	if m.cursor >= len(m.filteredEntries) {
		m.cursor = max(0, len(m.filteredEntries)-1)
	}
}

// changeDirectory はディレクトリ変更♠
func (m *Model) changeDirectory(newDir string) error {
	absDir := filepath.Join(m.currentDir, newDir)
	entries, err := ScanFiles(absDir, m.config)
	if err != nil {
		return err
	}

	m.currentDir = absDir
	m.allEntries = entries
	m.query = ""
	m.cursor = 0
	m.updateFilter()
	return nil
}

// View は画面描画♥
func (m *Model) View() string {
	var b strings.Builder

	// ヘッダー: カレントディレクトリとクエリ♠
	b.WriteString(fmt.Sprintf("\033[1;36m%s\033[0m ", m.currentDir))
	b.WriteString(fmt.Sprintf("\033[2m[%d files]\033[0m\n", len(m.allEntries)))
	b.WriteString(fmt.Sprintf("> %s\033[K\n", m.query))
	b.WriteString(strings.Repeat("─", min(m.width, 80)) + "\n")

	// ファイルリスト表示♥
	for i, entry := range m.filteredEntries {
		cursor := "  "
		if i == m.cursor {
			cursor = "\033[1;33m>\033[0m "
		}

		icon := " "
		color := "\033[0m"
		if entry.IsDir {
			icon = "📁"
			color = "\033[1;34m" // ディレクトリは青♠
		} else {
			icon = "📄"
		}

		// パス表示: DirPath/Name形式♧
		displayPath := entry.Name
		if entry.DirPath != "." {
			displayPath = filepath.Join(entry.DirPath, entry.Name)
		}

		b.WriteString(fmt.Sprintf("%s %s %s%s\033[0m\n",
			cursor, icon, color, displayPath))
	}

	// フッター: 操作説明♥
	b.WriteString("\n")
	b.WriteString("\033[2m[j/k]移動 [Enter]選択 [q]終了\033[0m")

	return b.String()
}

// HandleInput は入力処理♠
func (m *Model) HandleInput(r rune) (bool, string, error) {
	switch {
	case r == m.keymap.Quit:
		return true, "", nil // 終了♧

	case r == m.keymap.Down:
		if m.cursor < len(m.filteredEntries)-1 {
			m.cursor++
		}

	case r == m.keymap.Up:
		if m.cursor > 0 {
			m.cursor--
		}

	case r == m.keymap.Enter:
		if len(m.filteredEntries) > 0 {
			selected := m.filteredEntries[m.cursor]
			if selected.IsDir {
				// ディレクトリドリルダウン♥
				return false, "", m.changeDirectory(selected.Path)
			}
			// ファイル選択: パスを返す♠
			fullPath := filepath.Join(m.currentDir, selected.Path)
			return true, fullPath, nil
		}

	case r == m.keymap.Backspace || r == m.keymap.DeleteQuery:
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.updateFilter()
		}

	default:
		// 通常文字: クエリに追加♧
		if r >= 32 && r < 127 {
			m.query += string(r)
			m.updateFilter()
		}
	}

	return false, "", nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getTerminalSize はターミナルサイズを取得♠
func getTerminalSize() (int, int) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 80, 24 // デフォルト値♧
	}
	return int(ws.Col), int(ws.Row)
}
