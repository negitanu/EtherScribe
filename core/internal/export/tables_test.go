package export

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/parse/fortigate"
	"github.com/negitanu/etherscribe/internal/parse/paloalto"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func tableNamed(t *testing.T, tables []sheetData, name string) sheetData {
	t.Helper()
	for _, table := range tables {
		if table.Name == name {
			return table
		}
	}
	t.Fatalf("missing table %s", name)
	return sheetData{}
}
func TestAddressObjectRowsMatchWallScribe(t *testing.T) {
	b, err := os.ReadFile("../../testdata/fortigate-object-parity.conf")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := fortigate.Parse(string(b), "fixture.conf")
	if err != nil {
		t.Fatal(err)
	}
	tables, err := reportTables(cfg, []string{"objects"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 5 {
		t.Fatalf("expected four inventories and coverage; got %d", len(tables))
	}
	addresses := tableNamed(t, tables, "アドレスオブジェクト")
	wantHeader := []string{"オブジェクト名", "タイプ", "値", "説明", "タグ", "VDOM / vsys"}
	if !reflect.DeepEqual(addresses.Rows[0], wantHeader) {
		t.Fatalf("header=%q", addresses.Rows[0])
	}
	if len(addresses.Rows) != 8 {
		t.Fatalf("rows=%d", len(addresses.Rows))
	}
	want := []string{`拠点 "A" LAN`, "subnet", "192.0.2.0/24", "更新後の説明 <script>alert(1)</script>", "業務 用, 重要", "営業 部"}
	found := false
	for _, row := range addresses.Rows[1:] {
		if reflect.DeepEqual(row, want) {
			found = true
		}
	}
	if !found {
		t.Fatalf("address name/type/value separated or lost: %q", addresses.Rows)
	}
	group := tableNamed(t, tables, "アドレスグループ")
	if group.Rows[0][0] != "グループ名" || group.Rows[0][1] != "メンバー" {
		t.Fatalf("group headers=%q", group.Rows[0])
	}
	for _, table := range tables {
		for _, row := range table.Rows {
			if len(row) != len(table.Rows[0]) {
				t.Fatalf("ragged rows in %s", table.Name)
			}
		}
	}
	html, err := RenderHTML(cfg, []string{"objects"})
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"addresses[", "services[", "addressGroups[", "<script>", "<nil>", "map["} {
		if strings.Contains(string(html), bad) {
			t.Errorf("internal path/raw value leaked: %s", bad)
		}
	}
	if !strings.Contains(string(html), "&lt;script&gt;") {
		t.Fatal("description not escaped")
	}
	// Record order cannot change the table schema or object ordering.
	cfg.Objects.Addresses[0], cfg.Objects.Addresses[6] = cfg.Objects.Addresses[6], cfg.Objects.Addresses[0]
	again, _ := reportTables(cfg, []string{"objects"})
	if !reflect.DeepEqual(addresses, tableNamed(t, again, "アドレスオブジェクト")) {
		t.Fatal("unstable table")
	}
	// Decode the actual XLSX sheet and compare its complete header/data grid.
	workbook, err := RenderExcel(cfg, []string{"objects"})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(workbook), int64(len(workbook)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if f.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		r, _ := f.Open()
		raw, _ := io.ReadAll(r)
		r.Close()
		var sheet struct {
			Rows []struct {
				Cells []struct {
					Value string `xml:"is>t"`
				} `xml:"c"`
			} `xml:"sheetData>row"`
		}
		if err := xml.Unmarshal(raw, &sheet); err != nil {
			t.Fatal(err)
		}
		var rows [][]string
		for _, r := range sheet.Rows {
			var row []string
			for _, c := range r.Cells {
				row = append(row, c.Value)
			}
			rows = append(rows, row)
		}
		if !reflect.DeepEqual(rows, addresses.Rows) {
			t.Fatalf("Excel differs from semantic table: %q", rows)
		}
	}
}
func TestSharedDynamicGroupsInReports(t *testing.T) {
	b, err := os.ReadFile("../../testdata/paloalto-object-parity.xml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := paloalto.Parse(string(b), "fixture.xml")
	if err != nil {
		t.Fatal(err)
	}
	tables, err := reportTables(cfg, []string{"objects"})
	if err != nil {
		t.Fatal(err)
	}
	group := tableNamed(t, tables, "アドレスグループ")
	expected := []string{"Dynamic Group", "", "動的グループの説明", "'production' and 'web'", "自動", "vsys1"}
	found := false
	for _, row := range group.Rows[1:] {
		if reflect.DeepEqual(row, expected) {
			found = true
		}
	}
	if !found {
		t.Fatalf("dynamic group columns=%q", group.Rows)
	}
	a := tableNamed(t, tables, "アドレスオブジェクト")
	if len(a.Rows) != 7 {
		t.Fatal("shared or same-name vsys address lost")
	}
}
func TestNestedRecordsKeepParentAndContext(t *testing.T) {
	cfg := &model.Config{Routing: model.RoutingSettings{BGP: []model.BGPSettings{
		{ASNumber: "65001", RouterID: "192.0.2.1", Context: "VRF-A", Neighbors: []model.BGPNeighbor{{IP: "203.0.113.1", RemoteAS: "65010"}}},
		{ASNumber: "65002", RouterID: "192.0.2.2", Context: "VRF-B", Neighbors: []model.BGPNeighbor{{IP: "203.0.113.1", RemoteAS: "65020"}}},
	}}, SecurityProfilesDetail: model.SecurityProfilesDetail{AppControl: []model.AppControlProfile{{Name: "Profile A", VDOM: "root", Applications: []model.AppControlEntry{{AppID: "900001", AppName: "Synthetic App"}}}}}}
	tables, err := reportTables(cfg, []string{"network", "security_profiles"})
	if err != nil {
		t.Fatal(err)
	}
	n := tableNamed(t, tables, "BGP・ネイバー")
	if !reflect.DeepEqual(n.Rows[0][:5], []string{"親設定", "親 VDOM / vsys", "親コンテキスト", "IP アドレス", "対向 AS"}) {
		t.Fatalf("headers=%q", n.Rows[0])
	}
	for i, want := range []string{"VRF-A", "VRF-B"} {
		if n.Rows[i+1][2] != want || !strings.Contains(n.Rows[i+1][0], []string{"65001", "65002"}[i]) {
			t.Fatalf("parents=%q", n.Rows)
		}
	}
	app := tableNamed(t, tables, "アプリケーション制御・アプリケーション")
	if !strings.Contains(app.Rows[1][0], "Profile A") || app.Rows[1][1] != "root" || app.Rows[1][3] != "900001" {
		t.Fatalf("app=%q", app.Rows)
	}
}
func TestEveryModelFieldHasHumanLabel(t *testing.T) {
	seen := map[reflect.Type]bool{}
	var visit func(reflect.Type)
	visit = func(typ reflect.Type) {
		typ = recordType(typ)
		if typ.Kind() != reflect.Struct || seen[typ] {
			return
		}
		seen[typ] = true
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			key := jsonKey(f)
			if _, ok := reportLabels[key]; !ok {
				t.Errorf("missing label: %s.%s (%s)", typ.Name(), f.Name, key)
			}
			visit(f.Type)
		}
	}
	visit(reflect.TypeOf(model.Config{}))
}
func TestSafeWorkbookNamesAndRowHeight(t *testing.T) {
	tables := []sheetData{{Name: strings.Repeat("長い名前", 12) + "[]:*?/\\", Rows: [][]string{{"項目", "説明"}, {"名前", strings.Repeat("日本語の長い説明", 40)}}}, {Name: strings.Repeat("長い名前", 12) + "[]:*?/\\"}, {Name: "TEST"}, {Name: "test"}}
	safe := safeSheetNames(tables)
	seen := map[string]bool{}
	for _, table := range safe {
		if len([]rune(table.Name)) > 31 || strings.ContainsAny(table.Name, "[]:*?/\\") || seen[strings.ToLower(table.Name)] {
			t.Fatalf("invalid sheet name %q", table.Name)
		}
		seen[strings.ToLower(table.Name)] = true
	}
	var sheet struct {
		Rows []struct {
			Height float64 `xml:"ht,attr"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal([]byte(sheetXML(tables[0].Rows)), &sheet); err != nil {
		t.Fatal(err)
	}
	if len(sheet.Rows) != 2 || sheet.Rows[1].Height < 200 || sheet.Rows[1].Height > 409 {
		t.Fatalf("long Japanese cells need sufficient valid row height: %+v", sheet.Rows)
	}
}

// A populated model exercises every nested branch, so a new renderer cannot
// silently omit data while merely producing valid HTML/XML.
func TestSemanticTablesPreserveEverySelectedModelValue(t *testing.T) {
	cfg := &model.Config{}
	next := 0
	var populate func(reflect.Value)
	populate = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				populate(v.Field(i))
			}
		case reflect.Pointer:
			v.Set(reflect.New(v.Type().Elem()))
			populate(v.Elem())
		case reflect.Slice:
			v.Set(reflect.MakeSlice(v.Type(), 1, 1))
			populate(v.Index(0))
		case reflect.String:
			next++
			v.SetString(fmt.Sprintf("fixture-value-%04d", next))
		case reflect.Bool:
			v.SetBool(true)
		}
	}
	populate(reflect.ValueOf(cfg).Elem())
	data, err := RenderJSON(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	tables, err := reportTables(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	for _, table := range tables {
		for _, row := range table.Rows {
			for _, cell := range row {
				output.WriteString(cell)
				output.WriteByte('\n')
			}
		}
	}
	var verify func(any)
	verify = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			for _, child := range v {
				verify(child)
			}
		case []any:
			for _, child := range v {
				verify(child)
			}
		case string:
			if strings.HasPrefix(v, "fixture-value-") && !strings.Contains(output.String(), v) {
				t.Errorf("selected field lost: %s", v)
			}
		}
	}
	for _, key := range reportOrder {
		verify(doc[key])
	}
	if strings.Contains(output.String(), cfg.System.AdminUsers[0].Password) {
		t.Fatal("password leaked")
	}
	if !strings.Contains(output.String(), "秘匿済み") {
		t.Fatal("missing redaction indicator")
	}
}

func TestOmittedFalseFlagsRemainDisabled(t *testing.T) {
	cfg := &model.Config{NATPolicies: []model.NATPolicy{{Name: "SNAT without port forwarding", PortForward: false}}}
	tables, err := reportTables(cfg, []string{"nat"})
	if err != nil {
		t.Fatal(err)
	}
	nat := tableNamed(t, tables, "NAT")
	for i, label := range nat.Rows[0] {
		if label == "ポート転送" {
			if nat.Rows[1][i] != "無効" {
				t.Fatalf("false flag lost: %q", nat.Rows[1])
			}
			return
		}
	}
	t.Fatal("port forwarding column missing")
}

func TestRouteMetricAndDistanceAreSeparateColumns(t *testing.T) {
	cfg := &model.Config{Routes: []model.Route{{Name: "LAB-ROUTE", Destination: "203.0.113.0/24", Distance: "10", Metric: "25"}}}
	tables, err := reportTables(cfg, []string{"network"})
	if err != nil {
		t.Fatal(err)
	}
	routes := tableNamed(t, tables, "ルート")
	if routes.Rows[0][0] != "名前" || routes.Rows[1][0] != "LAB-ROUTE" {
		t.Fatal("route identity must remain the first column")
	}
	for label, want := range map[string]string{"管理距離 / Preference": "10", "メトリック": "25"} {
		found := false
		for i, header := range routes.Rows[0] {
			if header == label {
				found = true
				if routes.Rows[1][i] != want {
					t.Fatalf("%s=%s", label, routes.Rows[1][i])
				}
			}
		}
		if !found {
			t.Fatalf("missing %s column", label)
		}
	}
}
