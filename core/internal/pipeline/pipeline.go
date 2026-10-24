package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/negitanu/etherscribe/internal/diff"
	"github.com/negitanu/etherscribe/internal/export"
	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/ingest"
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/parse/fortigate"
	"github.com/negitanu/etherscribe/internal/parse/junos"
	"github.com/negitanu/etherscribe/internal/parse/paloalto"
)

type ConvertRequest struct {
	Path     string
	Format   string
	Preset   string
	Sections []string
	Output   string
}

type ConvertResult struct {
	Bytes      []byte
	Format     string
	Summary    map[string]any
	OutputPath string
}

func ConvertFile(req ConvertRequest) (*ConvertResult, error) {
	data, err := os.ReadFile(req.Path)
	if err != nil {
		return nil, err
	}
	in, err := ingest.ValidateBytes(filepath.Base(req.Path), data)
	if err != nil {
		return nil, err
	}
	cfg, err := ParseContent(in.Content, in.Filename)
	if err != nil {
		return nil, err
	}
	cfg.Encoding = in.Encoding
	sections, err := export.ResolveSections(req.Preset, req.Sections)
	if err != nil {
		return nil, err
	}
	format := strings.ToLower(req.Format)
	if format == "" {
		format = "json"
	}
	out, err := Render(cfg, format, sections)
	if err != nil {
		return nil, err
	}
	res := &ConvertResult{
		Bytes:  out,
		Format: format,
		Summary: map[string]any{
			"hostname":   cfg.DeviceInfo.Hostname,
			"model":      cfg.DeviceInfo.Model,
			"osVersion":  cfg.DeviceInfo.OSVersion,
			"deviceType": cfg.DeviceInfo.DeviceType,
			"interfaces": len(cfg.Interfaces),
			"vlans":      len(cfg.VLANs),
			"policies":   len(cfg.Policies),
			"routes":     len(cfg.Routes),
			"coverage":   len(cfg.Coverage),
			"encoding":   in.Encoding,
		},
	}
	if req.Output != "" {
		if err := os.WriteFile(req.Output, out, 0o644); err != nil {
			return nil, err
		}
		res.OutputPath = req.Output
	}
	return res, nil
}

func ParseContent(content, filename string) (*model.Config, error) {
	id := identify.Identify(content)
	if id == nil {
		return nil, fmt.Errorf("機器種別不明のエラー: 識別できませんでした")
	}
	switch id.Type {
	case identify.FortiGate:
		return fortigate.Parse(content, filename)
	case identify.PaloAlto:
		return paloalto.Parse(content, filename)
	case identify.JunosEX:
		return junos.ParseEX(content, filename)
	default:
		return nil, fmt.Errorf("機器種別不明のエラー")
	}
}

func Render(cfg *model.Config, format string, sections []string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		return export.RenderJSON(cfg, sections)
	case "html":
		return export.RenderHTML(cfg, sections)
	case "excel", "xlsx":
		return export.RenderExcel(cfg, sections)
	default:
		return nil, fmt.Errorf("未対応の出力形式です: %s", format)
	}
}

type CompareRequest struct {
	Before string
	After  string
	Format string
	Output string
}

func CompareFiles(req CompareRequest) (*ConvertResult, error) {
	beforeData, err := os.ReadFile(req.Before)
	if err != nil {
		return nil, err
	}
	afterData, err := os.ReadFile(req.After)
	if err != nil {
		return nil, err
	}
	bin, err := ingest.ValidateBytes(filepath.Base(req.Before), beforeData)
	if err != nil {
		return nil, err
	}
	ain, err := ingest.ValidateBytes(filepath.Base(req.After), afterData)
	if err != nil {
		return nil, err
	}
	bcfg, err := ParseContent(bin.Content, bin.Filename)
	if err != nil {
		return nil, err
	}
	acfg, err := ParseContent(ain.Content, ain.Filename)
	if err != nil {
		return nil, err
	}
	rep, err := diff.Compare(bcfg, acfg)
	if err != nil {
		return nil, err
	}
	format := strings.ToLower(req.Format)
	if format == "" {
		format = "json"
	}
	var out []byte
	switch format {
	case "json":
		out, err = diff.RenderJSON(rep)
	case "html":
		out, err = diff.RenderHTML(rep)
	default:
		return nil, fmt.Errorf("差分の未対応形式です: %s", format)
	}
	if err != nil {
		return nil, err
	}
	res := &ConvertResult{
		Bytes:  out,
		Format: format,
		Summary: map[string]any{
			"changes": len(rep.Changes),
		},
	}
	if req.Output != "" {
		if err := os.WriteFile(req.Output, out, 0o644); err != nil {
			return nil, err
		}
		res.OutputPath = req.Output
	}
	return res, nil
}
