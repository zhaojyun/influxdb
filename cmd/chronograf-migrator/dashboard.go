package main

import (
	"fmt"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/chronograf"
)

func convert1To2Cell(cell chronograf.DashboardCell) *influxdb.Cell {
	c := &influxdb.Cell{
		ID: 1,
		CellProperty: influxdb.CellProperty{
			X: cell.X,
			Y: cell.Y,
			W: cell.W,
			H: cell.H,
		},
	}

	v := influxdb.View{
		ViewContents: influxdb.ViewContents{
			Name: cell.Name,
		},
	}

	switch cell.Type {
	case "line":
		v.Properties = influxdb.XYViewProperties{
			Queries:    convertQueries(cell.Queries),
			Axes:       convertAxes(cell.Axes),
			Type:       "xy",
			Legend:     convertLegend(cell.Legend),
			Geom:       "line",
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			Position:   "overlaid",
		}
	case "line-stacked":
		v.Properties = influxdb.XYViewProperties{
			Queries:    convertQueries(cell.Queries),
			Axes:       convertAxes(cell.Axes),
			Type:       "xy",
			Legend:     convertLegend(cell.Legend),
			Geom:       "line", // TODO(desa): maybe this needs to be stacked?
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			Position:   "stacked",
		}
	case "line-stepplot":
		v.Properties = influxdb.XYViewProperties{
			Queries:    convertQueries(cell.Queries),
			Axes:       convertAxes(cell.Axes),
			Type:       "xy",
			Legend:     convertLegend(cell.Legend),
			Geom:       "step",
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			Position:   "overlaid",
		}
	case "bar":
		v.Properties = influxdb.XYViewProperties{
			Queries:    convertQueries(cell.Queries),
			Axes:       convertAxes(cell.Axes),
			Type:       "xy",
			Legend:     convertLegend(cell.Legend),
			Geom:       "bar",
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			Position:   "overlaid",
		}
	case "line-plus-single-stat":
		v.Properties = influxdb.LinePlusSingleStatProperties{
			Queries:    convertQueries(cell.Queries),
			Axes:       convertAxes(cell.Axes),
			Legend:     convertLegend(cell.Legend),
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			Position:   "overlaid",
		}
	case "single-stat":
		v.Properties = influxdb.SingleStatViewProperties{
			Queries:    convertQueries(cell.Queries),
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			// TODO(desa): what to do about ShowNoteWhenEmpty?
		}
	case "gauge":
		v.Properties = influxdb.GaugeViewProperties{
			Queries:    convertQueries(cell.Queries),
			ViewColors: convertColors(cell.CellColors),
			Note:       cell.Note,
			// TODO(desa): what to do about ShowNoteWhenEmpty?
		}
	case "table":
		v.Properties = influxdb.TableViewProperties{
			Queries:    convertQueries(cell.Queries),
			ViewColors: convertColors(cell.CellColors),
			//TableOptions
			//FieldOptions
			Note: cell.Note,
			// TODO(desa): what to do about ShowNoteWhenEmpty?
		}
	case "note":
		v.Properties = influxdb.MarkdownViewProperties{
			Note: cell.Note,
		}
	case "alerts", "news", "guide":
		// TODO(desa): these do not have 2.x equivalents
		v.Properties = influxdb.EmptyViewProperties{}
	default:
		v.Properties = influxdb.EmptyViewProperties{}
	}

	c.View = &v

	return c
}

