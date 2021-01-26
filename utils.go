package grok

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func random(n float64) float64 {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return math.Round(r.Float64() * n)
}

func mod(dividendo float64, divisor float64) float64 {
	return math.Round(dividendo - (math.Floor(dividendo/divisor) * divisor))
}

//GeneratorCPF ...
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

//GeneratorCNPJ ...
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

//GeneratorIDBase ...
func GeneratorIDBase(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return strings.ToUpper(string(b))
}
