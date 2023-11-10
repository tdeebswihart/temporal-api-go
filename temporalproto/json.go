package temporalproto

import "encoding/json"

var (
	objectOpen  = json.Delim('{')
	objectClose = json.Delim('}')
	arrayOpen   = json.Delim('[')
	arrayClose  = json.Delim(']')
)
