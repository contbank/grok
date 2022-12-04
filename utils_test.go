package grok_test

import (
	"context"
	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestOnlyLettersOrSpaces(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"1234567890qwertyuiop", "qwertyuiop"},
		{"1234567890qwert yuiop", "qwert yuiop"},
		{"teste do espaço", "teste do espaço"},
		{"teste de pontuação. no meio. da frase.", "teste de pontuação no meio da frase"},
		{"(11) 99999-9999", " "},
		{"$#@1%^&*(A)21$", "A"},
		{"@xpto#", "xpto"},
	}

	for _, item := range items {
		result := grok.OnlyLettersOrSpaces(item.input)
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

func TestHasDigit(t *testing.T) {
	var items = []struct {
		input    string
		expected bool
	}{
		{"1234567890qwertyuiop", true},
		{"(11) 99999-9999", true},
		{"$#@1%^&*(A)21$", true},
		{"C4rro", true},
		{"@xpto#", false},
		{"teste", false},
	}

	for _, item := range items {
		result := grok.HasDigit(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestMaskCPF(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"47526039856", "475******56"},
		{"78764134040", "787******40"},
		{"98760212063", "987******63"},
	}

	for _, item := range items {
		result := grok.MaskCPF(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestMaskCNPJ(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"79571729000178", "795*********78"},
		{"30911104000119", "309*********19"},
		{"18556347000180", "185*********80"},
	}

	for _, item := range items {
		result := grok.MaskCNPJ(item.input)
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

func TestGeneratorDigitableLine(t *testing.T) {
	assert.Greater(t, len(grok.GeneratorDigitableLine()), 0)
}

func TestRemoveSpecialCharacters(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"Remoção de acentuação", "Remocao de acentuacao"},
		{"Remoção de acentuação e exclamação!", "Remocao de acentuacao e exclamacao"},
		{"áàéèúùâêíóò", "aaeeuuaeioo"},
		{"(11) 99999-9999", "11 999999999"},
		{"$#@1%^&*(A)21$#", "1A21"},
		{"@xpto#", "xpto"},
	}

	for _, item := range items {
		result := grok.RemoveSpecialCharacters(item.input)
		assert.Equal(t, item.expected, result)
	}
}

func TestShortenString(t *testing.T) {
	var items = []struct {
		input    string
		length   int
		expected string
	}{
		{"Endereco", 4, "Ende"},
		{"Rua Teste", 4, "Rua"},
		{"Rua Teste", 5, "Rua T"},
		{" Rua Teste", 5, "Rua"},
	}

	for _, item := range items {
		result := grok.ShortenString(item.input, item.length)
		assert.Equal(t, item.expected, result)
	}
}

func TestZipCode(t *testing.T) {
	var items = []struct {
		input    string
		expected string
	}{
		{"01301100", "01301100"},
		{"1301100", "01301100"},
		{"91370-170", "91370170"},
		{"01311930", "01311930"},
		{"01311-930", "01311930"},
		{"1311930", "01311930"},
	}

	for _, item := range items {
		result := grok.ZipCode(item.input)
		assert.Equal(t, item.expected, result)
	}
}

// TestFormatCurrencyToString ...
func TestFormatCurrencyToString(t *testing.T) {
	assert.Equal(t, "R$ 0,01", grok.FormatCurrencyToString(0.01))
	assert.Equal(t, "R$ 0,91", grok.FormatCurrencyToString(0.91))
	assert.Equal(t, "R$ 1,00", grok.FormatCurrencyToString(1.00))
	assert.Equal(t, "R$ 1,01", grok.FormatCurrencyToString(1.01))
	assert.Equal(t, "R$ 5,06", grok.FormatCurrencyToString(5.06))
	assert.Equal(t, "R$ 51,06", grok.FormatCurrencyToString(51.06))
	assert.Equal(t, "R$ 576,06", grok.FormatCurrencyToString(576.06))
	assert.Equal(t, "R$ 8.576,06", grok.FormatCurrencyToString(8576.06))
	assert.Equal(t, "R$ 576,06", grok.FormatCurrencyToString(0576.06))
	assert.Equal(t, "R$ 3.576,00", grok.FormatCurrencyToString(3576))
	assert.Equal(t, "R$ 23.576,00", grok.FormatCurrencyToString(23576))
	assert.Equal(t, "R$ 197.576,06", grok.FormatCurrencyToString(197576.06))
	assert.Equal(t, "R$ 1.000.576,06", grok.FormatCurrencyToString(1000576.06))
}

// TestFormatDateTimeToString_ONLY_DATE ...
func TestFormatDateTimeToString_ONLY_DATE(t *testing.T) {
	datetime := time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "05/12/2022", grok.FormatDateTimeToString(datetime, grok.ONLY_DATE))
	datetime = time.Date(2022, 02, 01, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022", grok.FormatDateTimeToString(datetime, grok.ONLY_DATE))
	datetime = time.Date(2022, 12, 13, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "13/12/2022", grok.FormatDateTimeToString(datetime, grok.ONLY_DATE))
	datetime = time.Date(2022, 01, 27, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "27/01/2022", grok.FormatDateTimeToString(datetime, grok.ONLY_DATE))
}

// TestFormatDateTimeToString_ONLY_TIME ...
func TestFormatDateTimeToString_ONLY_TIME(t *testing.T) {
	datetime := time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "22:05:01", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME))
	datetime = time.Date(2022, 02, 01, 2, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "02:42:39", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME))
	datetime = time.Date(2022, 02, 01, 12, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "12:42:39", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME))
	datetime = time.Date(2022, 02, 01, 24, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "00:42:39", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME))
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, "21:00:00", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME))
}

