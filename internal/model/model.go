package model

// Diagnostic is a tool-independent finding. Paths are repository-relative.
type Diagnostic struct {
	Tool        string `json:"tool"`
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Message     string `json:"message"`
	Fingerprint string `json:"fingerprint"`
	New         bool   `json:"new"`
}

type Check struct {
	Tool        string       `json:"tool"`
	Version     string       `json:"version"`
	Status      string       `json:"status"`
	Error       string       `json:"error,omitempty"`
	Analyzed    int          `json:"analyzed"`
	Total       int          `json:"total"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Result struct {
	SchemaVersion  int      `json:"schema_version"`
	ScoringVersion int      `json:"scoring_version"`
	Mode           string   `json:"mode"`
	Status         string   `json:"status"`
	Complete       bool     `json:"analysis_complete"`
	Grade          string   `json:"grade"`
	Score          *float64 `json:"score"`
	Files          int      `json:"files"`
	NewErrors      int      `json:"new_errors"`
	Checks         []Check  `json:"checks"`
}
