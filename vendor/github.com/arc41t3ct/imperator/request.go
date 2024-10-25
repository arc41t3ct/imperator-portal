package imperator

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (i *Imperator) RequestReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(data); err != nil {
		return err
	}
	err := decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body can only have a single json value")
	}
	return nil
}
