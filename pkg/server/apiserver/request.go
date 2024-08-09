package apiserver

import (
	"encoding/json"
	"errors"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Request struct {
	Request *http.Request
	Session *Session
}

func (r *Request) Unmarshal(v any) error {
	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	err = util.Validator.Struct(v)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, fieldError := range validationErrors {
				return fmt.Errorf("request params validate failed on field '%s': %v", fieldError.Field(), fieldError)
			}
		}
		return fmt.Errorf("request params validate failed, %v", err)
	}
	return nil
}
