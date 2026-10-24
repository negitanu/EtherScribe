package fortigate

import (
	"strings"

	"github.com/negitanu/etherscribe/internal/appcatalog"
	"github.com/negitanu/etherscribe/internal/model"
)

// convertSecurity ports WallScribe converters/security.py.
// Application ids resolve through the bundled catalog; unknown ids keep the raw id as the name.
func convertSecurity(cfg *model.Config, scope Node, vdom string) {
	addBasicProfiles(cfg, scope, vdom)
	addAntivirus(cfg, scope, vdom)
	addWebfilter(cfg, scope, vdom)
	addAppControl(cfg, scope, vdom)
	addIPS(cfg, scope, vdom)
	addSSLInspection(cfg, scope, vdom)
}

func addBasicProfiles(cfg *model.Config, scope Node, vdom string) {
	types := []struct{ section, kind string }{
		{"antivirus profile", "antivirus"},
		{"webfilter profile", "webfilter"},
		{"ips sensor", "ips"},
		{"firewall ssl-ssh-profile", "ssl-inspection"},
		{"application list", "app-control"},
	}
	for _, pt := range types {
		profiles := scope.Map(pt.section)
		if profiles == nil {
			continue
		}
		for name, raw := range profiles {
			if name == "_name" {
				continue
			}
			pn, ok := raw.(Node)
			if !ok {
				continue
			}
			cfg.SecurityProfiles = append(cfg.SecurityProfiles, model.SecurityProfile{
				Name:        firstNonEmpty(pn.String("_name"), name),
				ProfileType: pt.kind,
				Enabled:     true,
				VDOM:        vdom,
				Description: pn.String("comment"),
			})
		}
	}
}

func addAntivirus(cfg *model.Config, scope Node, vdom string) {
	profiles := scope.Map("antivirus profile")
	if profiles == nil {
		return
	}
	protos := []string{"http", "ftp", "imap", "pop3", "smtp", "mapi", "cifs", "ssh"}
	for name, raw := range profiles {
		if name == "_name" {
			continue
		}
		pn, ok := raw.(Node)
		if !ok {
			continue
		}
		var enabled []string
		for _, proto := range protos {
			block := pn.Map(proto)
			if block == nil {
				continue
			}
			if block.String("options") != "" || block.String("av-scan") == "enable" {
				enabled = append(enabled, proto)
			}
		}
		cfg.SecurityProfilesDetail.Antivirus = append(cfg.SecurityProfilesDetail.Antivirus, model.AntivirusProfile{
			Name:      firstNonEmpty(pn.String("_name"), name),
			Enabled:   true,
			ScanMode:  pn.String("scan-mode"),
			Protocols: enabled,
			VDOM:      vdom,
		})
	}
}

func addWebfilter(cfg *model.Config, scope Node, vdom string) {
	profiles := scope.Map("webfilter profile")
	if profiles == nil {
		return
	}
	for name, raw := range profiles {
		if name == "_name" {
			continue
		}
		pn, ok := raw.(Node)
		if !ok {
			continue
		}
		var cats []string
		if ftgd := pn.Map("ftgd-wf"); ftgd != nil {
			if filters := ftgd.Map("filters"); filters != nil {
				for id, fr := range filters {
					if id == "_name" {
						continue
					}
					fn, ok := fr.(Node)
					if !ok {
						continue
					}
					if cat := fn.String("category"); cat != "" {
						cats = append(cats, cat)
					}
				}
			}
		}
		cfg.SecurityProfilesDetail.Webfilter = append(cfg.SecurityProfilesDetail.Webfilter, model.WebFilterProfile{
			Name:       firstNonEmpty(pn.String("_name"), name),
			Enabled:    true,
			Categories: cats,
			VDOM:       vdom,
		})
	}
}

func addAppControl(cfg *model.Config, scope Node, vdom string) {
	profiles := scope.Map("application list")
	if profiles == nil {
		return
	}
	for name, raw := range profiles {
		if name == "_name" {
			continue
		}
		pn, ok := raw.(Node)
		if !ok {
			continue
		}
		var categories []string
		var apps []model.AppControlEntry
		if entries := pn.Map("entries"); entries != nil {
			for id, er := range entries {
				if id == "_name" {
					continue
				}
				en, ok := er.(Node)
				if !ok {
					continue
				}
				if cat := en.String("category"); cat != "" {
					categories = append(categories, cat)
				}
				action := en.String("action")
				for _, appID := range applicationIDs(en) {
					entry := model.AppControlEntry{AppID: appID, AppName: appID, Action: action}
					if info, ok := appcatalog.Lookup(appID); ok {
						if info.AppName != "" {
							entry.AppName = info.AppName
						}
						entry.Category = info.Category
						entry.Risk = info.Risk
					}
					apps = append(apps, entry)
				}
			}
		}
		cfg.SecurityProfilesDetail.AppControl = append(cfg.SecurityProfilesDetail.AppControl, model.AppControlProfile{
			Name:         firstNonEmpty(pn.String("_name"), name),
			Enabled:      true,
			Categories:   categories,
			Applications: apps,
			VDOM:         vdom,
		})
	}
}

func applicationIDs(n Node) []string {
	var out []string
	for _, item := range n.List("application") {
		for _, id := range strings.Fields(item) {
			if id != "" {
				out = append(out, id)
			}
		}
	}
	return out
}

func addIPS(cfg *model.Config, scope Node, vdom string) {
	profiles := scope.Map("ips sensor")
	if profiles == nil {
		return
	}
	for name, raw := range profiles {
		if name == "_name" {
			continue
		}
		pn, ok := raw.(Node)
		if !ok {
			continue
		}
		var sigs []string
		if entries := pn.Map("entries"); entries != nil {
			for id, er := range entries {
				if id == "_name" {
					continue
				}
				en, ok := er.(Node)
				if !ok {
					continue
				}
				if rule := en.String("rule"); rule != "" {
					sigs = append(sigs, rule)
				}
			}
		}
		cfg.SecurityProfilesDetail.IPS = append(cfg.SecurityProfilesDetail.IPS, model.IPSProfile{
			Name:       firstNonEmpty(pn.String("_name"), name),
			Enabled:    true,
			Signatures: sigs,
			Action:     pn.String("action"),
			VDOM:       vdom,
		})
	}
}

func addSSLInspection(cfg *model.Config, scope Node, vdom string) {
	profiles := scope.Map("firewall ssl-ssh-profile")
	if profiles == nil {
		return
	}
	for name, raw := range profiles {
		if name == "_name" {
			continue
		}
		pn, ok := raw.(Node)
		if !ok {
			continue
		}
		cfg.SecurityProfilesDetail.SSLInspection = append(cfg.SecurityProfilesDetail.SSLInspection, model.SSLInspectionProfile{
			Name:    firstNonEmpty(pn.String("_name"), name),
			Enabled: true,
			Mode:    detectSSLMode(pn),
			VDOM:    vdom,
		})
	}
}

func detectSSLMode(profile Node) string {
	name := strings.ToLower(profile.String("_name"))
	if strings.Contains(name, "deep-inspection") {
		return "deep-inspection"
	}
	if ssl := profile.Map("ssl"); ssl != nil && ssl.String("inspect-all") == "enable" {
		return "deep-inspection"
	}
	for _, proto := range []string{"https", "ftps", "imaps", "pop3s", "smtps", "dot", "ssh"} {
		block := profile.Map(proto)
		if block != nil && block.String("status") == "deep-inspection" {
			return "deep-inspection"
		}
	}
	return "certificate-inspection"
}
