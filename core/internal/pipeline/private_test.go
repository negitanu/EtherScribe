package pipeline

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/ingest"
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/parse/fortigate"
)

type privateSummary struct {
	Vendor                                                                   string
	Interfaces, Addresses, Services, Policies, Routes, LAGs, VLANs, Coverage int
}
type sourceInventory struct {
	counts      map[string]int
	policyOrder []string
}

// This test is deliberately opt-in. Do not turn real configurations into fixtures,
// persist render results, log paths, or include raw errors in failure messages.
func TestPrivateConfigurations(t *testing.T) {
	root := os.Getenv("ETHERSCRIBE_PRIVATE_DATA_DIR")
	if root == "" {
		t.Skip("private validation is opt-in")
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return errors.New("input directory unavailable")
		}
		if d.IsDir() || d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".conf", ".xml", ".set", ".txt":
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal("private validation: input directory unavailable")
	}
	if len(files) == 0 {
		t.Fatal("private validation: no supported inputs")
	}
	sort.Strings(files)
	for i, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("sample %d: read failed", i+1)
			continue
		}
		summary, err := validatePrivateSample(data, strings.ToLower(filepath.Ext(path)))
		if err != nil {
			t.Errorf("sample %d: %s", i+1, err.Error())
			continue
		}
		t.Logf("sample %d: PASS vendor=%s interfaces=%d addresses=%d services=%d policies=%d routes=%d lags=%d vlans=%d coverage=%d; JSON/HTML/Excel validated in memory", i+1, summary.Vendor, summary.Interfaces, summary.Addresses, summary.Services, summary.Policies, summary.Routes, summary.LAGs, summary.VLANs, summary.Coverage)
	}
}
func validatePrivateSample(data []byte, ext string) (summary privateSummary, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("parser panic (details suppressed)")
		}
	}()
	input, err := ingest.ValidateBytes("private"+ext, data)
	if err != nil {
		return summary, errors.New("input validation failed")
	}
	cfg, err := ParseContent(input.Content, "private"+ext)
	if err != nil {
		return summary, errors.New("parse failed")
	}
	if len(cfg.ParseErrors) > 0 {
		return summary, errors.New("parse warnings detected")
	}
	var source sourceInventory
	switch cfg.DeviceInfo.DeviceType {
	case identify.FortiGate:
		source = fortiSourceInventory(input.Content)
	case identify.PaloAlto:
		source, err = paloSourceInventory(input.Content)
		if err != nil {
			return summary, errors.New("source inventory failed")
		}
	default:
		return summary, errors.New("inventory validation unsupported for this vendor")
	}
	actual := modelInventory(cfg)
	for key, want := range source.counts {
		if actual[key] != want {
			return summary, fmt.Errorf("inventory mismatch: %s source=%d parsed=%d", key, want, actual[key])
		}
	}
	if cfg.DeviceInfo.DeviceType == identify.FortiGate {
		var order []string
		for _, p := range cfg.Policies {
			order = append(order, p.VDOM+"/"+p.PolicyID)
		}
		if !reflect.DeepEqual(order, source.policyOrder) {
			return summary, errors.New("policy order mismatch")
		}
	}
	for _, format := range []string{"json", "html", "excel"} {
		data, err := Render(cfg, format, nil)
		if err != nil || len(data) == 0 {
			return summary, errors.New("render failed")
		}
		if format == "excel" {
			zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				return summary, errors.New("Excel archive invalid")
			}
			for _, file := range zr.File {
				r, err := file.Open()
				if err != nil {
					return summary, errors.New("Excel entry invalid")
				}
				decoder := xml.NewDecoder(r)
				for {
					_, err = decoder.Token()
					if err != nil {
						break
					}
				}
				r.Close()
				if err != io.EOF {
					return summary, errors.New("Excel XML invalid")
				}
			}
		}
	}
	summary = privateSummary{Vendor: string(cfg.DeviceInfo.DeviceType), Interfaces: len(cfg.Interfaces), Addresses: len(cfg.Objects.Addresses), Services: len(cfg.Objects.Services), Policies: len(cfg.Policies), Routes: len(cfg.Routes), LAGs: len(cfg.LAGs), VLANs: len(cfg.VLANs), Coverage: len(cfg.Coverage)}
	return summary, nil
}
func modelInventory(cfg *model.Config) map[string]int {
	return map[string]int{"interfaces": len(cfg.Interfaces), "addresses": len(cfg.Objects.Addresses), "groups": len(cfg.Objects.AddressGroups), "services": len(cfg.Objects.Services), "serviceGroups": len(cfg.Objects.ServiceGroups), "policies": len(cfg.Policies), "routes": len(cfg.Routes), "nat": len(cfg.NATPolicies), "lags": len(cfg.LAGs), "vlans": len(cfg.VLANs)}
}

