package grok

import (
	"reflect"
	"strings"

	"github.com/Nhanderu/brdoc"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// Validator ...
	Validator = NewValidator()
)

// NewValidator ...
func NewValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterValidation("objectid", ObjectID)
	validate.RegisterValidation("cnpj", CNPJ)

	return validate
}

//ObjectID ...
func ObjectID(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		if _, err := primitive.ObjectIDFromHex(field.String()); err != nil {
			return false
		}
		return true
	default:
		return false
	}
}

//CNPJ ...
func CNPJ(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		s := strings.Replace(field.String(), ".", "", -1)
		s = strings.Replace(s, "-", "", -1)
		s = strings.Replace(s, "/", "", -1)
		return brdoc.IsCNPJ(s)
	default:
		return false
	}
}
