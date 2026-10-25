package pipeline

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestPANOSDHCPAndSyslog(t *testing.T) {
	cfg := groundedConfig(t, "paloalto-dhcp-syslog.xml")
	if len(cfg.DHCPServers) != 3 || len(cfg.DHCPRelays) != 1 {
		t.Fatalf("servers=%+v relays=%+v", cfg.DHCPServers, cfg.DHCPRelays)
	}
	d := cfg.DHCPServers[0]
	if d.Interface != "ethernet1/1" || d.Mode != "auto" || d.Context != "vsys1" || d.ProbeIP == nil || !*d.ProbeIP || d.LeaseType != "timeout" || d.LeaseTime != "1440" || d.Gateway != "192.0.2.1" || d.Netmask != "255.255.255.0" {
		t.Fatalf("DHCP=%+v", d)
	}
	if !reflect.DeepEqual(d.IPPools, []string{"192.0.2.20-192.0.2.60", "198.51.100.0/24"}) || d.StartIP != "192.0.2.20" || d.EndIP != "192.0.2.60" || !reflect.DeepEqual(d.DNSServers, []string{"192.0.2.53", "inherited"}) {
		t.Fatalf("DHCP pools/options=%+v", d)
	}
	if len(d.Reservations) != 2 || d.Reservations[0].MAC != "02:00:00:00:00:30" || !reflect.DeepEqual(d.ExcludeIPs, []string{"192.0.2.31"}) {
		t.Fatalf("reserved=%+v", d)
	}
	if cfg.DHCPServers[1].Mode != "disabled" || cfg.DHCPServers[1].LeaseType != "unlimited" || cfg.DHCPServers[1].LeaseTime != "" || cfg.DHCPServers[1].ProbeIP == nil || *cfg.DHCPServers[1].ProbeIP {
		t.Fatal("disabled/unlimited/probe lost")
	}
	if cfg.DHCPServers[2].StartIP != "" || cfg.DHCPServers[2].EndIP != "" || cfg.DHCPServers[2].IPPools[0] != "192.0.2.128/25" {
		t.Fatal("CIDR pool misrepresented as range")
	}
	r := cfg.DHCPRelays[0]
	if !r.Enabled || r.Context != "vsys1" || !reflect.DeepEqual(r.Servers, []string{"192.0.2.70", "192.0.2.71"}) {
		t.Fatalf("relay=%+v", r)
	}
	if cfg.Logging.SyslogConfigured == nil || !*cfg.Logging.SyslogConfigured || len(cfg.Logging.SyslogServers) != 2 {
		t.Fatalf("logging=%+v", cfg.Logging)
	}
	s := cfg.Logging.SyslogServers[0]
	if s.Profile != "SHARED-SIEM" || s.Context != "shared" || s.Transport != "SSL" || s.Port != "6514" || s.Format != "IETF" || s.Facility != "LOG_LOCAL4" || s.Status != "configured" {
		t.Fatalf("Syslog=%+v", s)
	}
	if cfg.Logging.SyslogServers[1].Context != "vsys1" || cfg.Logging.SyslogServers[1].Profile != "VSYS-SIEM" {
		t.Fatal("vsys syslog missing")
	}
	paths := map[string]bool{}
	for _, c := range cfg.Coverage {
		paths[c.Path] = true
		if strings.Contains(c.Path, "UNMODELED") || strings.Contains(c.Note, "UNMODELED") {
			t.Fatal("coverage leaked values")
		}
	}
	for _, path := range []string{"devices/entry/network/dhcp/interface/entry/server/option/dns-suffix", "devices/entry/network/dhcp/interface/entry/relay/ipv6", "devices/entry/vsys/entry/log-settings/syslog/entry/format"} {
		if !paths[path] {
			t.Errorf("partial coverage missing: %s", path)
		}
	}
}

