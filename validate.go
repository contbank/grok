package grok

import (
	"github.com/Nhanderu/brdoc"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"strconv"
	"strings"
)

var (
	// Validator ...
	Validator = NewValidator()
)

// NewValidator ...
func NewValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(JSONTagName)

	validate.RegisterValidation("objectid", ObjectID)
	validate.RegisterValidation("cnpj", CNPJ)
	validate.RegisterValidation("cpf", CPF)
	validate.RegisterValidation("cnpjcpf", CNPJCPF)
	validate.RegisterValidation("phone", Phone(false))
	validate.RegisterValidation("cellphone", Phone(true))
	validate.RegisterValidation("phonecellphone", PhoneOrCellphone())
	validate.RegisterValidation("fullname", FullName)
	validate.RegisterValidation("validdatetime", ValidDatetime)

	return validate
}

//JSONTagName ...
func JSONTagName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
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

//CPF ...
func CPF(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		s := strings.Replace(field.String(), ".", "", -1)
		s = strings.Replace(s, "-", "", -1)
		return brdoc.IsCPF(s)
	default:
		return false
	}
}

//CNPJCPF ...
func CNPJCPF(fl validator.FieldLevel) bool {
	result := CNPJ(fl)

	if !result {
		result = CPF(fl)
	}

	return result
}

// Phone ...
func Phone(isCellphone bool) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		if field.Kind() != reflect.String {
			return false
		}

		phone := field.String()

		if !IsOnlyDigits(phone) {
			return false
		}

		if (isCellphone && len(phone) != 11) || (!isCellphone && len(phone) != 10) {
			return false
		}

		if isCellphone && phone[2] != '9' {
			return false
		}

		if ddd, _ := strconv.Atoi(phone[:2]); ddd < 11 || ddd > 99 {
			return false
		}

		return true
	}
}

// PhoneOrCellphone ...
func PhoneOrCellphone() func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		if field.Kind() != reflect.String {
			return false
		}

		phone := field.String()
		phone = OnlyDigits(phone)

		if len(phone) != 11 && len(phone) != 10 {
			return false
		}

		if (len(phone) == 11) && phone[2] != '9' {
			return false
		}

		if ddd, _ := strconv.Atoi(phone[:2]); ddd < 11 || ddd > 99 {
			return false
		}

		return true
	}
}

// FullName ...
func FullName(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		name := strings.TrimSpace(field.String())
		parts := strings.Split(name, " ")

		if len(parts) < 2 {
			return false
		}

		for _, p := range parts {
			if len(p) < 1 || HasDigit(p) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// ValidDatetime ...
func ValidDatetime(fl validator.FieldLevel) bool {
	if fl.Field().IsZero() {
		return false
	}
	return true
}