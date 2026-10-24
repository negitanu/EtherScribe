package model

import "github.com/negitanu/etherscribe/internal/identify"

// DefaultContext is the implicit global virtual context (empty Context means this).
const DefaultContext = "default"

// Device roles for preset / section filtering.
const (
	RoleFirewall = "firewall"
	RoleSwitch   = "switch"
)

// Virtual context kinds (VDOM / vsys / Juniper RI / cisco VRF).
const (
	ContextVDOM            = "vdom"
	ContextVsys            = "vsys"
	ContextRoutingInstance = "routing-instance"
	ContextVRF             = "vrf"
	ContextDefault         = "default"
)

// Config is the unified intermediate model (firewall + switch fabric).
type Config struct {
	SourceFile             string                 `json:"sourceFile,omitempty"`
	DeviceInfo             DeviceInfo             `json:"deviceInfo"`
	VirtualContexts        []VirtualContext       `json:"virtualContexts,omitempty"`
	System                 SystemSettings         `json:"systemSettings"`
	Interfaces             []Interface            `json:"interfaces"`
	VLANs                  []VLAN                 `json:"vlans,omitempty"`
	LAGs                   []LAG                  `json:"lags,omitempty"`
	STP                    *STPSettings           `json:"stp,omitempty"`
	Routes                 []Route                `json:"routes"`
	Policies               []Policy               `json:"firewallPolicies"`
	NATPolicies            []NATPolicy            `json:"natPolicies"`
	VPN                    VPNSettings            `json:"vpn"`
	Objects                Objects                `json:"objects"`
	SecurityProfiles       []SecurityProfile      `json:"securityProfiles,omitempty"`
	SecurityProfilesDetail SecurityProfilesDetail `json:"securityProfilesDetail,omitempty"`
	DHCPServers            []DHCPServer           `json:"dhcpServers,omitempty"`
	Routing                RoutingSettings        `json:"routing,omitempty"`
	HA                     HASettings             `json:"ha,omitempty"`
	Logging                LoggingSettings        `json:"logging,omitempty"`
	Coverage               []CoverageItem         `json:"coverage"`
	ParseErrors            []string               `json:"parseErrors,omitempty"`
	Encoding               string                 `json:"encoding,omitempty"`
}

type DeviceInfo struct {
	Hostname        string              `json:"hostname"`
	Model           string              `json:"model"`
	OSVersion       string              `json:"osVersion"`
	DeviceType      identify.DeviceType `json:"deviceType"`
	Role            string              `json:"role,omitempty"`         // firewall | switch
	VendorFamily    string              `json:"vendorFamily,omitempty"` // fortigate | paloalto | junos-ex | …
	OperationMode   string              `json:"operationMode"`
	VDOMEnabled     bool                `json:"vdomEnabled"`
	VDOMList        []string            `json:"vdomList"`
	OSVersionSource string              `json:"osVersionSource,omitempty"`
}

// VirtualContext is VDOM / vsys / routing-instance / VRF / default.
type VirtualContext struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type VLAN struct {
	Name        string `json:"name"`
	VLANID      string `json:"vlanId"`
	Description string `json:"description,omitempty"`
	L3Interface string `json:"l3Interface,omitempty"` // irb.x / VlanX
	Context     string `json:"context,omitempty"`
}

type LAG struct {
	Name        string   `json:"name"`
	Members     []string `json:"members,omitempty"`
	LACPEnabled bool     `json:"lacpEnabled"`
	Description string   `json:"description,omitempty"`
	Context     string   `json:"context,omitempty"`
}

type STPSettings struct {
	Protocol string   `json:"protocol,omitempty"` // rstp | mstp | vstp
	Enabled  bool     `json:"enabled"`
	Bridges  []string `json:"bridges,omitempty"`
	Context  string   `json:"context,omitempty"`
}