func convert1To2Variable(t chronograf.Template) (influxdb.Variable, error) {
	v := influxdb.Variable{
		Description: t.Label,
		Name:        t.Var[1 : len(t.Var)-1], // trims `:` from variables prefix and suffix
	}

	switch t.Type {
	case "influxql", "databases", "fieldKeys", "tagKeys", "tagValues", "measurements":
		if t.Query == nil {
			return v, fmt.Errorf("expected template variable to have non-nil query")
		}
	}

	switch t.Type {
	case "influxql":
		v.Arguments = &influxdb.VariableArguments{
			Type: "query",
			Values: influxdb.VariableQueryValues{
				Query:    fmt.Sprintf("// %s", t.Query.Command),
				Language: "flux",
			},
		}
	case "databases":
		v.Arguments = &influxdb.VariableArguments{
			Type: "query",
			Values: influxdb.VariableQueryValues{
				Query:    fmt.Sprintf("// SHOW DATABASES %s", t.Query.DB),
				Language: "flux",
			},
		}
	case "fieldKeys":
		v.Arguments = &influxdb.VariableArguments{
			Type: "query",
			Values: influxdb.VariableQueryValues{
				Query:    fmt.Sprintf("// SHOW FIELD KEYS FOR %s", t.Query.Measurement),
				Language: "flux",
			},
		}
	case "tagKeys":
		v.Arguments = &influxdb.VariableArguments{
			Type: "query",
			Values: influxdb.VariableQueryValues{
				Query:    fmt.Sprintf("// SHOW TAG KEYS FOR %s", t.Query.Measurement),
				Language: "flux",
			},
		}
	case "tagValues":
		v.Arguments = &influxdb.VariableArguments{
			Type: "query",
			Values: influxdb.VariableQueryValues{
				Query:    fmt.Sprintf("// SHOW TAG VALUES FOR %s", t.Query.TagKey),
				Language: "flux",
			},
		}
	case "measurements":
		v.Arguments = &influxdb.VariableArguments{
			Type: "query",
			Values: influxdb.VariableQueryValues{
				Query:    fmt.Sprintf("// SHOW MEASUREMENTS ON %s", t.Query.DB),
				Language: "flux",
			},
		}
	case "csv", "constant", "text":
		values := influxdb.VariableConstantValues{}
		for _, val := range t.Values {
			values = append(values, val.Value)
		}
		v.Arguments = &influxdb.VariableArguments{
			Type:   "constant",
			Values: values,
		}
	case "map":
		values := influxdb.VariableMapValues{}
		for _, val := range t.Values {
			values[val.Key] = val.Value
		}
		v.Arguments = &influxdb.VariableArguments{
			Type:   "map",
			Values: values,
		}
	default:
		return v, fmt.Errorf("unknown variable type %s", t.Type)
	}

	return v, nil
}

func Convert1To2Dashboard(d1 chronograf.Dashboard) (influxdb.Dashboard, []influxdb.Variable, error) {
	cells := []*influxdb.Cell{}
	for _, cell := range d1.Cells {
		cells = append(cells, convert1To2Cell(cell))
	}

	d2 := influxdb.Dashboard{
		Name:  d1.Name,
		Cells: cells,
	}

	vars := []influxdb.Variable{}
	for _, template := range d1.Templates {
		v, err := convert1To2Variable(template)
		if err != nil {
			return influxdb.Dashboard{}, nil, err
		}

		vars = append(vars, v)
	}

	return d2, vars, nil
}

func convertAxes(a map[string]chronograf.Axis) map[string]influxdb.Axis {
	m := map[string]influxdb.Axis{}
	for k, v := range a {
		m[k] = influxdb.Axis{
			Bounds: v.Bounds,
			Label:  v.Label,
			Prefix: v.Prefix,
			Suffix: v.Suffix,
			Base:   v.Base,
			Scale:  v.Scale,
		}
	}

	if _, exists := m["x"]; !exists {
		m["x"] = influxdb.Axis{}
	}
	if _, exists := m["y"]; !exists {
		m["y"] = influxdb.Axis{}
	}

	return m
}

func convertLegend(l chronograf.Legend) influxdb.Legend {
	return influxdb.Legend{
		Type:        l.Type,
		Orientation: l.Orientation,
	}
}

func convertColors(cs []chronograf.CellColor) []influxdb.ViewColor {
	vs := []influxdb.ViewColor{}

	hasTextColor := false
	hasThresholdColor := false
	for _, c := range cs {
		if c.Type == "text" {
			hasTextColor = true
		}
		if c.Type == "threshold" {
			hasThresholdColor = true
		}

		v := influxdb.ViewColor{
			ID:   c.ID,
			Type: c.Type,
			Hex:  c.Hex,
			Name: c.Name,
		}
		vs = append(vs, v)
	}

	if !hasTextColor {
		vs = append(vs, influxdb.ViewColor{
			ID:    "base",
			Type:  "text",
			Hex:   "#00C9FF",
			Name:  "laser",
			Value: 0,
		})
	}

	if !hasThresholdColor {
		vs = append(vs, influxdb.ViewColor{
			ID:    "t",
			Type:  "threshold",
			Hex:   "#4591ED",
			Name:  "ocean",
			Value: 80,
		})
	}

	return vs
}

func convertQueries(qs []chronograf.DashboardQuery) []influxdb.DashboardQuery {
	ds := []influxdb.DashboardQuery{}
	for _, q := range qs {
		d := influxdb.DashboardQuery{
			// TODO(desa): possibly we should try to compile the query to flux that we can show the user.
			Text:     "// " + q.Command,
			EditMode: "advanced",
		}

		_ = q

		ds = append(ds, d)
	}

	if len(ds) == 0 {
		d := influxdb.DashboardQuery{
			// TODO(desa): possibly we should try to compile the query to flux that we can show the user.
			Text:     "// cell had no queries",
			EditMode: "advanced",
			BuilderConfig: influxdb.BuilderConfig{
				// TODO(desa): foo
				Buckets: []string{"bucket"},
			},
		}
		ds = append(ds, d)
	}

	return ds
}
