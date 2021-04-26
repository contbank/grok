package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
)

var onlyDigitsItems = []struct {
	input    string
	expected string
}{
	{"1234567890qwertyuiop", "1234567890"},
	{"(11) 99999-9999", "11999999999"},
	{"$#@1%^&*()21$", "121"},
	{"@xpto#", ""},
}

func TestOnlyDigits(t *testing.T) {
	for _, item := range onlyDigitsItems {
		result := grok.OnlyDigits(item.input)
		assert.Equal(t, item.expected, result)
	}
}

var maskEmailItems = []struct {
	input    string
	expected string
}{
	{"email@email.com", "*****@email.com"},
	{"usertesteemail123213@email.com", "********************@email.com"},
	{"user", "****"},
}

func TestMaskEmail(t *testing.T) {
	for _, item := range maskEmailItems {
		result := grok.MaskEmail(item.input)
		assert.Equal(t, item.expected, result)
	}
}

var maskCellphoneItems = []struct {
	input    string
	expected string
}{
	{"76989527657", "76*****7657"},
	{"58998758620", "58*****8620"},
}

func TestMaskCellphone(t *testing.T) {
	for _, item := range maskCellphoneItems {
		result := grok.MaskCellphone(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestGeneratorCellphone(t *testing.T) {
	phone := grok.GeneratorCellphone()
	assert.Equal(t, 11, len(phone))
}
