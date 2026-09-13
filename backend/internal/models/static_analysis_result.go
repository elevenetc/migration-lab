package models

import "encoding/json"

// StaticAnalysisResult is what static analysis returns: the parsed timeline and
// the warnings raised over it. Its runtime counterpart is
// RuntimeAnalysisResult.
type StaticAnalysisResult struct {
	Migrations []*Migration `json:"migrations"`
	Warnings   []Warning    `json:"warnings"`
}

func (r *StaticAnalysisResult) MarshalJSON() ([]byte, error) {
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

	return json.Marshal(map[string]any{
		"migrations": migrations,
		"warnings":   warnings,
	})
}