// TestFormatDateTimeToString_ONLY_TIME_EXTENSION ...
func TestFormatDateTimeToString_ONLY_TIME_EXTENSION(t *testing.T) {
	datetime := time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "22h05min", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 2, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "02h42min", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 12, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "12h42min", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 24, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "00h42min", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, "21h00min", grok.FormatDateTimeToString(datetime, grok.ONLY_TIME_EXTENSION))
}

// TestFormatDateTimeToString_DATETIME ...
func TestFormatDateTimeToString_DATETIME(t *testing.T) {
	datetime := time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "05/12/2022 22:05:01", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 02, 01, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 22:05:01", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 12, 13, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "13/12/2022 22:05:01", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 01, 27, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "27/01/2022 22:05:01", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "05/12/2022 22:05:01", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 02, 01, 2, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 02:42:39", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 02, 01, 12, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 12:42:39", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 02, 01, 24, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "02/02/2022 00:42:39", grok.FormatDateTimeToString(datetime, grok.DATETIME))
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 21:00:00", grok.FormatDateTimeToString(datetime, grok.DATETIME))
}

// TestFormatDateTimeToString_DATETIME_EXTENSION ...
func TestFormatDateTimeToString_DATETIME_EXTENSION(t *testing.T) {
	datetime := time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "05/12/2022 às 22h05min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 às 22h05min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 12, 13, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "13/12/2022 às 22h05min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 01, 27, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "27/01/2022 às 22h05min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 12, 05, 22, 05, 01, 99999999, time.UTC)
	assert.Equal(t, "05/12/2022 às 22h05min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 2, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 às 02h42min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 12, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 às 12h42min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 24, 42, 39, 99999999, time.UTC)
	assert.Equal(t, "02/02/2022 às 00h42min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, "01/02/2022 às 21h00min", grok.FormatDateTimeToString(datetime, grok.DATETIME_EXTENSION))
}

// TestIsValidDatetime ...
func TestIsValidDatetime(t *testing.T) {
	datetime := time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	assert.True(t, grok.IsValidDatetime(&datetime))

	assert.False(t, grok.IsValidDatetime(nil))

	datetime = time.Now()
	assert.True(t, grok.IsValidDatetime(&datetime))
}

// TestChangeDate ...
func TestChangeDate(t *testing.T) {
	// +1 day
	datetime := time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	expected := time.Date(2022, 02, 02, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, &expected, grok.ChangeDate(datetime, 24))

	// +10 days
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	expected = time.Date(2022, 02, 11, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, &expected, grok.ChangeDate(datetime, 24*10))

	// +90 days
	datetime = time.Date(2022, 12, 01, 21, 00, 00, 99999999, time.UTC)
	expected = time.Date(2023, 03, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, &expected, grok.ChangeDate(datetime, 24*90))

	// -1 day
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	expected = time.Date(2022, 01, 31, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, &expected, grok.ChangeDate(datetime, -24))

	// -10 days
	datetime = time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	expected = time.Date(2022, 01, 22, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, &expected, grok.ChangeDate(datetime, -24*10))
}

// TestOnlyDate ...
func TestOnlyDate(t *testing.T) {
	assert.Nil(t, grok.OnlyDate(nil))

	datetime := time.Date(2022, 02, 01, 21, 00, 00, 99999999, time.UTC)
	result := grok.OnlyDate(&datetime)
	assert.Equal(t, 2022, result.Year())
	assert.Equal(t, time.Month(02), result.Month())
	assert.Equal(t, 01, result.Day())
	assert.Equal(t, 00, result.Hour())
	assert.Equal(t, 00, result.Minute())
	assert.Equal(t, 00, result.Second())
	assert.Equal(t, 00, result.Nanosecond())

	datetime = time.Date(2022, 12, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, time.Month(12), grok.OnlyDate(&datetime).Month())

	datetime = time.Date(2022, 6, 01, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, time.Month(6), grok.OnlyDate(&datetime).Month())

	datetime = time.Date(2022, 12, 19, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, 19, grok.OnlyDate(&datetime).Day())

	datetime = time.Date(2022, 12, 31, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, 31, grok.OnlyDate(&datetime).Day())

	datetime = time.Date(2022, 11, 30, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, 30, grok.OnlyDate(&datetime).Day())

	datetime = time.Date(2022, 6, 4, 21, 00, 00, 99999999, time.UTC)
	assert.Equal(t, 4, grok.OnlyDate(&datetime).Day())
}

// TestWorkerRequestID ...
func TestWorkerRequestID(t *testing.T) {
	ctx := grok.GenerateNewWorkerRequestID(context.Background())
	assert.NotNil(t, grok.GetWorkerRequestID(ctx))
}

// TestRequestID ...
func TestRequestID(t *testing.T) {
	ctx := grok.GenerateNewWorkerRequestID(context.Background())
	ctx = grok.GenerateNewRequestID(ctx)
	assert.NotNil(t, grok.GetWorkerRequestID(ctx))
	assert.NotNil(t, grok.GetRequestID(ctx))
	assert.NotEqual(t, grok.GetRequestID(ctx), grok.GetWorkerRequestID(ctx))
}

// TestToTitle ...
func TestToTitle(t *testing.T) {
	assert.Equal(t, "TESTE", grok.ToTitle("teste"))
	assert.Equal(t, "TESTE", grok.ToTitle("Teste"))
	assert.Equal(t, "NOME DA PESSOA", grok.ToTitle("nome da pessoa"))
	assert.Equal(t, "NOME DA PESSOA", grok.ToTitle("nome DA pessoa"))
	assert.Equal(t, "NOME DA PESSOA", grok.ToTitle("Nome da pessoa"))
	assert.Equal(t, "NOME DA PESSOA", grok.ToTitle("Nome da Pessoa"))
	assert.Equal(t, "NOME DA PESSOA DO CPF 1234", grok.ToTitle("Nome da Pessoa do CPF 1234"))
}
