package codecs

import "encoding/json"

type JSONCodec struct{}

func (JSONCodec) Marshal(data any) ([]byte, error)      { return json.Marshal(data) }
func (JSONCodec) Unmarshal(data []byte, dest any) error { return json.Unmarshal(data, dest) }
