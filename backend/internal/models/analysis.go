package models

import "encoding/json"

type OperationID struct {
	MigrationID string `json:"migrationId"`
	TableName   string `json:"tableName"`
}

type Warning interface {
	warningType() string
	GetOperationID() OperationID
}

type AccessExclusiveLock struct {
	OperationID OperationID `json:"operationId"`
	TableName   string      `json:"tableName"`
	Message     string      `json:"message"`
}

func (a AccessExclusiveLock) warningType() string         { return "ACCESS_EXCLUSIVE_LOCK" }
func (a AccessExclusiveLock) GetOperationID() OperationID { return a.OperationID }

type AnalysisResult struct {
	Migrations []*Migration `json:"migrations"`
	Warnings   []Warning    `json:"warnings"`
}

func (r *AnalysisResult) MarshalJSON() ([]byte, error) {
	warnings := make([]json.RawMessage, len(r.Warnings))
	for i, w := range r.Warnings {
		data, err := MarshalWarning(w)
		if err != nil {
			return nil, err
		}
		warnings[i] = data
	}

	migrations := make([]json.RawMessage, len(r.Migrations))
	for i, m := range r.Migrations {
		data, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		migrations[i] = data
	}

	return json.Marshal(map[string]interface{}{
		"migrations": migrations,
		"warnings":   warnings,
	})
}

func MarshalWarning(w Warning) ([]byte, error) {
	wrapper := map[string]interface{}{
		"type": w.warningType(),
	}

	data, err := json.Marshal(w)
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
