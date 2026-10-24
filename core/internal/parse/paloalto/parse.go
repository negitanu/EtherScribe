package paloalto

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/model"
)

type configRoot struct {
	XMLName xml.Name `xml:"config"`
	Version string   `xml:"version,attr"`
	Detail  string   `xml:"detail-version,attr"`
	Devices devices  `xml:"devices"`
	Shared  shared   `xml:"shared"`
}

type devices struct {
	Entries []deviceEntry `xml:"entry"`
}

type deviceEntry struct {
	Name         string       `xml:"name,attr"`
	Network      network      `xml:"network"`
	Vsys         vsysBlk      `xml:"vsys"`
	DeviceConfig deviceConfig `xml:"deviceconfig"`
}

type deviceConfig struct {
	System systemCfg `xml:"system"`
}

type systemCfg struct {
	Hostname string `xml:"hostname"`
}

type network struct {
	Interface     interfaceBlk `xml:"interface"`
	VirtualRouter vrBlk        `xml:"virtual-router"`
	IKE           ikeBlk       `xml:"ike"`
}

type ikeBlk struct {
	Gateway ikeGatewayBlk `xml:"gateway"`
}

type ikeGatewayBlk struct {
	Entries []ikeGatewayEntry `xml:"entry"`
}

type ikeGatewayEntry struct {
	Name           string          `xml:"name,attr"`
	Authentication ikeAuth         `xml:"authentication"`
	LocalAddress   ikeLocalAddress `xml:"local-address"`
	PeerAddress    ikePeerAddress  `xml:"peer-address"`
	Protocol       ikeProtocol     `xml:"protocol"`
}

type ikeAuth struct {
	PreSharedKey ikePSK `xml:"pre-shared-key"`
}

type ikePSK struct {
	Key string `xml:"key"`
}

type ikeLocalAddress struct {
	Interface string `xml:"interface"`
}

type ikePeerAddress struct {
	IP string `xml:"ip"`
}

type ikeProtocol struct {
	IKEV1 ikeVersionBlk `xml:"ikev1"`
	IKEV2 ikeVersionBlk `xml:"ikev2"`
}

type ikeVersionBlk struct {
	IKECryptoProfile string `xml:"ike-crypto-profile"`
}

type interfaceBlk struct {
	Ethernet ethernetBlk `xml:"ethernet"`
	VLAN     entryList   `xml:"vlan"`
}

type ethernetBlk struct {
	Entries []ifaceEntry `xml:"entry"`
}

type ifaceEntry struct {
	Name   string `xml:"name,attr"`
	Layer3 layer3 `xml:"layer3"`
}

type layer3 struct {
	IP ipBlk `xml:"ip"`
}

type ipBlk struct {
	Entries []named `xml:"entry"`
}

type named struct {
	Name string `xml:"name,attr"`
}

type vrBlk struct {
	Entries []vrEntry `xml:"entry"`
}

type vrEntry struct {
	Name         string   `xml:"name,attr"`
	RoutingTable routeTbl `xml:"routing-table"`
}

type routeTbl struct {
	IP   staticRoutes `xml:"ip"`
	IPv6 staticRoutes `xml:"ipv6"`
}

type staticRoutes struct {
	Static staticRouteList `xml:"static-route"`
}

type staticRouteList struct {
	Entries []routeEntry `xml:"entry"`
}

type routeEntry struct {
	Name        string  `xml:"name,attr"`
	Destination string  `xml:"destination"`
	Nexthop     nexthop `xml:"nexthop"`
}

type nexthop struct {
	IP     string `xml:"ip-address"`
	NextVR string `xml:"next-vr"`
}

type vsysBlk struct {
	Entries []vsysEntry `xml:"entry"`
}

type vsysEntry struct {
	Name         string     `xml:"name,attr"`
	Zone         entryList  `xml:"zone"`
	Rulebase     rulebase   `xml:"rulebase"`
	Import       importBlk  `xml:"import"`
	Address      addrBlk    `xml:"address"`
	AddressGroup addrGrpBlk `xml:"address-group"`
	Service      svcBlk     `xml:"service"`
	ServiceGroup svcGrpBlk  `xml:"service-group"`
}

type addrBlk struct {
	Entries []addrEntry `xml:"entry"`
}

type addrEntry struct {
	Name        string `xml:"name,attr"`
	IPNetmask   string `xml:"ip-netmask"`
	FQDN        string `xml:"fqdn"`
	IPRange     string `xml:"ip-range"`
	Description string `xml:"description"`
}

type addrGrpBlk struct {
	Entries []addrGrpEntry `xml:"entry"`
}

type addrGrpEntry struct {
	Name        string     `xml:"name,attr"`
	Static      memberList `xml:"static"`
	Description string     `xml:"description"`
}

