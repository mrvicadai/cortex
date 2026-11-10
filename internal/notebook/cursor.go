package notebook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

type cursorEnvelope struct {
	Project   string          `json:"project"`
	Operation string          `json:"operation"`
	State     json.RawMessage `json:"state"`
}

func (s *Store) encodeCursor(project, operation string, state any) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(cursorEnvelope{Project: project, Operation: operation, State: data})
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	cursor := base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if len(cursor) > 8192 {
		return "", invalid("cursor capacity")
	}
	return cursor, nil
}

func (s *Store) decodeCursor(cursor, project, operation string, state any) error {
	if len(cursor) > 8192 {
		return invalid("cursor")
	}
	parts := strings.Split(cursor, ".")
	if len(parts) != 2 {
		return invalid("cursor")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return invalid("cursor")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return invalid("cursor")
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return invalid("cursor")
	}
	return decodeCursorState(payload, project, operation, state)
}

func decodeCursorState(payload []byte, project, operation string, state any) error {
	var envelope cursorEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return invalid("cursor")
	}
	if envelope.Project != project || envelope.Operation != operation {
		return invalid("cursor operation or project")
	}
	if err := json.Unmarshal(envelope.State, state); err != nil {
		return invalid("cursor state")
	}
	return nil
}
