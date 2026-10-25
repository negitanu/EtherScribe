package pipeline

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

func TestReadableReportParity(t *testing.T) {
	for _, fixture := range []string{"fortigate-essence.conf", "paloalto-vsys.xml", "junos-ex-els.set", "fortigate-object-parity.conf", "paloalto-object-parity.xml", "fortigate-vdom-boundaries.conf", "paloalto-network-ops.xml"} {
		t.Run(fixture, func(t *testing.T) {
			for _, format := range []string{"html", "excel"} {
				res, err := ConvertFile(ConvertRequest{Path: testdata(fixture), Format: format, Preset: "all"})
				if err != nil {
					t.Fatal(err)
				}
				text := string(res.Bytes)
				if format == "excel" {
					zr, err := zip.NewReader(bytes.NewReader(res.Bytes), int64(len(res.Bytes)))
					if err != nil {
						t.Fatal(err)
					}
					var all strings.Builder
					for _, f := range zr.File {
						r, err := f.Open()
						if err != nil {
							t.Fatal(err)
						}
						b, err := io.ReadAll(r)
						r.Close()
						if err != nil {
							t.Fatal(err)
						}
						decoder := xml.NewDecoder(bytes.NewReader(b))
						for {
							_, err := decoder.Token()
							if err == io.EOF {
								break
							}
							if err != nil {
								t.Fatalf("%s: %v", f.Name, err)
							}
						}
						all.Write(b)
					}
					text = all.String()
					for _, marker := range []string{"state=\"frozen\"", "autoFilter", "wrapText=\"1\"", "解析範囲", "ルート", "STP"} {
						if !strings.Contains(text, marker) {
							t.Errorf("Excel missing %s", marker)
						}
					}
				} else {
					for _, marker := range []string{"<nav", "@media print", "解析範囲", "システム設定", "<thead>"} {
						if !strings.Contains(text, marker) {
							t.Errorf("HTML missing %s", marker)
						}
					}
				}
				for _, secret := range []string{"ENC ", "super-secret", "plaintext-password"} {
					if strings.Contains(text, secret) {
						t.Errorf("%s leaked secret marker", format)
					}
				}
				if fixture == "junos-ex-els.set" {
					for _, marker := range []string{"アクセス VLAN", "トランク VLAN", "LACP 有効", "コンテキスト"} {
						if !strings.Contains(text, marker) {
							t.Errorf("%s missing %s", format, marker)
						}
					}
				}
			}
		})
	}
}
func TestReportSectionsAndEscaping(t *testing.T) {
	cfg, err := ParseContent("config system global\n set hostname \"<script>alert(1)</script>\"\nend\n", "synthetic.conf")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Render(cfg, "html", []string{"device_info"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "<script>") || !strings.Contains(string(b), "&lt;script&gt;") {
		t.Fatal("HTML not escaped")
	}
	for _, format := range []string{"html", "excel"} {
		res, err := ConvertFile(ConvertRequest{Path: testdata("junos-ex-els.set"), Format: format, Sections: []string{"vlans"}})
		if err != nil {
			t.Fatal(err)
		}
		text := string(res.Bytes)
		if format == "excel" {
			zr, _ := zip.NewReader(bytes.NewReader(res.Bytes), int64(len(res.Bytes)))
			for _, f := range zr.File {
				if f.Name == "xl/workbook.xml" {
					r, _ := f.Open()
					b, _ := io.ReadAll(r)
					r.Close()
					text = string(b)
				}
			}
		}
		if strings.Contains(text, "機器概要") || strings.Contains(text, "ポリシー") {
			t.Errorf("%s includes unselected section", format)
		}
		if !strings.Contains(text, "VLAN") || !strings.Contains(text, "解析範囲") {
			t.Errorf("%s missing selected section/coverage", format)
		}
	}
}
