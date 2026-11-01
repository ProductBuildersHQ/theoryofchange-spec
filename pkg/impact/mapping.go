package impact

// CSVMapping tells the importer how the columns of a spreadsheet map onto
// observations. It is read from YAML or JSON.
type CSVMapping struct {
	Schema string `json:"$schema,omitempty"`
	// ModelID is the impact model the spreadsheet measures.
	ModelID string     `json:"modelId" jsonschema:"required"`
	Columns CSVColumns `json:"columns" jsonschema:"required"`
	// IndicatorAliases map spreadsheet labels (a column heading's cell
	// values) to indicator IDs. Matching ignores case and surrounding space.
	IndicatorAliases map[string]string `json:"indicatorAliases,omitempty"`
	// MetricAliases map labels to metric IDs, for sheets that report a
	// metric by pathway and leave the importer to choose the indicator.
	MetricAliases map[string]string `json:"metricAliases,omitempty"`
	// DateFormat is a Go time layout; the default is 2006-01-02.
	DateFormat string `json:"dateFormat,omitempty"`
	// Period labels the observation set, for example 2026-Q4.
	Period string `json:"period,omitempty"`
}

// CSVColumns names the spreadsheet columns by their headings. Only Date and
// one of Indicator or IndicatorFixed are required, and a value comes from
// Value or Level.
type CSVColumns struct {
	// Indicator is the column holding an indicator ID, an alias, or (with
	// Pathway) a metric ID or alias.
	Indicator string `json:"indicator,omitempty"`
	// IndicatorFixed is the one indicator a whole sheet is about.
	IndicatorFixed string `json:"indicatorFixed,omitempty"`
	Date           string `json:"date" jsonschema:"required"`
	Value          string `json:"value,omitempty"`
	// Level is the column holding a qualitative level.
	Level string `json:"level,omitempty"`
	// Pathway picks the indicator when Indicator names a metric used in
	// several pathways.
	Pathway string `json:"pathway,omitempty"`
	Quality string `json:"quality,omitempty"`
	Source  string `json:"source,omitempty"`
	Note    string `json:"note,omitempty"`
}