// Count source definitions independently of the configuration tree builder.
// The stack stores identifiers only in memory to deduplicate re-edits and check order.
func fortiSourceInventory(content string) sourceInventory {
	out := sourceInventory{counts: map[string]int{"interfaces": 0, "addresses": 0, "groups": 0, "services": 0, "serviceGroups": 0, "policies": 0, "routes": 0, "nat": 0}}
	kinds := map[string]string{"system interface": "interfaces", "firewall address": "addresses", "firewall address6": "addresses", "firewall addrgrp": "groups", "firewall addrgrp6": "groups", "firewall service custom": "services", "firewall service group": "serviceGroups", "firewall policy": "policies", "router static": "routes", "router static6": "routes", "firewall vip": "nat", "firewall ippool": "nat"}
	type frame struct{ kind, section, name string }
	var stack []frame
	seen := map[string]bool{}
	lines, _ := fortigate.SeparateLogicalLines(content)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, parts[0]))
		switch parts[0] {
		case "config":
			stack = append(stack, frame{kind: "config", section: rest})
		case "edit":
			if len(stack) == 0 {
				continue
			}
			section := stack[len(stack)-1].section
			name := rest
			if decoded, err := strconv.Unquote(rest); err == nil {
				name = decoded
			} else {
				name = strings.Trim(rest, "'")
			}
			scope := "root"
			for _, f := range stack {
				if f.kind == "edit" && f.section == "vdom" {
					scope = f.name
				}
			}
			if kind, ok := kinds[section]; ok {
				key := section + "\x00" + scope + "\x00" + name
				if !seen[key] {
					seen[key] = true
					out.counts[kind]++
					if kind == "policies" {
						out.policyOrder = append(out.policyOrder, scope+"/"+name)
					}
				}
			}
			stack = append(stack, frame{kind: "edit", section: section, name: name})
		case "next":
			if len(stack) > 0 && stack[len(stack)-1].kind == "edit" {
				stack = stack[:len(stack)-1]
			}
		case "end":
			if len(stack) > 0 && stack[len(stack)-1].kind == "edit" && stack[len(stack)-1].section == "vdom" {
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return out
}
func paloSourceInventory(content string) (sourceInventory, error) {
	out := sourceInventory{counts: map[string]int{"interfaces": 0, "addresses": 0, "groups": 0, "services": 0, "serviceGroups": 0, "policies": 0, "routes": 0, "nat": 0, "lags": 0, "vlans": 0}}
	paths := map[string]string{}
	for _, prefix := range []string{"config/shared", "config/devices/entry/vsys/entry"} {
		for tag, key := range map[string]string{"address": "addresses", "address-group": "groups", "service": "services", "service-group": "serviceGroups"} {
			paths[prefix+"/"+tag+"/entry"] = key
		}
	}
	base := "config/devices/entry/network/"
	for _, kind := range []string{"ethernet", "aggregate-ethernet"} {
		paths[base+"interface/"+kind+"/entry"] = "interfaces"
		for _, mode := range []string{"layer2", "layer3", "virtual-wire"} {
			paths[base+"interface/"+kind+"/entry/"+mode+"/units/entry"] = "interfaces"
		}
	}
	for _, kind := range []string{"vlan", "loopback", "tunnel"} {
		paths[base+"interface/"+kind+"/units/entry"] = "interfaces"
	}
	paths[base+"vlan/entry"] = "vlans"
	for _, family := range []string{"ip", "ipv6"} {
		paths[base+"virtual-router/entry/routing-table/"+family+"/static-route/entry"] = "routes"
	}
	paths["config/devices/entry/vsys/entry/rulebase/security/rules/entry"] = "policies"
	paths["config/devices/entry/vsys/entry/rulebase/nat/rules/entry"] = "nat"
	decoder := xml.NewDecoder(strings.NewReader(content))
	var stack []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, errors.New("source XML invalid")
		}
		switch token := token.(type) {
		case xml.StartElement:
			stack = append(stack, token.Name.Local)
			path := strings.Join(stack, "/")
			if key, ok := paths[path]; ok {
				out.counts[key]++
			}
			if path == base+"interface/aggregate-ethernet/entry" {
				out.counts["lags"]++
			}
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		}
	}
	return out, nil
}
func TestPrivateValidationWithSyntheticInputs(t *testing.T) {
	for _, fixture := range []string{"fortigate-vdom-boundaries.conf", "paloalto-network-ops.xml"} {
		data, err := os.ReadFile(testdata(fixture))
		if err != nil {
			t.Fatal(err)
		}
		_, err = validatePrivateSample(data, filepath.Ext(fixture))
		if err != nil {
			t.Fatalf("synthetic private validation: %v", err)
		}
	}
	_, err := validatePrivateSample([]byte(`<config version="11"><SECRET_SENTINEL></config>`), ".xml")
	if err == nil || strings.Contains(err.Error(), "SECRET_SENTINEL") {
		t.Fatal("private validator must suppress raw parse diagnostics")
	}
}
