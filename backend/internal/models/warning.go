package models

import "maps"

import "encoding/json"

// Warning is what static analysis raises. Its runtime counterpart is
// RuntimeFinding, which carries the same fields so both stay renderable through
// one path.
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

// MarshalWarning flattens a warning into its fields plus the discriminating
// type, the way the frontend's tagged union reads it.
func MarshalWarning(w Warning) ([]byte, error) {
	wrapper := map[string]any{
		"type": w.warningType(),
	}

	data, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}

	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	maps.Copy(wrapper, fields)

	return json.Marshal(wrapper)
}
