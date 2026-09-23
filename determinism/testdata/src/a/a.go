package a

import "encoding/json/v2"

type config struct {
	Rules map[string]string `json:"rules,omitempty"`
}

// badBareMarshal reproduces the original SDK SaveJSON bug class: a bare
// json.Marshal over a map-bearing value.
func badBareMarshal(c config) ([]byte, error) {
	return json.Marshal(c) // want `json.Marshal without an explicit json.Deterministic option`
}

// goodDeterministic is the fixed shape SaveJSON has carried since v0.3.1.
func goodDeterministic(c config) ([]byte, error) {
	return json.Marshal(c, json.Deterministic(true))
}

// deliberateOptOut documents intent: unstable bytes are acceptable here.
func deliberateOptOut(c config) ([]byte, error) {
	return json.Marshal(c, json.Deterministic(false))
}

// goodSpread passes an options slice; opaque spreads are not verifiable
// statically and must not be flagged (conservative, no false positive).
func goodSpread(c config, opts []json.Options) ([]byte, error) {
	return json.Marshal(c, opts...)
}

// goodOtherPackageMarshal proves the rule scopes to encoding/json/v2 only.
func goodOtherPackageMarshal(c config) ([]byte, error) {
	return marshalShim(c)
}

func marshalShim(v any) ([]byte, error) {
	return json.Marshal(v, json.Deterministic(true))
}
