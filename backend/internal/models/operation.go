package models

import "encoding/json"

type Column struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Constraints []string `json:"constraints"`
}

type Migration struct {
	ID         string      `json:"id"`
	Version    string      `json:"version"`
	Timestamp  int64       `json:"timestamp"`
	Operations []Operation `json:"operations"`
	SQL        string      `json:"-"`
}

type Operation interface {
	operationType() string
	GetMigrationID() string
}

type operationWrapper struct {
	Type string `json:"type"`
}

func (m *Migration) MarshalJSON() ([]byte, error) {
	type Alias Migration
	ops := make([]json.RawMessage, len(m.Operations))
	for i, op := range m.Operations {
		data, err := MarshalOperation(op)
		if err != nil {
			return nil, err
		}
		ops[i] = data
	}
	return json.Marshal(&struct {
		*Alias
		Operations []json.RawMessage `json:"operations"`
	}{
		Alias:      (*Alias)(m),
		Operations: ops,
	})
}

func MarshalOperation(op Operation) ([]byte, error) {
	wrapper := map[string]interface{}{
		"type": op.operationType(),
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
	MigrationID   string   `json:"migrationId"`
	TableName     string   `json:"tableName"`
	Columns       []Column `json:"columns"`
	IsPartitioned bool     `json:"isPartitioned"`
	PartitionOf   *string  `json:"partitionOf,omitempty"`
}

func (c CreateTable) operationType() string  { return "CREATE_TABLE" }
func (c CreateTable) GetMigrationID() string { return c.MigrationID }

type AddColumn struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
	Column      Column `json:"column"`
}

func (a AddColumn) operationType() string  { return "ADD_COLUMN" }
func (a AddColumn) GetMigrationID() string { return a.MigrationID }

type DropColumn struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
	ColumnName  string `json:"columnName"`
}

func (d DropColumn) operationType() string  { return "DROP_COLUMN" }
func (d DropColumn) GetMigrationID() string { return d.MigrationID }

type DropTable struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
}

func (d DropTable) operationType() string  { return "DROP_TABLE" }
func (d DropTable) GetMigrationID() string { return d.MigrationID }

type AlterColumnType struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
	ColumnName  string `json:"columnName"`
	NewType     string `json:"newType"`
}

func (a AlterColumnType) operationType() string  { return "ALTER_COLUMN_TYPE" }
func (a AlterColumnType) GetMigrationID() string { return a.MigrationID }

type SetNotNull struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
	ColumnName  string `json:"columnName"`
}

func (s SetNotNull) operationType() string  { return "SET_NOT_NULL" }
func (s SetNotNull) GetMigrationID() string { return s.MigrationID }

type DropNotNull struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
	ColumnName  string `json:"columnName"`
}

func (d DropNotNull) operationType() string  { return "DROP_NOT_NULL" }
func (d DropNotNull) GetMigrationID() string { return d.MigrationID }

type SetDefault struct {
	MigrationID  string `json:"migrationId"`
	TableName    string `json:"tableName"`
	ColumnName   string `json:"columnName"`
	DefaultValue string `json:"defaultValue"`
}

func (s SetDefault) operationType() string  { return "SET_DEFAULT" }
func (s SetDefault) GetMigrationID() string { return s.MigrationID }

type DropDefault struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
	ColumnName  string `json:"columnName"`
}

func (d DropDefault) operationType() string  { return "DROP_DEFAULT" }
func (d DropDefault) GetMigrationID() string { return d.MigrationID }

type RenameTable struct {
	MigrationID  string `json:"migrationId"`
	TableName    string `json:"tableName"`
	NewTableName string `json:"newTableName"`
}

func (r RenameTable) operationType() string  { return "RENAME_TABLE" }
func (r RenameTable) GetMigrationID() string { return r.MigrationID }

type RenameColumn struct {
	MigrationID   string `json:"migrationId"`
	TableName     string `json:"tableName"`
	ColumnName    string `json:"columnName"`
	NewColumnName string `json:"newColumnName"`
}

func (r RenameColumn) operationType() string  { return "RENAME_COLUMN" }
func (r RenameColumn) GetMigrationID() string { return r.MigrationID }

type AddConstraint struct {
	MigrationID    string `json:"migrationId"`
	TableName      string `json:"tableName"`
	ConstraintName string `json:"constraintName"`
	ConstraintType string `json:"constraintType"`
}

func (a AddConstraint) operationType() string  { return "ADD_CONSTRAINT" }
func (a AddConstraint) GetMigrationID() string { return a.MigrationID }

type DropConstraint struct {
	MigrationID    string `json:"migrationId"`
	TableName      string `json:"tableName"`
	ConstraintName string `json:"constraintName"`
}

func (d DropConstraint) operationType() string  { return "DROP_CONSTRAINT" }
func (d DropConstraint) GetMigrationID() string { return d.MigrationID }