type svcBlk struct {
	Entries []svcEntry `xml:"entry"`
}

type svcEntry struct {
	Name     string   `xml:"name,attr"`
	Protocol svcProto `xml:"protocol"`
}

type svcProto struct {
	TCP svcPort `xml:"tcp"`
	UDP svcPort `xml:"udp"`
}

type svcPort struct {
	Port string `xml:"port"`
}

type svcGrpBlk struct {
	Entries []svcGrpEntry `xml:"entry"`
}

type svcGrpEntry struct {
	Name    string     `xml:"name,attr"`
	Members memberList `xml:"members"`
}

type entryList struct {
	Entries []zoneEntry `xml:"entry"`
}

type zoneEntry struct {
	Name    string      `xml:"name,attr"`
	Network zoneNetwork `xml:"network"`
}

type zoneNetwork struct {
	Layer3 memberList `xml:"layer3"`
}

type rulebase struct {
	Security securityRules `xml:"security"`
	NAT      natRules      `xml:"nat"`
}

type natRules struct {
	Rules natRulesList `xml:"rules"`
}

type natRuleEntry struct {
	Name                   string          `xml:"name,attr"`
	From                   memberList      `xml:"from"`
	To                     memberList      `xml:"to"`
	Source                 memberList      `xml:"source"`
	Destination            memberList      `xml:"destination"`
	Service                string          `xml:"service"`
	ToInterface            string          `xml:"to-interface"`
	Disabled               string          `xml:"disabled"`
	Description            string          `xml:"description"`
	SourceTranslation      *srcTranslation `xml:"source-translation"`
	DestinationTranslation *dstTranslation `xml:"destination-translation"`
}

type srcTranslation struct {
	DynamicIPAndPort *dynamicIPPort `xml:"dynamic-ip-and-port"`
	StaticIP         *staticIPTrans `xml:"static-ip"`
}

type dynamicIPPort struct {
	InterfaceAddress  *ifaceAddr `xml:"interface-address"`
	TranslatedAddress memberList `xml:"translated-address"`
}

type ifaceAddr struct {
	Interface string `xml:"interface"`
	IP        string `xml:"ip"`
}

type staticIPTrans struct {
	TranslatedAddress string `xml:"translated-address"`
}

type dstTranslation struct {
	TranslatedAddress string `xml:"translated-address"`
	TranslatedPort    string `xml:"translated-port"`
}

type securityRules struct {
	Rules rulesList `xml:"rules"`
}

type rulesList struct {
	Entries []ruleEntry `xml:"entry"`
}

type natRulesList struct {
	Entries []natRuleEntry `xml:"entry"`
}

type ruleEntry struct {
	Name        string     `xml:"name,attr"`
	From        memberList `xml:"from"`
	To          memberList `xml:"to"`
	Source      memberList `xml:"source"`
	Destination memberList `xml:"destination"`
	Service     memberList `xml:"service"`
	Application memberList `xml:"application"`
	Action      string     `xml:"action"`
	Disabled    string     `xml:"disabled"`
	LogEnd      string     `xml:"log-end"`
}

type memberList struct {
	Members []string `xml:"member"`
}

type importBlk struct {
	Network importNet `xml:"network"`
}

type importNet struct {
	Interface memberList `xml:"interface"`
}

type shared struct {
	Application entryList `xml:"application"`
}

