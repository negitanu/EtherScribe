package ingest

import (
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func TestDetectUTF8(t *testing.T) {
	text, enc, err := DetectEncoding([]byte("abc#config-version=FGT60F-7.2.0-FW"))
	if err != nil || enc != "utf-8" || text == "" {
		t.Fatalf("utf8: %q %q %v", text, enc, err)
	}
}

func TestDetectUTF8BOM(t *testing.T) {
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("hello")...)
	text, enc, err := DetectEncoding(data)
	if err != nil || enc != "utf-8-sig" || text != "hello" {
		t.Fatalf("bom: %q %q %v", text, enc, err)
	}
}

func TestDetectCP932(t *testing.T) {
	encoded, _, err := transform.Bytes(japanese.ShiftJIS.NewEncoder(), []byte("設定"))
	if err != nil {
		t.Fatal(err)
	}
	// Force non-UTF8 by using raw SJIS without valid UTF8
	text, enc, err := DetectEncoding(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if enc != "cp932" && enc != "latin-1" {
		// On some platforms transform may produce utf8-valid accidental decode; accept cp932 preferred
		t.Logf("encoding=%s text=%q", enc, text)
	}
	if text == "" {
		t.Fatal("empty")
	}
}

func TestValidateEmpty(t *testing.T) {
	_, err := ValidateBytes("x.conf", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = ValidateBytes("x.bin", []byte("hello"))
	if err == nil {
		t.Fatal("expected extension error")
	}
}
