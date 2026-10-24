package diff

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/version"
)

const Schema = "etherscribe.config-diff"

type ChangeKind string

const (
	Added    ChangeKind = "added"
	Removed  ChangeKind = "removed"
	Modified ChangeKind = "modified"
)

type Change struct {
	Kind   ChangeKind     `json:"kind"`
	Entity string         `json:"entity"`
	Key    string         `json:"key"`
	Before map[string]any `json:"before,omitempty"`
	After  map[string]any `json:"after,omitempty"`
	Fields []string       `json:"fields,omitempty"`
}

type Report struct {
	Schema        string            `json:"schema"`
	SchemaVersion int               `json:"schemaVersion"`
	GeneratedAt   string            `json:"generatedAt"`
	Generator     map[string]string `json:"generator"`
	BeforeFile    string            `json:"beforeFile"`
	AfterFile     string            `json:"afterFile"`
	Changes       []Change          `json:"changes"`
}

func Compare(before, after *model.Config) (*Report, error) {
	if before.DeviceInfo.DeviceType != after.DeviceInfo.DeviceType ||
		before.DeviceInfo.DeviceType == identify.Unknown ||
		after.DeviceInfo.DeviceType == identify.Unknown {
		return nil, fmt.Errorf("機器種別が一致しないため比較できません")
	}
	rep := &Report{
		Schema:        Schema,
		SchemaVersion: 1,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Generator:     map[string]string{"name": version.Name, "version": version.Version},
		BeforeFile:    before.SourceFile,
		AfterFile:     after.SourceFile,
	}

	rep.Changes = append(rep.Changes, diffMaps("policy", indexPolicies(before.Policies), indexPolicies(after.Policies))...)
	rep.Changes = append(rep.Changes, diffMaps("interface", indexInterfaces(before.Interfaces), indexInterfaces(after.Interfaces))...)
	rep.Changes = append(rep.Changes, diffMaps("route", indexRoutes(before.Routes), indexRoutes(after.Routes))...)
	return rep, nil
}

func indexPolicies(items []model.Policy) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, p := range items {
		key := p.VDOM + "/" + p.PolicyID
		out[key] = map[string]any{
			"name": p.Name, "action": p.Action, "enabled": p.Enabled,
			"nat": p.NATEnabled, "srcintf": p.SourceInterface, "dstintf": p.DestinationInterface,
		}
	}
	return out
}

func indexInterfaces(items []model.Interface) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, i := range items {
		key := i.VDOM + "/" + i.Name
		out[key] = map[string]any{"ip": i.IPAddress, "type": i.InterfaceType, "status": i.Status}
	}
	return out
}

func indexRoutes(items []model.Route) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, r := range items {
		key := r.VDOM + "/" + r.Name + "/" + r.Destination
		out[key] = map[string]any{"gateway": r.Gateway, "interface": r.Interface, "enabled": r.Enabled}
	}
	return out
}

func diffMaps(entity string, before, after map[string]map[string]any) []Change {
	var changes []Change
	for k, bv := range before {
		av, ok := after[k]
		if !ok {
			changes = append(changes, Change{Kind: Removed, Entity: entity, Key: k, Before: bv})
			continue
		}
		var fields []string
		for fk, fv := range bv {
			if fmt.Sprint(av[fk]) != fmt.Sprint(fv) {
				fields = append(fields, fk)
			}
		}
		for fk := range av {
			if _, ok := bv[fk]; !ok {
				fields = append(fields, fk)
			}
		}
		if len(fields) > 0 {
			changes = append(changes, Change{Kind: Modified, Entity: entity, Key: k, Before: bv, After: av, Fields: fields})
		}
	}
	for k, av := range after {
		if _, ok := before[k]; !ok {
			changes = append(changes, Change{Kind: Added, Entity: entity, Key: k, After: av})
		}
	}
	return changes
}

func RenderJSON(rep *Report) ([]byte, error) {
	return json.MarshalIndent(rep, "", "  ")
}

func RenderHTML(rep *Report) ([]byte, error) {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="ja"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>EtherScribe 設定差分</title><style>body{font-family:system-ui,sans-serif;margin:32px;color:#243650;background:#f5f7fb}h1{color:#244b7a}table{border-collapse:collapse;width:100%;background:white;font-size:12px}th,td{padding:12px;border-bottom:1px solid #dce4ef;text-align:left;vertical-align:top;overflow-wrap:anywhere}th{background:#e7effb}pre{white-space:pre-wrap;max-width:420px;margin:0}.added{color:#16805b}.removed{color:#b34242}.modified{color:#996500}@media print{body{margin:0;background:white}tr{break-inside:avoid}thead{display:table-header-group}}</style></head><body><h1>設定差分レポート</h1>`)
	fmt.Fprintf(&b, "<p>%s → %s</p><p>変更 %d 件 · インターフェース / ルート / ポリシー</p>", html.EscapeString(rep.BeforeFile), html.EscapeString(rep.AfterFile), len(rep.Changes))
	b.WriteString(`<table><thead><tr><th>区分</th><th>対象</th><th>識別子</th><th>変更前</th><th>変更後</th></tr></thead><tbody>`)
	names := map[ChangeKind]string{Added: "追加", Removed: "削除", Modified: "変更"}
	for _, c := range rep.Changes {
		before, _ := json.MarshalIndent(c.Before, "", "  ")
		after, _ := json.MarshalIndent(c.After, "", "  ")
		kind := string(c.Kind)
		if _, ok := names[c.Kind]; !ok {
			kind = "unknown"
		}
		fmt.Fprintf(&b, `<tr><td class="%s">%s</td><td>%s</td><td>%s</td><td><pre>%s</pre></td><td><pre>%s</pre></td></tr>`, kind, html.EscapeString(names[c.Kind]), html.EscapeString(c.Entity), html.EscapeString(c.Key), html.EscapeString(string(before)), html.EscapeString(string(after)))
	}
	if len(rep.Changes) == 0 {
		b.WriteString(`<tr><td colspan="5">比較対象の項目に変更はありません。</td></tr>`)
	}
	b.WriteString(`</tbody></table><p>EtherScribe · ローカル生成</p></body></html>`)
	return []byte(b.String()), nil
}
