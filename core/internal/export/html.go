package export

import (
	"fmt"
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/redaction"
	"github.com/negitanu/etherscribe/internal/version"
	"html"
	"strings"
	"time"
)

func RenderHTML(cfg *model.Config, sections []string) ([]byte, error) {
	tables, err := reportTables(cfg, sections)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="ja"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>EtherScribe パラメータシート</title><style>
:root{font-family:-apple-system,BlinkMacSystemFont,"Hiragino Sans",sans-serif;color:#243650;background:#f3f6fa;font-size:13px}*{box-sizing:border-box}body{margin:0}header{background:#14253d;color:white;padding:40px max(24px,calc((100% - 1280px)/2))}header small{color:#9bb5da;letter-spacing:2px}h1{font-size:28px;margin:12px 0}header p{color:#b1c3dc;font-size:11px}.layout{max-width:1360px;margin:auto;display:grid;grid-template-columns:190px minmax(0,1fr);gap:28px;padding:28px}nav{align-self:start;position:sticky;top:20px}nav a{display:block;text-decoration:none;color:#536e91;padding:8px;font-size:12px}nav a:hover{color:#2369d6;background:#e8effa}main{min-width:0}section{background:white;border:1px solid #dce4ef;border-radius:8px;padding:24px;margin-bottom:22px;scroll-margin-top:20px}h2{font-size:18px;margin:0 0 20px;color:#284d80}.table-wrap{overflow:auto}table{border-collapse:collapse;width:100%;font-size:11px}th,td{padding:11px 12px;border-bottom:1px solid #e3e9f2;text-align:left;vertical-align:top;white-space:pre-wrap;overflow-wrap:anywhere;min-width:100px;max-width:420px}th{background:#edf3fc;color:#49658b;font-weight:600}tr:nth-child(even) td{background:#f8fafd}.notice{font-size:11px;line-height:1.9;color:#788ca9;margin:0 0 20px}footer{padding:24px;text-align:center;color:#7a8fa9;font-size:10px}@media(max-width:760px){.layout{display:block;padding:16px}nav{position:static;display:flex;flex-wrap:wrap;margin-bottom:18px}section{padding:16px}}@media print{@page{size:A4 landscape;margin:12mm}:root{background:white;font-size:10px}header{background:white;color:#14253d;padding:0 0 16px;border-bottom:2px solid #284d80}header p,header small{color:#4a6487}.layout{display:block;padding:16px 0}nav{display:none}section{border:0;padding:0;break-before:auto}h2{break-after:avoid}tr{break-inside:avoid}thead{display:table-header-group}.table-wrap{overflow:visible}th,td{min-width:0;max-width:none;padding:5px;font-size:8px}footer{padding:8px}}
</style></head><body><header><small>ETHERSCRIBE / PARAMETER SHEET</small>`)
	fmt.Fprintf(&b, "<h1>%s</h1><p>生成 %s · %s %s · 未対応・解析範囲 %d 件 · 解析警告 %d 件</p></header>", esc(redaction.ScrubSecrets(cfg.DeviceInfo.Hostname)+" パラメータシート"), esc(time.Now().Format("2006-01-02 15:04 -07:00")), version.Name, version.Version, len(cfg.Coverage), len(cfg.ParseErrors))
	b.WriteString(`<div class="layout"><nav aria-label="レポートの目次">`)
	for i, t := range tables {
		fmt.Fprintf(&b, `<a href="#section-%d">%02d · %s</a>`, i, i+1, esc(t.Name))
	}
	b.WriteString(`</nav><main><p class="notice">設定ファイルに記載された情報を出力しています。秘密値は伏せています。空欄は未設定または取得対象外です。未反映の設定は解析範囲を確認してください。</p>`)
	for i, t := range tables {
		fmt.Fprintf(&b, `<section id="section-%d"><h2>%02d / %s</h2><div class="table-wrap"><table><thead><tr>`, i, i+1, esc(t.Name))
		for _, cell := range t.Rows[0] {
			fmt.Fprintf(&b, "<th scope=\"col\">%s</th>", esc(cell))
		}
		b.WriteString("</tr></thead><tbody>")
		for _, row := range t.Rows[1:] {
			b.WriteString("<tr>")
			for _, cell := range row {
				if cell == "" {
					cell = "—"
				}
				fmt.Fprintf(&b, "<td>%s</td>", esc(cell))
			}
			b.WriteString("</tr>")
		}
		b.WriteString("</tbody></table></div></section>")
	}
	b.WriteString("</main></div><footer>EtherScribe · Network documentation · ローカル生成</footer></body></html>")
	return []byte(b.String()), nil
}
func esc(s string) string { return html.EscapeString(s) }
