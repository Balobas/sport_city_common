package httpUtils

import (
	"encoding/json"
	"net/http"
)

func DecodeJsonRequest(r *http.Request, dest interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return err
	}
	return nil
}
