package ingest

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// DetectEncoding ports WallScribe parsers.base.detect_encoding:
// try utf-8-sig, utf-8, cp932, latin-1 in order.
func DetectEncoding(data []byte) (text string, encoding string, err error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("空の入力です")
	}

	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		rest := data[3:]
		if utf8.Valid(rest) {
			return string(rest), "utf-8-sig", nil
		}
	}

	if utf8.Valid(data) {
		return string(data), "utf-8", nil
	}

	if decoded, _, decErr := transform.Bytes(japanese.ShiftJIS.NewDecoder(), data); decErr == nil && utf8.Valid(decoded) {
		return string(decoded), "cp932", nil
	}

	buf := make([]rune, len(data))
	for i, b := range data {
		buf[i] = rune(b)
	}
	return string(buf), "latin-1", nil
}
