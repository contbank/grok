package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
)

func TestOnlyLetters(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"1234567890qwertyuiop", "qwertyuiop"},
		{"(11) 99999-9999", ""},
		{"$#@1%^&*(A)21$", "A"},
		{"@xpto#", "xpto"},
	}

	for _, item := range items {
		result := grok.OnlyLetters(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestOnlyDigits(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"1234567890qwertyuiop", "1234567890"},
		{"(11) 99999-9999", "11999999999"},
		{"$#@1%^&*()21$", "121"},
		{"@xpto#", ""},
	}

	for _, item := range items {
		result := grok.OnlyDigits(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestOnlyLettersOrDigits(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"1234567890qwertyuiop", "1234567890qwertyuiop"},
		{"(11) 99999-9999", "11999999999"},
		{"$#@1%^&*(A)21$", "1A21"},
		{"@xpto#", "xpto"},
	}

	for _, item := range items {
		result := grok.OnlyLettersOrDigits(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestMaskEmail(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"email@email.com", "*****@email.com"},
		{"usertesteemail123213@email.com", "********************@email.com"},
		{"user", "****"},
	}

	for _, item := range items {
		result := grok.MaskEmail(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestMaskCellphone(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"76989527657", "76*****7657"},
		{"58998758620", "58*****8620"},
	}

	for _, item := range items {
		result := grok.MaskCellphone(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestGeneratorCellphone(t *testing.T) {
	phone := grok.GeneratorCellphone()
	assert.Equal(t, 11, len(phone))
}