type SystemSettings struct {
	DNSPrimary   string   `json:"dnsPrimary"`
	DNSSecondary string   `json:"dnsSecondary"`
	Timezone     string   `json:"timezone"`
	HTTPSPort    string   `json:"httpsPort"`
	SSHPort      string   `json:"sshPort"`
	NTPServers   []string `json:"ntpServers"`
	AdminUsers   []Admin  `json:"adminUsers"`
}

type Admin struct {
	Username   string   `json:"username"`
	Profile    string   `json:"profile"`
	VDOM       string   `json:"vdom"`
	TrustHosts []string `json:"trustHosts"`
	Password   string   `json:"password,omitempty"`
}

type Interface struct {
	Name            string   `json:"name"`
	Alias           string   `json:"alias"`
	InterfaceType   string   `json:"interfaceType"` // physical | vlan | lag | irb | svi | l3 | …
	IPAddress       string   `json:"ipAddress"`
	VLANID          string   `json:"vlanId"`
	VDOM            string   `json:"vdom"`
	Context         string   `json:"context,omitempty"` // virtual context; empty ⇒ default
	Zone            string   `json:"zone"`
	Role            string   `json:"role"`
	AllowedAccess   []string `json:"allowedAccess"`
	Status          string   `json:"status"`
	Description     string   `json:"description"`
	ParentInterface string   `json:"parentInterface"`
	// L2 switchport (empty on pure L3 / firewall IF)
	SwitchportMode string   `json:"switchportMode,omitempty"` // access | trunk | ""
	AccessVLAN     string   `json:"accessVlan,omitempty"`
	TrunkVLANs     []string `json:"trunkVlans,omitempty"`
	NativeVLAN     string   `json:"nativeVlan,omitempty"`
	LAGParent      string   `json:"lagParent,omitempty"`
}

type Route struct {
	Name        string `json:"name"`
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Distance    string `json:"distance"`
	VDOM        string `json:"vdom"`
	Context     string `json:"context,omitempty"`
	RouteType   string `json:"routeType"`
	Enabled     bool   `json:"enabled"`
}

type Policy struct {
	PolicyID             string   `json:"policyId"`
	Name                 string   `json:"name"`
	SourceInterface      []string `json:"sourceInterface"`
	DestinationInterface []string `json:"destinationInterface"`
	SourceAddress        []string `json:"sourceAddress"`
	DestinationAddress   []string `json:"destinationAddress"`
	Service              []string `json:"service"`
	Action               string   `json:"action"`
	NATEnabled           bool     `json:"natEnabled"`
	LogEnabled           bool     `json:"logEnabled"`
	SecurityProfiles     []string `json:"securityProfiles,omitempty"`
	VDOM                 string   `json:"vdom"`
	Enabled              bool     `json:"enabled"`
	Description          string   `json:"description"`
	Schedule             string   `json:"schedule"`
}

type NATPolicy struct {
	Name                  string `json:"name"`
	NATType               string `json:"natType"` // vip | ippool | central-snat
	ExternalInterface     string `json:"externalInterface,omitempty"`
	ExternalIP            string `json:"externalIp,omitempty"`
	InternalIP            string `json:"internalIp,omitempty"`
	TranslatedSource      string `json:"translatedSource,omitempty"`
	OriginalSource        string `json:"originalSource,omitempty"`
	OriginalDestination   string `json:"originalDestination,omitempty"`
	TranslatedDestination string `json:"translatedDestination,omitempty"`
	Interface             string `json:"interface,omitempty"`
	PortForward           bool   `json:"portForward,omitempty"`
	OriginalPort          string `json:"originalPort,omitempty"`
	TranslatedPort        string `json:"translatedPort,omitempty"`
	PoolType              string `json:"poolType,omitempty"`
	VDOM                  string `json:"vdom"`
	Description           string `json:"description,omitempty"`
}

type VPNSettings struct {
	IPSecPhase1 []IPSecPhase1    `json:"ipsecPhase1"`
	IPSecPhase2 []IPSecPhase2    `json:"ipsecPhase2"`
	SSLVPN      []SSLVPNSettings `json:"sslVpn,omitempty"`
}

