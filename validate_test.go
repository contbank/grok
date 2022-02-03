package grok_test

import (
	"fmt"
	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestFullName(t *testing.T) {
	var names = []struct {
		input    string
		expected bool
	}{
		{"João da Silva", true},
		{"Matheus Santos", true},
		{"Maria Souza", true},
		{"Ana Maria", true},
		{"manuella melo e cysne", true},
		{"Br a sil", true},
		{"A B C", true},

		{"     ", false},
		{"M4ria da S1lva", false},
		{"Marcos Santos2", false},
	}

	validate := grok.NewValidator()

	for _, item := range names {
		err := validate.Var(item.input, "fullname")
		assert.Equal(t, item.expected, err == nil, fmt.Sprintf("provided name: %s", item.input))
	}
}