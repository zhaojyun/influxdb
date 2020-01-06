package pkger

import "github.com/influxdata/influxdb"

// DashboardToResource is an exported call to dashboardToResource used by the
// chronograf-migrator. It is in a separate file so that it may be maintained distinctly
// from the internal one.
func DashboardToResource(d influxdb.Dashboard, name string) Resource {
	return dashboardToResource(d, name)
}

// VariableToResource is an exported call to variableToResource used by the
// chronograf-migrator. It is in a separate file so that it may be maintained distinctly
// from the internal one.
func VariableToResource(v influxdb.Variable, name string) Resource {
	return variableToResource(v, name)
}