type SSLVPNSettings struct {
	Portal          string   `json:"portal,omitempty"`
	ListenPort      string   `json:"listenPort,omitempty"`
	ListenInterface string   `json:"listenInterface,omitempty"`
	TunnelIPPool    string   `json:"tunnelIpPool,omitempty"`
	UserGroups      []string `json:"userGroups,omitempty"`
	Mode            string   `json:"mode,omitempty"` // tunnel | web
	VDOM            string   `json:"vdom"`
}

type IPSecPhase1 struct {
	Name           string `json:"name"`
	RemoteGateway  string `json:"remoteGateway"`
	Interface      string `json:"interface"`
	IKEVersion     string `json:"ikeVersion"`
	Proposal       string `json:"proposal"`
	Encryption     string `json:"encryption"`
	Authentication string `json:"authentication"`
	DHGroup        string `json:"dhGroup"`
	Lifetime       string `json:"lifetime"`
	PSKConfigured  bool   `json:"pskConfigured"` // never export raw PSK
	VDOM           string `json:"vdom"`
}

type IPSecPhase2 struct {
	Name           string `json:"name"`
	Phase1Name     string `json:"phase1Name"`
	Proposal       string `json:"proposal"`
	Encryption     string `json:"encryption"`
	Authentication string `json:"authentication"`
	PFS            string `json:"pfs"`
	Lifetime       string `json:"lifetime"`
	LocalSubnet    string `json:"localSubnet"`
	RemoteSubnet   string `json:"remoteSubnet"`
	VDOM           string `json:"vdom"`
}

type Objects struct {
	Addresses     []AddressObject `json:"addresses"`
	AddressGroups []NamedMembers  `json:"addressGroups"`
	Services      []ServiceObject `json:"services"`
	ServiceGroups []NamedMembers  `json:"serviceGroups"`
}

type AddressObject struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	VDOM        string `json:"vdom"`
	Description string `json:"description,omitempty"`
}

type ServiceObject struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
	VDOM     string `json:"vdom"`
}

type NamedMembers struct {
	Name        string   `json:"name"`
	Members     []string `json:"members"`
	VDOM        string   `json:"vdom"`
	Description string   `json:"description,omitempty"`
}

// SecurityProfile is the inventory row (WallScribe SecurityProfile).
type SecurityProfile struct {
	Name        string `json:"name"`
	ProfileType string `json:"profileType"` // antivirus, webfilter, ips, ssl-inspection, app-control
	Enabled     bool   `json:"enabled"`
	VDOM        string `json:"vdom"`
	Description string `json:"description,omitempty"`
}

// SecurityProfilesDetail holds per-type profile bodies.
type SecurityProfilesDetail struct {
	Antivirus     []AntivirusProfile     `json:"antivirus,omitempty"`
	Webfilter     []WebFilterProfile     `json:"webfilter,omitempty"`
	AppControl    []AppControlProfile    `json:"appControl,omitempty"`
	IPS           []IPSProfile           `json:"ips,omitempty"`
	SSLInspection []SSLInspectionProfile `json:"sslInspection,omitempty"`
}

type AntivirusProfile struct {
	Name      string   `json:"name"`
	Enabled   bool     `json:"enabled"`
	ScanMode  string   `json:"scanMode,omitempty"`
	Protocols []string `json:"protocols,omitempty"`
	VDOM      string   `json:"vdom"`
}

type WebFilterProfile struct {
	Name       string   `json:"name"`
	Enabled    bool     `json:"enabled"`
	Categories []string `json:"categories,omitempty"`
	VDOM       string   `json:"vdom"`
}

type AppControlEntry struct {
	AppID    string `json:"appId"`
	AppName  string `json:"appName"`
	Category string `json:"category,omitempty"`
	Risk     string `json:"risk,omitempty"`
	Action   string `json:"action,omitempty"`
}

type AppControlProfile struct {
	Name         string            `json:"name"`
	Enabled      bool              `json:"enabled"`
	Categories   []string          `json:"categories,omitempty"`
	Applications []AppControlEntry `json:"applications,omitempty"`
	VDOM         string            `json:"vdom"`
}

