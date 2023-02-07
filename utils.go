package grok

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	ZIPCODE_LENGTH = 8
)

type DateTimeFormat string

const (
	// ONLY_DATE Format : DD/MM/YYYY
	ONLY_DATE DateTimeFormat = "ONLY_DATE"
	// ONLY_TIME Format : 15:50:02
	ONLY_TIME DateTimeFormat = "ONLY_TIME"
	// ONLY_TIME_EXTENSION Format : 15h50min
	ONLY_TIME_EXTENSION DateTimeFormat = "ONLY_TIME_EXTENSION"
	// DATETIME Format : DD/MM/YYYY às 15:50:02
	DATETIME DateTimeFormat = "DATETIME"
	// DATETIME_EXTENSION Format : DD/MM/YYYY às 15h50min
	DATETIME_EXTENSION DateTimeFormat = "DATETIME_EXTENSION"
)

const (
	// CNPJFormatPattern ...
	CNPJFormatPattern string = `([\d]{2})([\d]{3})([\d]{3})([\d]{4})([\d]{2})`
	// CPFFormatPattern ...
	CPFFormatPattern string = `([\d]{3})([\d]{3})([\d]{3})([\d]{2})`
	// BarcodePattern ...
	BarcodePattern string = `([\d]{5})([\d]{5})([\d]{5})([\d]{6})([\d]{5})([\d]{6})([\d]{1})([\d]{14})`
)

func random(n float64) float64 {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return math.Round(r.Float64() * n)
}

func mod(dividendo float64, divisor float64) float64 {
	return math.Round(dividendo - (math.Floor(dividendo/divisor) * divisor))
}

// GeneratorCPF ...
func GeneratorCPF() string {
	cpfString := ""

	rand.Seed(time.Now().UTC().UnixNano())
	cpf := rand.Perm(9)
	cpf = append(cpf, verify(cpf, len(cpf)))
	cpf = append(cpf, verify(cpf, len(cpf)))

	for _, c := range cpf {
		cpfString += strconv.Itoa(c)
	}

	return cpfString
}

func verify(data []int, n int) int {
	var total int

	for i := 0; i < n; i++ {
		total += data[i] * (n + 1 - i)
	}

	total = total % 11
	if total < 2 {
		return 0
	}
	return 11 - total
}

// GeneratorCNPJ ...
func GeneratorCNPJ() string {
	var n float64
	var n9 float64
	var n10 float64
	var n11 float64
	var n12 float64

	n = 9
	n9 = 0
	n10 = 0
	n11 = 0
	n12 = 1

	var n1 = random(n)
	var n2 = random(n)
	var n3 = random(n)
	var n4 = random(n)
	var n5 = random(n)
	var n6 = random(n)
	var n7 = random(n)
	var n8 = random(n)

	var d1 = n12*2 + n11*3 + n10*4 + n9*5 + n8*6 + n7*7 + n6*8 + n5*9 + n4*2 + n3*3 + n2*4 + n1*5
	d1 = 11 - (mod(d1, 11))
	if d1 >= 10 {
		d1 = 0
	}
	var d2 = d1*2 + n12*3 + n11*4 + n10*5 + n9*6 + n8*7 + n7*8 + n6*9 + n5*2 + n4*3 + n3*4 + n2*5 + n1*6
	d2 = 11 - (mod(d2, 11))
	if d2 >= 10 {
		d2 = 0
	}

	resultado := fmt.Sprintf("%d%d.%d%d%d.%d%d%d/%d%d%d%d-%d%d", int(n1), int(n2), int(n3), int(n4), int(n5), int(n6), int(n7), int(n8), int(n9), int(n10), int(n11), int(n12), int(d1), int(d2))

	return resultado
}

// ShortenString ...
func ShortenString(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return strings.TrimSpace(string(runes[:i]))
	}
	return s
}

// GeneratorIDBase ...
func GeneratorIDBase(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return strings.ToUpper(string(b))
}

