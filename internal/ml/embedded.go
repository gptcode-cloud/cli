package ml

import _ "embed"

//go:embed assets/complexity_model.json
var complexityModel []byte

//go:embed assets/intent_model.json
var intentModel []byte

var embeddedModels = map[string][]byte{
	"complexity":           complexityModel,
	"complexity_detection": complexityModel,
	"intent":               intentModel,
}