// Parse parses Palo Alto XML export (MVP subset).
func Parse(content, filename string) (*model.Config, error) {
	cfg := &model.Config{
		SourceFile: filename,
		DeviceInfo: model.DeviceInfo{
			DeviceType: identify.PaloAlto, Role: model.RoleFirewall, VendorFamily: "paloalto",
			Model: "PA Series", OperationMode: "NAT/Route",
		},
		System:  model.SystemSettings{HTTPSPort: "443", SSHPort: "22"},
		Objects: model.Objects{},
	}

	var root configRoot
	if err := xml.Unmarshal([]byte(content), &root); err != nil {
		return nil, fmt.Errorf("XMLパースエラー: %w", err)
	}

	if root.Detail != "" {
		cfg.DeviceInfo.OSVersion = root.Detail
		cfg.DeviceInfo.OSVersionSource = "detail-version"
	} else if root.Version != "" {
		cfg.DeviceInfo.OSVersion = root.Version
		cfg.DeviceInfo.OSVersionSource = "version"
	}
	if id := identify.Identify(content); id != nil && cfg.DeviceInfo.OSVersion == "" {
		cfg.DeviceInfo.OSVersion = id.OSVersion
	}

	var device *deviceEntry
	for i := range root.Devices.Entries {
		e := &root.Devices.Entries[i]
		if e.Name == "localhost.localdomain" || device == nil {
			device = e
		}
	}
	if device == nil {
		cfg.ParseErrors = append(cfg.ParseErrors, "devices/entry が見つかりません")
		cfg.Coverage = append(cfg.Coverage, model.CoverageItem{Path: "devices", Reason: "デバイス節欠落"})
		return cfg, nil
	}

	if device.DeviceConfig.System.Hostname != "" {
		cfg.DeviceInfo.Hostname = device.DeviceConfig.System.Hostname
	} else {
		cfg.DeviceInfo.Hostname = device.Name
	}

	type ifaceHint struct {
		zone string
		vdom string
	}
	hints := map[string]ifaceHint{}

	for _, e := range device.Network.Interface.Ethernet.Entries {
		ip := ""
		if len(e.Layer3.IP.Entries) > 0 {
			ip = e.Layer3.IP.Entries[0].Name
		}
		cfg.Interfaces = append(cfg.Interfaces, model.Interface{
			Name:          e.Name,
			InterfaceType: "physical",
			IPAddress:     ip,
			Status:        "up",
		})
	}

	for _, vr := range device.Network.VirtualRouter.Entries {
		for _, r := range vr.RoutingTable.IP.Static.Entries {
			gw := r.Nexthop.IP
			if gw == "" && r.Nexthop.NextVR != "" {
				gw = "next-vr:" + r.Nexthop.NextVR
			}
			cfg.Routes = append(cfg.Routes, model.Route{
				Name:        r.Name,
				Destination: r.Destination,
				Gateway:     gw,
				VDOM:        vr.Name,
				Context:     vr.Name,
				RouteType:   "static",
				Enabled:     true,
			})
		}
		for _, r := range vr.RoutingTable.IPv6.Static.Entries {
			cfg.Routes = append(cfg.Routes, model.Route{
				Name:        r.Name,
				Destination: r.Destination,
				RouteType:   "static6",
				VDOM:        vr.Name,
				Context:     vr.Name,
				Enabled:     true,
			})
		}
	}

	var vsysNames []string
	for _, vs := range device.Vsys.Entries {
		vsysNames = append(vsysNames, vs.Name)
		for _, m := range vs.Import.Network.Interface.Members {
			h := hints[m]
			h.vdom = vs.Name
			hints[m] = h
		}
		for _, z := range vs.Zone.Entries {
			for _, m := range z.Network.Layer3.Members {
				h := hints[m]
				h.zone = z.Name
				if h.vdom == "" {
					h.vdom = vs.Name
				}
				hints[m] = h
			}
		}
		convertVsysObjects(cfg, vs)
		convertVsysNAT(cfg, vs)
		for i, rule := range vs.Rulebase.Security.Rules.Entries {
			action := rule.Action
			if action == "" {
				action = "unknown"
			}
			cfg.Policies = append(cfg.Policies, model.Policy{
				PolicyID:             fmt.Sprintf("%d", i+1),
				Name:                 rule.Name,
				SourceInterface:      rule.From.Members,
				DestinationInterface: rule.To.Members,
				SourceAddress:        rule.Source.Members,
				DestinationAddress:   rule.Destination.Members,
				Service:              rule.Service.Members,
				Action:               action,
				LogEnabled:           rule.LogEnd == "yes",
				VDOM:                 vs.Name,
				Enabled:              rule.Disabled != "yes",
			})
		}
	}
	for i := range cfg.Interfaces {
		h, ok := hints[cfg.Interfaces[i].Name]
		if !ok {
			continue
		}
		cfg.Interfaces[i].Zone = h.zone
		cfg.Interfaces[i].VDOM = h.vdom
	}
	if len(vsysNames) > 0 {
		cfg.DeviceInfo.VDOMList = vsysNames
		cfg.DeviceInfo.VDOMEnabled = len(vsysNames) > 1
	}

	convertIKEGateways(cfg, device)

	if len(root.Shared.Application.Entries) > 0 {
		cfg.Coverage = append(cfg.Coverage, model.CoverageItem{
			Path:   "shared/application",
			Reason: "アプリケーション定義はオブジェクト詳細へ未展開（件数のみ把握）",
			Note:   fmt.Sprintf("%d entries", len(root.Shared.Application.Entries)),
		})
	}
	return cfg, nil
}