func serviceReportText(t *testing.T, data []byte, format string) string {
	t.Helper()
	if format != "excel" {
		return string(data)
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		b.Write(data)
	}
	return b.String()
}
func TestVendorServiceReports(t *testing.T) {
	for _, format := range []string{"json", "html", "excel"} {
		t.Run(format, func(t *testing.T) {
			for _, sections := range [][]string{nil, {"logging"}, {"network"}} {
				res, err := ConvertFile(ConvertRequest{Path: testdata("paloalto-dhcp-syslog.xml"), Format: format, Sections: sections})
				if err != nil {
					t.Fatal(err)
				}
				text := serviceReportText(t, res.Bytes, format)
				for _, bad := range []string{"FortiAnalyzer", "fortianalyzer", "HA グループ", "UNMODELED-FORMAT", "UNMODELED-SUFFIX"} {
					if strings.Contains(text, bad) {
						t.Fatalf("inappropriate data in Palo Alto report: %s", bad)
					}
				}
				markers := []string{"SHARED-SIEM", "collector-b", "6514"}
				if len(sections) == 0 || sections[0] == "network" {
					markers = []string{"198.51.100.0/24", "02:00:00:00:00:30", "192.0.2.71"}
				}
				for _, m := range markers {
					if !strings.Contains(text, m) {
						t.Fatalf("missing %s", m)
					}
				}
				if format == "json" {
					var doc map[string]any
					if err := json.Unmarshal(res.Bytes, &doc); err != nil {
						t.Fatal(err)
					}
					if len(sections) == 0 || sections[0] == "logging" {
						logging := doc["logging"].(map[string]any)
						if logging["syslogConfigured"] != true {
							t.Fatal("Syslog presence missing")
						}
					}
				}
			}
			absent := groundedConfig(t, "corpus-paloalto-complex.xml")
			if absent.Logging.SyslogConfigured == nil || *absent.Logging.SyslogConfigured {
				t.Fatal("absent Syslog not marked false")
			}
			// Even stale/injected foreign vendor values must not resurrect foreign fields.
			absent.Logging.FortiAnalyzerServer = "FOREIGN-VENDOR-MARKER"
			absent.Logging.FortiAnalyzerStatus = "enabled"
			rendered, err := Render(absent, format, []string{"logging"})
			if err != nil {
				t.Fatal(err)
			}
			text := serviceReportText(t, rendered, format)
			if strings.Contains(text, "FortiAnalyzer") || strings.Contains(text, "fortianalyzer") || strings.Contains(text, "FOREIGN-VENDOR-MARKER") {
				t.Fatal("foreign field present")
			}
			expected := `"syslogConfigured": false`
			if format != "json" {
				expected = "設定なし"
			}
			if !strings.Contains(text, expected) {
				t.Fatal("missing Syslog absence")
			}
			fg := groundedConfig(t, "fortigate-service-report.conf")
			if fg.Logging.SyslogConfigured == nil || !*fg.Logging.SyslogConfigured || len(fg.Logging.SyslogServers) != 2 || fg.Logging.SyslogServers[1].Status != "disabled" {
				t.Fatal("FortiGate Syslog lost")
			}
			rendered, err = Render(fg, format, []string{"logging"})
			if err != nil {
				t.Fatal(err)
			}
			text = serviceReportText(t, rendered, format)
			if !strings.Contains(text, "203.0.113.200") || !strings.Contains(text, "198.51.100.200") {
				t.Fatal("FortiGate service data lost")
			}
		})
	}
}

func TestJunosDoesNotClaimUnparsedSyslogIsAbsent(t *testing.T) {
	cfg := groundedConfig(t, "corpus-junos-complex.set")
	if cfg.Logging.SyslogConfigured != nil {
		t.Fatal("unparsed Syslog given definite presence")
	}
	data, err := Render(cfg, "html", []string{"logging"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "未解析") || strings.Contains(string(data), "FortiAnalyzer") {
		t.Fatal("inaccurate Junos service output")
	}
}