// OnlyLetters ...
func OnlyLetters(value string) string {
	var newValue string

	for _, c := range value {
		if unicode.IsLetter(c) {
			newValue += string(c)
		}
	}

	return newValue
}

// OnlyLettersOrSpaces ...
func OnlyLettersOrSpaces(value string) string {
	var newValue string

	for _, c := range value {
		if unicode.IsLetter(c) || unicode.IsSpace(c) {
			newValue += string(c)
		}
	}

	return newValue
}

// OnlyDigits ...
func OnlyDigits(value string) string {

	var newValue string

	for _, c := range value {
		if unicode.IsDigit(c) {
			newValue += string(c)
		}
	}

	return newValue
}

// IsOnlyDigits ...
func IsOnlyDigits(value string) bool {
	for _, c := range value {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}

// HasDigit ...
func HasDigit(value string) bool {
	for _, c := range value {
		if unicode.IsDigit(c) {
			return true
		}
	}

	return false
}

// OnlyLettersOrDigits ...
func OnlyLettersOrDigits(value string) string {

	var newValue string

	for _, c := range value {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			newValue += string(c)
		}
	}

	return newValue
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// RemoveSpecialCharacters ...
func RemoveSpecialCharacters(value string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	newValue, _, _ := transform.String(t, value)

	var result string

	for _, c := range newValue {
		if unicode.IsLetter(c) || unicode.IsDigit(c) || unicode.IsSpace(c) {
			result += string(c)
		}
	}

	return result
}

// MaskCPF ...
func MaskCPF(value string) string {
	return value[:3] + strings.Repeat("*", 6) + value[9:]
}

// MaskCNPJ ...
func MaskCNPJ(value string) string {
	return value[:3] + strings.Repeat("*", 9) + value[12:]
}

// MaskEmail ...
func MaskEmail(value string) string {
	separator := "@"
	parts := strings.Split(value, separator)
	mask := strings.Repeat("*", len(parts[0]))

	if !strings.Contains(value, separator) {
		return mask
	}

	domain := parts[1]

	return mask + separator + domain
}

// MaskCellphone ...
func MaskCellphone(value string) string {
	ddd := value[:2]
	mask := strings.Repeat("*", 5)
	rest := value[7:]

	return ddd + mask + rest
}

// GeneratorCellphone ...
func GeneratorCellphone() string {
	phoneString := ""
	rand.Seed(time.Now().UTC().UnixNano())

	var dddArray = []string{
		"11", "12", "13", "14", "15", "16", "17", "18", "19",
		"21", "22", "24", "27", "28", "31", "32", "33", "34",
		"35", "37", "38", "41", "42", "43", "44", "45", "46",
		"47", "48", "49", "51", "53", "54", "55", "61", "62",
		"63", "64", "65", "66", "67", "68", "69", "71", "73",
		"74", "75", "77", "79", "81", "82", "83", "84", "85",
		"86", "87", "88", "89", "91", "92", "93", "94", "95",
		"96", "97", "98", "99",
	}

	phone := rand.Perm(8)
	for _, c := range phone {
		phoneString += strconv.Itoa(c)
	}

	phoneString = dddArray[rand.Intn(2)] + "9" + phoneString

	return phoneString
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// ToTitle ...
func ToTitle(value string) string {
	return strings.TrimSpace(strings.ToTitle(strings.ToLower(value)))
}

// GeneratorDigitableLine ...
func GeneratorDigitableLine() string {
	digitableString := ""
	rand.Seed(time.Now().UTC().UnixNano())

	digitable := rand.Perm(24)
	for _, c := range digitable {
		digitableString += strconv.Itoa(c)
	}

	return OnlyDigits(digitableString)
}

// ZipCode ...
func ZipCode(value string) string {
	aux := OnlyDigits(value)
	if len(aux) < ZIPCODE_LENGTH {
		for i := 1; len(aux) < 8; i++ {
			aux = "0" + aux
		}
	}
	return aux
}

// FormatCurrencyToString returns BRL format 99,99
func FormatCurrencyToString(amount float64, hasCurrencySymbol bool) string {
	lang := message.NewPrinter(language.BrazilianPortuguese)
	var result string

	if hasCurrencySymbol {
		result = lang.Sprintf("R$ %.2f", amount)
	} else {
		result = lang.Sprintf("%.2f", amount)
	}

	return result
}

// IsValidDatetime ...
func IsValidDatetime(datetime *time.Time) bool {
	if datetime == nil || datetime.IsZero() {
		return false
	}
	return true
}

// ChangeDate ...
func ChangeDate(datetime time.Time, hours int32) *time.Time {
	aux := datetime.Add(time.Duration(hours) * time.Hour)
	return &aux
}

// OnlyDate ...
func OnlyDate(datetime *time.Time) *time.Time {
	if datetime == nil {
		return nil
	}
	response := time.Date(datetime.Year(), datetime.Month(), datetime.Day(), 00, 00, 00, 00, time.UTC)
	return &response
}

// FormatDateTimeToString ...
func FormatDateTimeToString(datetime time.Time, format DateTimeFormat) string {
	switch format {
	case ONLY_DATE:
		return datetime.Format("02/01/2006")
	case ONLY_TIME:
		return datetime.Format("15:04:05")
	case ONLY_TIME_EXTENSION:
		return datetime.Format("15h04min")
	case DATETIME:
		return datetime.Format("02/01/2006 15:04:05")
	case DATETIME_EXTENSION:
		return datetime.Format("02/01/2006 às 15h04min")
	default:
		return datetime.Format("02/01/2006 15:04:05")
	}
}

// FormatBarcode format: 99999.99999 99999.999999 99999.999999 9 99999999999999
func FormatBarcode(str string) string {
	expr, err := regexp.Compile(BarcodePattern)

	if err != nil {
		return str
	}

	return expr.ReplaceAllString(str, "$1.$2 $3.$4 $5.$6 $7 $8")
}

// FormatCNPJ returns a formatted CNPJ:
// 99.999.999/9999-99
func FormatCNPJ(str string, returnType bool) string {
	var result string
	expr, err := regexp.Compile(CNPJFormatPattern)

	if err != nil {
		return str
	}

	if returnType {
		result = fmt.Sprintf("CNPJ - %s", expr.ReplaceAllString(str, "$1.$2.$3/$4-$5"))
	} else {
		result = expr.ReplaceAllString(str, "$1.$2.$3/$4-$5")
	}
	return result
}

// FormatCPF returns a formatted CPF:
// 999.999.999-99
func FormatCPF(str string, returnType bool) string {
	var result string
	expr, err := regexp.Compile(CPFFormatPattern)

	if err != nil {
		return str
	}

	if returnType {
		result = fmt.Sprintf("CPF - %s", expr.ReplaceAllString(str, "$1.$2.$3-$4"))
	} else {
		result = expr.ReplaceAllString(str, "$1.$2.$3-$4")
	}

	return result
}

// FormatCNPJOrCPF format a CPF or CNPJ
func FormatCNPJOrCPF(document string, returnType bool) string {
	if len(document) == 11 {
		return FormatCPF(document, returnType)
	} else if len(document) == 14 {
		return FormatCNPJ(document, returnType)
	}
	return document
}

// GenerateNewRequestID ...
func GenerateNewRequestID(ctx context.Context) context.Context {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, "Request-Id", requestID)
	return ctx
}

// GenerateNewWorkerRequestID ...
func GenerateNewWorkerRequestID(ctx context.Context) context.Context {
	workerRequestID := uuid.New().String()
	ctx = context.WithValue(ctx, "Worker-Request-Id", workerRequestID)
	return ctx
}

// GetRequestID ...
func GetRequestID(ctx context.Context) string {
	requestID, _ := ctx.Value("Request-Id").(string)
	return requestID
}

// GetWorkerRequestID ...
func GetWorkerRequestID(ctx context.Context) string {
	workerRequestID, _ := ctx.Value("Worker-Request-Id").(string)
	return workerRequestID
}
