package httpUtils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/balobas/sport_city_common/logger"
)

type ErrorResponse struct {
	Error error  `json:"error"`
	Msg   string `json:"message"`
}

func WriteErrorResponse(w http.ResponseWriter, status int, err error) {
	log := logger.Logger()

	SetContentTypeApplicationJson(w)

	w.WriteHeader(status)
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	encodeErr := json.NewEncoder(w).Encode(ErrorResponse{Error: err, Msg: errMsg})
	if encodeErr != nil {
		log.Error().Err(encodeErr).Msg("failed to encode error response")
		w.Write([]byte(encodeErr.Error()))
		return
	}
}

func SetContentTypeApplicationJson(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func WriteResponseJson(w http.ResponseWriter, resp interface{}) {
	SetContentTypeApplicationJson(w)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log := logger.Logger()
		log.Error().Err(err).Msg("failed to encode response")

		WriteErrorResponse(w, http.StatusInternalServerError, errors.New("failed to encode response"))
		return
	}
}
