package dbstats

type DatabaseStats struct {
	Tables []Table `json:"tables"`
}

func (d *DatabaseStats) NonNil() {
	for i := range d.Tables {
		d.Tables[i].NonNil()
	}
}

type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	ForeignKeys []ForeignKey `json:"foreignKeys"`
	Indexes     []Index      `json:"indexes"`
}

func (t *Table) NonNil() {
	if t.Columns == nil {
		t.Columns = []Column{}
	}

	if t.ForeignKeys == nil {
		t.ForeignKeys = []ForeignKey{}
	}

	if t.Indexes == nil {
		t.Indexes = []Index{}
	}
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
