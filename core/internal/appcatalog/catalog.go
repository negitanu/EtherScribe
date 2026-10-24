package appcatalog

import (
	"embed"
	"encoding/csv"
	"io"
	"strings"
	"sync"
)

//go:embed appid.csv
var raw embed.FS

// Info is a resolved application id (WallScribe get_app_info subset).
type Info struct {
	AppName     string
	Category    string
	Risk        string
	Description string
}

var (
	once sync.Once
	byID map[string]Info
)

// Lookup returns catalog data for an application id.
// Unknown ids are not an error; callers keep the raw id as the name.
func Lookup(appID string) (Info, bool) {
	once.Do(load)
	info, ok := byID[strings.TrimSpace(appID)]
	return info, ok
}

func load() {
	byID = map[string]Info{}
	f, err := raw.Open("appid.csv")
	if err != nil {
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.TrimSpace(h)] = i
	}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		id := field(row, col, "app_id")
		if id == "" {
			continue
		}
		byID[id] = Info{
			AppName:     field(row, col, "app_name"),
			Category:    field(row, col, "category"),
			Risk:        field(row, col, "risk"),
			Description: field(row, col, "description"),
		}
	}
}

func field(row []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}
