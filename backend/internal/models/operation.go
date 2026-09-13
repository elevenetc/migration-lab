package models

import "encoding/json"

type Column struct {
	Name        string   `json:"name"`
	Type        SQLType  `json:"type"`
	Constraints []string `json:"constraints"`
	// Deparsed DEFAULT expression, empty when the column has no default or the
	// expression could not be deparsed. Constraints carries the "DEFAULT" token
	// either way, so the two cases stay distinguishable.
	DefaultExpr string `json:"defaultExpr,omitempty"`
	// DefaultVolatile reports whether the DEFAULT expression contains a volatile
	// function, read from the expression tree rather than from DefaultExpr. A
	// volatile default forces a rewrite; a non-volatile one is stored once in the
	// catalog. Classification input only, so it stays out of the API.
	DefaultVolatile bool `json:"-"`
}

type Migration struct {
	ID         string      `json:"id"`
	Version    string      `json:"version"`
	Timestamp  int64       `json:"timestamp"`
	Statements []Statement `json:"statements"`
	SQL        string      `json:"-"`
}

type Statement struct {
	Index      int         `json:"index"` // position in Migration.Statements
	Kind       string      `json:"kind"`  // CREATE_TABLE | ALTER_TABLE | DROP_TABLE | RENAME
	SQL        string      `json:"sql"`
	Operations []Operation `json:"operations"`
}

type Operation interface {
	operationType() string
}

func (s Statement) MarshalJSON() ([]byte, error) {
	type Alias Statement
	ops := make([]json.RawMessage, len(s.Operations))
	for i, op := range s.Operations {
		data, err := MarshalOperation(op)
		if err != nil {
			return nil, err
		}
		ops[i] = data
	}
	return json.Marshal(&struct {
		Alias
		Operations       []json.RawMessage `json:"operations"`
		PerformanceClass PerformanceClass  `json:"performanceClass"`
	}{
		Alias:            (Alias)(s),
		Operations:       ops,
		PerformanceClass: WorstPerformanceClass(s.Operations),
	})
}

func MarshalOperation(op Operation) ([]byte, error) {
	wrapper := map[string]interface{}{
		"type":             op.operationType(),
		"performanceClass": ClassifyPerformance(op),
	}

	data, err := json.Marshal(op)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	for k, v := range fields {
		wrapper[k] = v
	}

	return json.Marshal(wrapper)
}

type CreateTable struct {
	TableName     string   `json:"tableName"`
	Columns       []Column `json:"columns"`
	IsPartitioned bool     `json:"isPartitioned"`
	PartitionOf   *string  `json:"partitionOf,omitempty"`
}

func (c CreateTable) operationType() string { return "CREATE_TABLE" }

type AddColumn struct {
	TableName string `json:"tableName"`
	Column    Column `json:"column"`
}

func (a AddColumn) operationType() string { return "ADD_COLUMN" }

type DropColumn struct {
	TableName  string `json:"tableName"`
	ColumnName string `json:"columnName"`
}

func (d DropColumn) operationType() string { return "DROP_COLUMN" }

type DropTable struct {
	TableName string `json:"tableName"`
}

func (d DropTable) operationType() string { return "DROP_TABLE" }

type AlterColumnType struct {
	TableName  string  `json:"tableName"`
	ColumnName string  `json:"columnName"`
	NewType    SQLType `json:"newType"`
	// PreviousType is the column's type before this operation, resolved by walking
	// the timeline; zero when no earlier migration declared it. Whether the change
	// rewrites the table depends on both types, so the classifier needs it.
	PreviousType SQLType `json:"-"`
}

func (a AlterColumnType) operationType() string { return "ALTER_COLUMN_TYPE" }

type SetNotNull struct {
	TableName  string `json:"tableName"`
	ColumnName string `json:"columnName"`
}

func (s SetNotNull) operationType() string { return "SET_NOT_NULL" }

type DropNotNull struct {
	TableName  string `json:"tableName"`
	ColumnName string `json:"columnName"`
}

func (d DropNotNull) operationType() string { return "DROP_NOT_NULL" }

type SetDefault struct {
	TableName    string `json:"tableName"`
	ColumnName   string `json:"columnName"`
	DefaultValue string `json:"defaultValue"`
}

func (s SetDefault) operationType() string { return "SET_DEFAULT" }

type DropDefault struct {
	TableName  string `json:"tableName"`
	ColumnName string `json:"columnName"`
}

func (d DropDefault) operationType() string { return "DROP_DEFAULT" }

type RenameTable struct {
	TableName    string `json:"tableName"`
	NewTableName string `json:"newTableName"`
}

func (r RenameTable) operationType() string { return "RENAME_TABLE" }

type RenameColumn struct {
	TableName     string `json:"tableName"`
	ColumnName    string `json:"columnName"`
	NewColumnName string `json:"newColumnName"`
}

func (r RenameColumn) operationType() string { return "RENAME_COLUMN" }

type AddConstraint struct {
	TableName      string `json:"tableName"`
	ConstraintName string `json:"constraintName"`
	ConstraintType string `json:"constraintType"`
	// NOT VALID: existing rows are not checked, so the constraint is added
	// without scanning the table.
	NotValid bool `json:"notValid"`
}

func (a AddConstraint) operationType() string { return "ADD_CONSTRAINT" }

type DropConstraint struct {
	TableName      string `json:"tableName"`
	ConstraintName string `json:"constraintName"`
}

func (d DropConstraint) operationType() string { return "DROP_CONSTRAINT" }
