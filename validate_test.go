package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
)

var validPhones = []string{
	"7689527657",
	"5898758620",
}

func TestPhone(t *testing.T) {
	validate := grok.NewValidator()

	for _, phone := range validPhones {
		err := validate.Var(phone, "phone")
		assert.NoError(t, err)
	}
}

var invalidPhones = []string{
	"+55 (11) 99999-9999",
	"(11) 99999-9999",
	"5511999999999",
	"11999999999",
	"11899999999",
	"10999999999",
	"XXXXXXXXXXX",
	"XYAS1199999999",
	"#!@#1199999999",
}

func TestPhoneInvalid(t *testing.T) {
	validate := grok.NewValidator()

	for _, phone := range invalidPhones {
		err := validate.Var(phone, "phone")
		assert.Error(t, err)
	}
}

var validCellphones = []string{
	"76989527657",
	"58998758620",
}

func TestCellphone(t *testing.T) {
	validate := grok.NewValidator()

	for _, phone := range validCellphones {
		err := validate.Var(phone, "cellphone")
		assert.NoError(t, err)
	}
}

var invalidCellphones = []string{
	"+55 (11) 99999-9999",
	"5511999999999",
	"999999999",
	"1199999999",
	"11899999999",
	"10999999999",
	"XXXXXXXXXXX",
	"XYAS11999999999",
	"#!@#11999999999",
}

func TestCellphoneInvalid(t *testing.T) {
	validate := grok.NewValidator()

	for _, phone := range invalidCellphones {
		err := validate.Var(phone, "cellphone")
		assert.Error(t, err)
	}
}
