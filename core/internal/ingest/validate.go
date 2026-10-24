package ingest

import (
	"fmt"
	"path/filepath"
	"strings"
)

const MaxInputBytes = 64 << 20 // 64 MiB

var allowedExt = map[string]struct{}{
	".conf": {},
	".xml":  {},
	".set":  {},
	".txt":  {},
}

type ValidatedInput struct {
	Filename string
	Content  string
	Encoding string
	Size     int
}

func ValidateBytes(filename string, data []byte) (ValidatedInput, error) {
	if len(data) == 0 {
		return ValidatedInput{}, fmt.Errorf("空ファイルです")
	}
	if len(data) > MaxInputBytes {
		return ValidatedInput{}, fmt.Errorf("入力サイズが上限(%d bytes)を超えています", MaxInputBytes)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		if _, ok := allowedExt[ext]; !ok {
			return ValidatedInput{}, fmt.Errorf("未対応の拡張子です: %s", ext)
		}
	}
	text, enc, err := DetectEncoding(data)
	if err != nil {
		return ValidatedInput{}, err
	}
	if strings.TrimSpace(text) == "" {
		return ValidatedInput{}, fmt.Errorf("空ファイルです")
	}
	base := filepath.Base(filename)
	return ValidatedInput{Filename: base, Content: text, Encoding: enc, Size: len(data)}, nil
}