func convertVsysNAT(cfg *model.Config, vs vsysEntry) {
	for _, rule := range vs.Rulebase.NAT.Rules.Entries {
		natType := "nat"
		translatedSrc, translatedDst, translatedPort := "", "", ""
		if rule.SourceTranslation != nil {
			natType = "snat"
			if dip := rule.SourceTranslation.DynamicIPAndPort; dip != nil {
				if dip.InterfaceAddress != nil {
					translatedSrc = dip.InterfaceAddress.Interface
					if dip.InterfaceAddress.IP != "" {
						translatedSrc += " (" + dip.InterfaceAddress.IP + ")"
					}
				}
				if translatedSrc == "" && len(dip.TranslatedAddress.Members) > 0 {
					translatedSrc = strings.Join(dip.TranslatedAddress.Members, ", ")
				}
			}
			if st := rule.SourceTranslation.StaticIP; st != nil {
				translatedSrc = st.TranslatedAddress
				natType = "static"
			}
		}
		if rule.DestinationTranslation != nil {
			if natType == "snat" {
				natType = "static"
			} else if natType == "nat" {
				natType = "dnat"
			}
			translatedDst = rule.DestinationTranslation.TranslatedAddress
			translatedPort = rule.DestinationTranslation.TranslatedPort
		}
		zoneInfo := ""
		if len(rule.From.Members) > 0 || len(rule.To.Members) > 0 {
			zoneInfo = strings.Join(rule.From.Members, ", ") + " -> " + strings.Join(rule.To.Members, ", ")
		}
		cfg.NATPolicies = append(cfg.NATPolicies, model.NATPolicy{
			Name:                  rule.Name,
			NATType:               natType,
			OriginalSource:        strings.Join(rule.Source.Members, ", "),
			OriginalDestination:   strings.Join(rule.Destination.Members, ", "),
			TranslatedSource:      translatedSrc,
			TranslatedDestination: translatedDst,
			TranslatedPort:        translatedPort,
			Interface:             zoneInfo,
			ExternalInterface:     rule.ToInterface,
			VDOM:                  vs.Name,
			Description:           rule.Description,
		})
	}
}

func convertIKEGateways(cfg *model.Config, device *deviceEntry) {
	for _, gw := range device.Network.IKE.Gateway.Entries {
		ikeVer := ""
		proposal := ""
		switch {
		case gw.Protocol.IKEV2.IKECryptoProfile != "":
			ikeVer = "2"
			proposal = gw.Protocol.IKEV2.IKECryptoProfile
		case gw.Protocol.IKEV1.IKECryptoProfile != "":
			ikeVer = "1"
			proposal = gw.Protocol.IKEV1.IKECryptoProfile
		}
		cfg.VPN.IPSecPhase1 = append(cfg.VPN.IPSecPhase1, model.IPSecPhase1{
			Name:          gw.Name,
			RemoteGateway: gw.PeerAddress.IP,
			Interface:     gw.LocalAddress.Interface,
			IKEVersion:    ikeVer,
			Proposal:      proposal,
			PSKConfigured: gw.Authentication.PreSharedKey.Key != "",
			VDOM:          "shared",
		})
	}
}

func convertVsysObjects(cfg *model.Config, vs vsysEntry) {
	vdom := vs.Name
	for _, a := range vs.Address.Entries {
		addrType, value := "ip-netmask", a.IPNetmask
		switch {
		case a.FQDN != "":
			addrType, value = "fqdn", a.FQDN
		case a.IPRange != "":
			addrType, value = "ip-range", a.IPRange
		}
		cfg.Objects.Addresses = append(cfg.Objects.Addresses, model.AddressObject{
			Name: a.Name, Type: addrType, Value: value, VDOM: vdom, Description: a.Description,
		})
	}
	for _, g := range vs.AddressGroup.Entries {
		cfg.Objects.AddressGroups = append(cfg.Objects.AddressGroups, model.NamedMembers{
			Name: g.Name, Members: g.Static.Members, VDOM: vdom, Description: g.Description,
		})
	}
	for _, s := range vs.Service.Entries {
		proto, port := "", ""
		switch {
		case s.Protocol.TCP.Port != "":
			proto, port = "tcp", s.Protocol.TCP.Port
		case s.Protocol.UDP.Port != "":
			proto, port = "udp", s.Protocol.UDP.Port
		}
		cfg.Objects.Services = append(cfg.Objects.Services, model.ServiceObject{
			Name: s.Name, Protocol: proto, Port: port, VDOM: vdom,
		})
	}
	for _, g := range vs.ServiceGroup.Entries {
		cfg.Objects.ServiceGroups = append(cfg.Objects.ServiceGroups, model.NamedMembers{
			Name: g.Name, Members: g.Members.Members, VDOM: vdom,
		})
	}
}

// DetectContent mirrors PaloAltoParser.detect_content_type.
func DetectContent(content string) bool {
	if strings.Contains(content, "<config version=") || strings.Contains(content, "<devices>") {
		return true
	}
	id := identify.Identify(content)
	return id != nil && id.Type == identify.PaloAlto
}
