package main

// KeyMap はキーバインド設定
type KeyMap struct {
	Quit        rune
	Up          rune
	Down        rune
	Enter       rune
	Backspace   rune
	DeleteQuery rune
}

// DefaultKeyMap はデフォルトキーマップ
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:        0x04, // 終了
		Up:          0x10, // 上移動
		Down:        0x0e, // 下移動
		Enter:       '\r', // 選択/ディレクトリ移動
		Backspace:   '\b', // クエリ削除
		DeleteQuery: 0x7f, // DELキー
	}
}

// TODO: 将来的に設定ファイル（~/.config/fuzzy-filer/config.yaml）から読み込む
// これでユーザーが自由にカスタマイズできる
