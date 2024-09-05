package kit

import (
	"encoding/json"

	"github.com/google/uuid"
)

// GenUUIDFromStruct generates a UUID from a struct
func GenUUIDFromStruct(structData interface{}) (string, error) {
	jsonBytes, err := json.Marshal(structData)
	if err != nil {
		return "", err
	}
	return GenUUIDFromBytes(jsonBytes), nil
}

// GenUUIDFromBytes generates a UUID from a byte slice
func GenUUIDFromBytes(bytes []byte) string {
	return uuid.NewSHA1(uuid.Nil, bytes).String()
}
