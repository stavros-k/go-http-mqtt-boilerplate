package dbstats

type DatabaseStats struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	ForeignKeys []ForeignKey `json:"foreignKeys,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
}

type Column struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	NotNull    bool    `json:"notNull"`
	Default    *string `json:"default,omitempty"`
	PrimaryKey bool    `json:"primaryKey"`
}

type ForeignKey struct {
	From  string `json:"from"`
	Table string `json:"table"`
	To    string `json:"to"`
}

type Index struct {
	Name    string   `json:"name"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
}
