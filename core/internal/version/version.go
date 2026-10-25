package version

const (
	Name    = "EtherScribe"
	Version = "1.0.0"
)

// Info is returned by CLI and RPC.
type Info struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Formats    []string `json:"formats"`
	Vendors    []string `json:"vendors"`
	RPCMethods []string `json:"rpcMethods"`
}

func Current() Info {
	return Info{
		Name:       Name,
		Version:    Version,
		Formats:    []string{"json", "html", "excel"},
		Vendors:    []string{"FortiGate", "PaloAlto", "JunosEX"},
		RPCMethods: []string{"engine.version", "convert.run", "compare.run", "sections.list", "cancel"},
	}
}
