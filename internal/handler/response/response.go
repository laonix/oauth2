package response

import "encoding/json"

type Error struct {
	Error string `json:"error"`
}

func NewErrorBody(message string) ([]byte, error) {
	resp := Error{
		Error: message,
	}

	return json.Marshal(resp)
}
