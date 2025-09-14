package httpUtils

import (
	"net/http"
	"strings"
)

const AccessJwtCookieName = "accessJwt"

func GetBearerToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) != 0 {
		return strings.Replace(bearer, "Bearer ", "", 1)
	}

	accessToken, _ := r.Cookie(AccessJwtCookieName)
	if accessToken == nil {
		return ""
	}
	return accessToken.Value
}
