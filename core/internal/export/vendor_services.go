package export

import (
	"reflect"
	"strings"

	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/model"
)

func fortiServicesApplicable(cfg *model.Config) bool {
	switch cfg.DeviceInfo.DeviceType {
	case identify.PaloAlto, identify.JunosEX:
		return false
	}
	return true
}

// Project the schema as well as the values, so omitted JSON fields do not
// reappear as unrelated vendor columns in HTML/Excel. Unknown internal models
// retain the generic schema for backward compatibility.
func vendorServiceType(cfg *model.Config, key string, typ reflect.Type) reflect.Type {
	if key != "logging" || fortiServicesApplicable(cfg) {
		return typ
	}
	var fields []reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !strings.HasPrefix(field.Name, "FortiAnalyzer") {
			fields = append(fields, field)
		}
	}
	return reflect.StructOf(fields)
}

func dhcpRelayApplicable(cfg *model.Config) bool {
	switch cfg.DeviceInfo.DeviceType {
	case identify.FortiGate, identify.JunosEX:
		return false
	}
	return true
}
