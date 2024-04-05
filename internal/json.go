package internal

import "encoding/json"

var (
	// Marshal is exported by vira/json package.
	Marshal = json.Marshal
	// Unmarshal is exported by vira/json package.
	Unmarshal = json.Unmarshal
	// MarshalIndent is exported by vira/json package.
	MarshalIndent = json.MarshalIndent
	// NewDecoder is exported by vira/json package.
	NewDecoder = json.NewDecoder
	// NewEncoder is exported by vira/json package.
	NewEncoder = json.NewEncoder
)
