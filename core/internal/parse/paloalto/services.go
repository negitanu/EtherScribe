package paloalto

import (
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
)

type dhcpBlock struct {
	Interfaces []dhcpInterface `xml:"interface>entry"`
}
type dhcpInterface struct {
	Name   string      `xml:"name,attr"`
	Server *dhcpServer `xml:"server"`
	Relay  struct {
		IP *struct {
			Enabled string     `xml:"enabled"`
			Servers memberList `xml:"server"`
		} `xml:"ip"`
	} `xml:"relay"`
}
type dhcpServer struct {
	Mode     string     `xml:"mode"`
	ProbeIP  string     `xml:"probe-ip"`
	Pools    memberList `xml:"ip-pool"`
	Reserved []struct {
		IP          string `xml:"name,attr"`
		MAC         string `xml:"mac"`
		Description string `xml:"description"`
	} `xml:"reserved>entry"`
	Option struct {
		Lease struct {
			Timeout   string    `xml:"timeout"`
			Unlimited *struct{} `xml:"unlimited"`
		} `xml:"lease"`
		Gateway string `xml:"gateway"`
		Netmask string `xml:"subnet-mask"`
		DNS     struct {
			Primary   string `xml:"primary"`
			Secondary string `xml:"secondary"`
		} `xml:"dns"`
	} `xml:"option"`
}
type logSettings struct {
	Syslog []syslogProfile `xml:"syslog>entry"`
}
type syslogProfile struct {
	Name    string `xml:"name,attr"`
	Servers []struct {
		Name      string `xml:"name,attr"`
		Server    string `xml:"server"`
		Transport string `xml:"transport"`
		Port      string `xml:"port"`
		Format    string `xml:"format"`
		Facility  string `xml:"facility"`
	} `xml:"server>entry"`
}

func convertDHCPServices(cfg *model.Config, block dhcpBlock) {
	for _, entry := range block.Interfaces {
		context := ""
		for _, iface := range cfg.Interfaces {
			if iface.Name == entry.Name {
				context = iface.VDOM
				break
			}
		}
		if entry.Server != nil {
			source := entry.Server
			server := model.DHCPServer{Interface: entry.Name, Mode: source.Mode, IPPools: source.Pools.Members, Netmask: source.Option.Netmask, Gateway: source.Option.Gateway, VDOM: context, Context: context}
			if source.ProbeIP != "" {
				probe := source.ProbeIP == "yes"
				server.ProbeIP = &probe
			}
			for _, dns := range []string{source.Option.DNS.Primary, source.Option.DNS.Secondary} {
				if dns != "" {
					server.DNSServers = append(server.DNSServers, dns)
				}
			}
			if source.Option.Lease.Unlimited != nil {
				server.LeaseType = "unlimited"
			} else if source.Option.Lease.Timeout != "" {
				server.LeaseType = "timeout"
				server.LeaseTime = source.Option.Lease.Timeout
			}
			if len(server.IPPools) > 0 {
				if start, end, ok := strings.Cut(server.IPPools[0], "-"); ok {
					server.StartIP = start
					server.EndIP = end
				}
			}
			for _, r := range source.Reserved {
				server.Reservations = append(server.Reservations, model.DHCPReservation{IP: r.IP, MAC: r.MAC, Description: r.Description})
				if r.MAC == "" {
					server.ExcludeIPs = append(server.ExcludeIPs, r.IP)
				}
			}
			cfg.DHCPServers = append(cfg.DHCPServers, server)
		}
		if entry.Relay.IP != nil {
			source := entry.Relay.IP
			cfg.DHCPRelays = append(cfg.DHCPRelays, model.DHCPRelay{Interface: entry.Name, Enabled: source.Enabled == "yes", Servers: source.Servers.Members, Context: context})
		}
	}
}
func convertSyslogProfiles(cfg *model.Config, settings logSettings, context string) {
	for _, profile := range settings.Syslog {
		for _, server := range profile.Servers {
			cfg.Logging.SyslogServers = append(cfg.Logging.SyslogServers, model.SyslogServer{Name: server.Name, Profile: profile.Name, Server: server.Server, Port: server.Port, Facility: server.Facility, Transport: server.Transport, Format: server.Format, Status: "configured", Context: context, VDOM: context})
		}
	}
}