type IPSProfile struct {
	Name       string   `json:"name"`
	Enabled    bool     `json:"enabled"`
	Signatures []string `json:"signatures,omitempty"`
	Action     string   `json:"action,omitempty"`
	VDOM       string   `json:"vdom"`
}

type SSLInspectionProfile struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
	VDOM    string `json:"vdom"`
}

type DHCPServer struct {
	Interface  string   `json:"interface"`
	StartIP    string   `json:"startIp,omitempty"`
	EndIP      string   `json:"endIp,omitempty"`
	Netmask    string   `json:"netmask,omitempty"`
	Gateway    string   `json:"gateway,omitempty"`
	DNSServers []string `json:"dnsServers,omitempty"`
	ExcludeIPs []string `json:"excludeIps,omitempty"`
	LeaseTime  string   `json:"leaseTime,omitempty"`
	VDOM       string   `json:"vdom"`
	Context    string   `json:"context,omitempty"`
}

// ContextOrDefault returns ctx or DefaultContext when empty.
func ContextOrDefault(ctx string) string {
	if ctx == "" {
		return DefaultContext
	}
	return ctx
}

type OSPFArea struct {
	AreaID         string   `json:"areaId"`
	AreaType       string   `json:"areaType,omitempty"`
	Authentication string   `json:"authentication,omitempty"`
	Networks       []string `json:"networks,omitempty"`
}

type OSPFInterface struct {
	Name           string `json:"name"`
	Interface      string `json:"interface"`
	Area           string `json:"area"`
	Cost           string `json:"cost,omitempty"`
	Priority       string `json:"priority,omitempty"`
	HelloInterval  string `json:"helloInterval,omitempty"`
	DeadInterval   string `json:"deadInterval,omitempty"`
	NetworkType    string `json:"networkType,omitempty"`
	Authentication string `json:"authentication,omitempty"`
	Passive        bool   `json:"passive"`
}

type OSPFRedistribute struct {
	Protocol   string `json:"protocol"`
	Status     bool   `json:"status"`
	Metric     string `json:"metric,omitempty"`
	MetricType string `json:"metricType,omitempty"`
	RouteMap   string `json:"routeMap,omitempty"`
}

type OSPFSettings struct {
	RouterID                    string             `json:"routerId"`
	DefaultInformationOriginate bool               `json:"defaultInformationOriginate"`
	DefaultMetric               string             `json:"defaultMetric,omitempty"`
	Distance                    string             `json:"distance,omitempty"`
	Areas                       []OSPFArea         `json:"areas,omitempty"`
	Interfaces                  []OSPFInterface    `json:"interfaces,omitempty"`
	Redistributes               []OSPFRedistribute `json:"redistributes,omitempty"`
	PassiveInterfaces           []string           `json:"passiveInterfaces,omitempty"`
	VDOM                        string             `json:"vdom"`
	Context                     string             `json:"context,omitempty"`
}

type BGPNeighbor struct {
	IP                  string `json:"ip"`
	RemoteAS            string `json:"remoteAs"`
	Description         string `json:"description,omitempty"`
	UpdateSource        string `json:"updateSource,omitempty"`
	EBGPMultihop        string `json:"ebgpMultihop,omitempty"`
	NextHopSelf         bool   `json:"nextHopSelf"`
	SoftReconfiguration bool   `json:"softReconfiguration"`
	RouteMapIn          string `json:"routeMapIn,omitempty"`
	RouteMapOut         string `json:"routeMapOut,omitempty"`
	Activate            bool   `json:"activate"`
	Shutdown            bool   `json:"shutdown"`
}

type BGPNetwork struct {
	Prefix   string `json:"prefix"`
	RouteMap string `json:"routeMap,omitempty"`
}

type BGPRedistribute struct {
	Protocol string `json:"protocol"`
	Status   bool   `json:"status"`
	RouteMap string `json:"routeMap,omitempty"`
}

