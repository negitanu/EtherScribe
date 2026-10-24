package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/negitanu/etherscribe/internal/export"
	"github.com/negitanu/etherscribe/internal/pipeline"
	"github.com/negitanu/etherscribe/internal/version"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type Server struct {
	in     io.Reader
	out    io.Writer
	mu     sync.Mutex
	cancel atomic.Pointer[context.CancelFunc]
}

func New(in io.Reader, out io.Writer) *Server {
	return &Server{in: in, out: out}
}

func (s *Server) Run() error {
	sc := bufio.NewScanner(s.in)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.write(Response{JSONRPC: "2.0", Error: &Error{Code: -32700, Message: "parse error"}})
			continue
		}
		s.handle(req)
	}
	return sc.Err()
}

func (s *Server) handle(req Request) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel.Store(&cancel)
	defer cancel()

	var result any
	var rpcErr *Error
	switch req.Method {
	case "engine.version":
		result = version.Current()
	case "sections.list":
		presets := make([]map[string]any, 0, len(export.Presets))
		for _, p := range export.Presets {
			presets = append(presets, map[string]any{
				"key": p.Key, "label": p.Label, "description": p.Description, "sections": p.Sections,
			})
		}
		result = map[string]any{"sections": export.SectionKeys, "presets": presets}
	case "cancel":
		if c := s.cancel.Load(); c != nil {
			(*c)()
		}
		result = map[string]any{"cancelled": true}
	case "convert.run":
		result, rpcErr = s.convert(ctx, req.Params)
	case "compare.run":
		result, rpcErr = s.compare(ctx, req.Params)
	default:
		rpcErr = &Error{Code: -32601, Message: "method not found: " + req.Method}
	}
	if req.ID == nil {
		return
	}
	s.write(Response{JSONRPC: "2.0", ID: req.ID, Result: result, Error: rpcErr})
}

type convertParams struct {
	Path     string   `json:"path"`
	Format   string   `json:"format"`
	Preset   string   `json:"preset"`
	Sections []string `json:"sections"`
	Output   string   `json:"output"`
}

func (s *Server) convert(ctx context.Context, raw json.RawMessage) (any, *Error) {
	var p convertParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, &Error{Code: -32602, Message: "invalid params"}
	}
	s.notify("progress", map[string]any{"phase": "start", "path": filepath.Base(p.Path)})
	if ctx.Err() != nil {
		return nil, &Error{Code: -32000, Message: "cancelled"}
	}
	out := p.Output
	if out == "" {
		tmp, err := os.CreateTemp("", "etherscribe-*.out")
		if err != nil {
			return nil, &Error{Code: -32001, Message: err.Error()}
		}
		out = tmp.Name()
		_ = tmp.Close()
	}
	res, err := pipeline.ConvertFile(pipeline.ConvertRequest{
		Path: p.Path, Format: p.Format, Preset: p.Preset, Sections: p.Sections, Output: out,
	})
	if err != nil {
		return nil, &Error{Code: -32002, Message: err.Error()}
	}
	s.notify("progress", map[string]any{"phase": "done"})
	return map[string]any{
		"outputPath": res.OutputPath,
		"format":     res.Format,
		"summary":    res.Summary,
	}, nil
}

type compareParams struct {
	Before string `json:"before"`
	After  string `json:"after"`
	Format string `json:"format"`
	Output string `json:"output"`
}

func (s *Server) compare(ctx context.Context, raw json.RawMessage) (any, *Error) {
	var p compareParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, &Error{Code: -32602, Message: "invalid params"}
	}
	if ctx.Err() != nil {
		return nil, &Error{Code: -32000, Message: "cancelled"}
	}
	out := p.Output
	if out == "" {
		tmp, err := os.CreateTemp("", "etherscribe-diff-*.out")
		if err != nil {
			return nil, &Error{Code: -32001, Message: err.Error()}
		}
		out = tmp.Name()
		_ = tmp.Close()
	}
	res, err := pipeline.CompareFiles(pipeline.CompareRequest{
		Before: p.Before, After: p.After, Format: p.Format, Output: out,
	})
	if err != nil {
		return nil, &Error{Code: -32002, Message: err.Error()}
	}
	return map[string]any{
		"outputPath": res.OutputPath,
		"format":     res.Format,
		"summary":    res.Summary,
	}, nil
}

func (s *Server) notify(method string, params any) {
	s.write(Notification{JSONRPC: "2.0", Method: method, Params: params})
}

func (s *Server) write(v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rpc marshal error: %v\n", err)
		return
	}
	s.out.Write(append(b, '\n'))
}
