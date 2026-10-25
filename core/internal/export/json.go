package export

import (
	"encoding/json"
	"time"

	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/redaction"
	"github.com/negitanu/etherscribe/internal/version"
)

const JSONSchema = "etherscribe.parameter-sheet"

type JSONDocument struct {
	Schema    string            `json:"schema"`
	Version   int               `json:"schemaVersion"`
	Generated string            `json:"generatedAt"`
	Generator map[string]string `json:"generator"`
	Sections  []string          `json:"sections,omitempty"`
	Redaction struct {
		Fields []string `json:"fields"`
	} `json:"redaction"`
	DeviceInfo             any `json:"deviceInfo,omitempty"`
	VirtualContexts        any `json:"virtualContexts,omitempty"`
	SystemSettings         any `json:"systemSettings,omitempty"`
	Interfaces             any `json:"interfaces,omitempty"`
	Routes                 any `json:"routes,omitempty"`
	VLANs                  any `json:"vlans,omitempty"`
	LAGs                   any `json:"lags,omitempty"`
	STP                    any `json:"stp,omitempty"`
	FirewallPolicies       any `json:"firewallPolicies,omitempty"`
	NATPolicies            any `json:"natPolicies,omitempty"`
	VPN                    any `json:"vpn,omitempty"`
	Objects                any `json:"objects,omitempty"`
	SecurityProfiles       any `json:"securityProfiles,omitempty"`
	SecurityProfilesDetail any `json:"securityProfilesDetail,omitempty"`
	DHCPRelays             any `json:"dhcpRelays,omitempty"`
	DHCPServers            any `json:"dhcpServers,omitempty"`
	Routing                any `json:"routing,omitempty"`
	HA                     any `json:"ha,omitempty"`
	Logging                any `json:"logging,omitempty"`
	Coverage               any `json:"coverage,omitempty"`
	ParseErrors            any `json:"parseErrors,omitempty"`
}

func RenderJSON(cfg *model.Config, sections []string) ([]byte, error) {
	doc := JSONDocument{
		Schema:    JSONSchema,
		Version:   1,
		Generated: time.Now().UTC().Format(time.RFC3339),
		Generator: map[string]string{"name": version.Name, "version": version.Version},
		Sections:  sections,
	}
	var redacted []string

	if SectionSelected(sections, "device_info") {
		doc.DeviceInfo = scrubValue(cfg.DeviceInfo, "deviceInfo", &redacted)
		doc.VirtualContexts = scrubValue(cfg.VirtualContexts, "virtualContexts", &redacted)
	}
	if SectionSelected(sections, "system_settings") {
		doc.SystemSettings = scrubValue(cfg.System, "systemSettings", &redacted)
	}
	if SectionSelected(sections, "network") {
		doc.Interfaces = scrubValue(cfg.Interfaces, "interfaces", &redacted)
		doc.Routes = scrubValue(cfg.Routes, "routes", &redacted)
		doc.DHCPRelays = scrubValue(cfg.DHCPRelays, "dhcpRelays", &redacted)
		doc.DHCPServers = scrubValue(cfg.DHCPServers, "dhcpServers", &redacted)
		doc.Routing = scrubValue(cfg.Routing, "routing", &redacted)
	}
	if SectionSelected(sections, "vlans") {
		doc.VLANs = scrubValue(cfg.VLANs, "vlans", &redacted)
	}
	if SectionSelected(sections, "lag") {
		doc.LAGs = scrubValue(cfg.LAGs, "lags", &redacted)
	}
	if SectionSelected(sections, "stp") {
		doc.STP = scrubValue(cfg.STP, "stp", &redacted)
	}
	if SectionSelected(sections, "policies") {
		doc.FirewallPolicies = scrubValue(cfg.Policies, "firewallPolicies", &redacted)
	}
	if SectionSelected(sections, "nat") {
		doc.NATPolicies = scrubValue(cfg.NATPolicies, "natPolicies", &redacted)
	}
	if SectionSelected(sections, "vpn") {
		doc.VPN = scrubValue(cfg.VPN, "vpn", &redacted)
	}
	if SectionSelected(sections, "objects") {
		doc.Objects = scrubValue(cfg.Objects, "objects", &redacted)
	}
	if SectionSelected(sections, "security_profiles") {
		doc.SecurityProfiles = scrubValue(cfg.SecurityProfiles, "securityProfiles", &redacted)
		doc.SecurityProfilesDetail = scrubValue(cfg.SecurityProfilesDetail, "securityProfilesDetail", &redacted)
	}
	if SectionSelected(sections, "ha") && fortiServicesApplicable(cfg) {
		doc.HA = scrubValue(cfg.HA, "ha", &redacted)
	}
	if SectionSelected(sections, "logging") {
		doc.Logging = scrubValue(cfg.Logging, "logging", &redacted)
		if !fortiServicesApplicable(cfg) {
			if logging, ok := doc.Logging.(map[string]any); ok {
				delete(logging, "fortianalyzerServer")
				delete(logging, "fortianalyzerStatus")
			}
		}
	}
	doc.Coverage = cfg.Coverage
	doc.ParseErrors = cfg.ParseErrors
	doc.Redaction.Fields = redacted
	return json.MarshalIndent(doc, "", "  ")
}

func scrubValue(v any, path string, redacted *[]string) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var generic any
	if err := json.Unmarshal(b, &generic); err != nil {
		return v
	}
	return walkScrub(generic, path, redacted)
}

func walkScrub(v any, path string, redacted *[]string) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, child := range t {
			childPath := path + "." + k
			if redaction.IsSecretField(k) {
				if s, ok := child.(string); ok && s != "" {
					out[k] = redaction.Redacted
					*redacted = append(*redacted, childPath)
					continue
				}
			}
			if s, ok := child.(string); ok {
				out[k] = redaction.ScrubSecrets(s)
				continue
			}
			out[k] = walkScrub(child, childPath, redacted)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, child := range t {
			out[i] = walkScrub(child, path+"[]", redacted)
		}
		return out
	case string:
		return redaction.ScrubSecrets(t)
	default:
		return v
	}
}
