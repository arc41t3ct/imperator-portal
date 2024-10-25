package imperator

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Validation struct {
	Errors map[string]string
}

func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validation) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// GetErrors return the error map
func (v *Validation) GetErrors() map[string]string {
	return v.Errors
}

// Has checks weather the given field is in the request data
func (v *Validation) Has(field string, r *http.Request) bool {
	fieldData := r.Form.Get(field)
	if fieldData != "" {
		return true
	}
	return false
}

// Required checks weather the request has the required fields
func (v *Validation) Required(r *http.Request, fields ...string) {
	for _, field := range fields {
		val := r.Form.Get(field)
		if strings.TrimSpace(val) == "" {
			v.AddError(field, "this field cannot be be blank")
		}
	}
}

// Check allows you to check a condition and set an error if it is not met
// Ex: validator.Check(len(someString) > 20, "password", "The password must be longer than 20 characters")
func (v *Validation) Check(condition bool, key, message string) {
	if !condition {
		v.AddError(key, message)
	}
}

func (v *Validation) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, "invalid e-mail address")
	}
}

func (v *Validation) IsInt(field, value string) {
	_, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "invalid integer")
	}
}

func (v *Validation) IsFloat(field, value string) {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.AddError(field, "invalid float")
	}
}

func (v *Validation) IsDateISO(field, value string) {
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		v.AddError(field, "invalid iso datetime")
	}
}

func (v *Validation) NoWhitespace(field, value string) {
	if govalidator.HasWhitespace(value) {
		v.AddError(field, "whitespace not permitted")
	}
}

func (i *Imperator) GetValidator() *Validation {
	return &Validation{
		Errors: make(map[string]string),
	}
}
