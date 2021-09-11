package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/treemux"
)

func UnmarshalJSON(
	w http.ResponseWriter,
	req treemux.Request,
	dst interface{},
	maxBytes int64,
) error {
	req.Body = http.MaxBytesReader(w, req.Body, maxBytes)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
