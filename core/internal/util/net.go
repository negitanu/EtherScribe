package util

import (
	"net"
	"strconv"
	"strings"
)

// IPToCIDR ports WallScribe parsers.utils.ip_to_cidr.
func IPToCIDR(ipSubnet string) string {
	ipSubnet = strings.TrimSpace(ipSubnet)
	if ipSubnet == "" {
		return ""
	}
	if strings.Contains(ipSubnet, "/") {
		return ipSubnet
	}
	parts := strings.Fields(ipSubnet)
	if len(parts) != 2 {
		return ipSubnet
	}
	ip, mask := parts[0], parts[1]
	if n, err := strconv.Atoi(mask); err == nil {
		if strings.Contains(ip, ":") {
			if n >= 0 && n <= 128 {
				return ip + "/" + mask
			}
		} else if n >= 0 && n <= 32 {
			return ip + "/" + mask
		}
	}
	parsedIP := net.ParseIP(ip)
	parsedMask := net.ParseIP(mask)
	if parsedIP != nil && parsedMask != nil {
		if m4 := parsedMask.To4(); m4 != nil {
			ones, _ := net.IPMask(m4).Size()
			return ip + "/" + strconv.Itoa(ones)
		}
	}
	return ipSubnet
}

// ToList ports parsers.utils.to_list for string-or-list values stored as []string already.
func ToList(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []string:
		return t
	default:
		return nil
	}
}