type BGPSettings struct {
	ASNumber      string            `json:"asNumber"`
	RouterID      string            `json:"routerId"`
	Neighbors     []BGPNeighbor     `json:"neighbors,omitempty"`
	Networks      []BGPNetwork      `json:"networks,omitempty"`
	Redistributes []BGPRedistribute `json:"redistributes,omitempty"`
	VDOM          string            `json:"vdom"`
	Context       string            `json:"context,omitempty"`
}

type PolicyRoute struct {
	SeqNum       string `json:"seqNum"`
	Src          string `json:"src,omitempty"`
	SrcNegate    bool   `json:"srcNegate"`
	Dst          string `json:"dst,omitempty"`
	DstNegate    bool   `json:"dstNegate"`
	Protocol     string `json:"protocol,omitempty"`
	InputDevice  string `json:"inputDevice,omitempty"`
	OutputDevice string `json:"outputDevice,omitempty"`
	Gateway      string `json:"gateway,omitempty"`
	Action       string `json:"action,omitempty"`
	Status       bool   `json:"status"`
	Comments     string `json:"comments,omitempty"`
	VDOM         string `json:"vdom"`
	Context      string `json:"context,omitempty"`
}

type RoutingSettings struct {
	OSPF         []OSPFSettings `json:"ospf,omitempty"`
	OSPF6        []OSPFSettings `json:"ospf6,omitempty"`
	BGP          []BGPSettings  `json:"bgp,omitempty"`
	PolicyRoutes []PolicyRoute  `json:"policyRoutes,omitempty"`
}

type HAHeartbeatInterface struct {
	Interface string `json:"interface"`
	Priority  string `json:"priority,omitempty"`
}

type HAManagementInterface struct {
	ID        string `json:"id"`
	Interface string `json:"interface"`
	Dst       string `json:"dst,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
}

type HASettings struct {
	Mode                      string                  `json:"mode"` // standalone | active-passive | active-active
	GroupID                   string                  `json:"groupId,omitempty"`
	GroupName                 string                  `json:"groupName,omitempty"`
	Priority                  string                  `json:"priority,omitempty"`
	MonitorInterfaces         []string                `json:"monitorInterfaces,omitempty"`
	HeartbeatInterfaces       []string                `json:"heartbeatInterfaces,omitempty"`
	HeartbeatInterfacesDetail []HAHeartbeatInterface  `json:"heartbeatInterfacesDetail,omitempty"`
	HAMgmtInterfaces          []HAManagementInterface `json:"haMgmtInterfaces,omitempty"`
	HAMgmtStatus              bool                    `json:"haMgmtStatus"`
	Preempt                   bool                    `json:"preempt"`
	SessionPickup             bool                    `json:"sessionPickup"`
	HBInterval                string                  `json:"hbInterval,omitempty"`
	HBLostThreshold           string                  `json:"hbLostThreshold,omitempty"`
	Encryption                bool                    `json:"encryption"`
	Authentication            bool                    `json:"authentication"`
	PasswordConfigured        bool                    `json:"passwordConfigured"` // never store raw password
}

type SyslogServer struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	Facility string `json:"facility,omitempty"`
	Status   string `json:"status"`
	VDOM     string `json:"vdom,omitempty"`
}

type SNMPSettings struct {
	Enabled             bool     `json:"enabled"`
	CommunityConfigured bool     `json:"communityConfigured"` // never store raw community
	Hosts               []string `json:"hosts,omitempty"`
	TrapHosts           []string `json:"trapHosts,omitempty"`
	Version             string   `json:"version,omitempty"`
	Username            string   `json:"username,omitempty"`
}

type LoggingSettings struct {
	SyslogServers       []SyslogServer `json:"syslogServers,omitempty"`
	SNMP                []SNMPSettings `json:"snmp,omitempty"`
	FortiAnalyzerServer string         `json:"fortianalyzerServer,omitempty"`
	FortiAnalyzerStatus string         `json:"fortianalyzerStatus,omitempty"`
}

type CoverageItem struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Note   string `json:"note,omitempty"`
}
