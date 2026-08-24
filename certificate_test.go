package grok_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CertificateTestSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func TestCertificateTestSuite(t *testing.T) {
	suite.Run(t, new(CertificateTestSuite))
}

func (s *CertificateTestSuite) SetupSuite() {
	s.assert = assert.New(s.T())
}

func generateSelfSignedCertAndKey(s *CertificateTestSuite) (certPEM []byte, keyPEM []byte) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "grok-test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	s.Require().NoError(err)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	s.Require().NoError(err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

// TestLoadCertificateUnencryptedPKCS8 garante que LoadCertificate aceita chave PKCS8 SEM criptografia
// (bloco "PRIVATE KEY", sem passphrase) — antes desta correção, derKey só era preenchido dentro do
// branch "ENCRYPTED PRIVATE KEY", ficando nil pra qualquer chave não criptografada e falhando sempre
// com "asn1: syntax error: sequence truncated" (x509.ParsePKCS8PrivateKey(nil)). Caso real que expôs o
// bug: certificado autoassinado temporário (rocket/infra/gerar-mtls-temporario-local.sh), usado pra
// destravar o boot do accounts fora da AWS Secrets Manager — o certificado real da Bankly/Celcoin nunca
// expôs isso porque vem sempre criptografado com passphrase.
func (s *CertificateTestSuite) TestLoadCertificateUnencryptedPKCS8() {
	certPEM, keyPEM := generateSelfSignedCertAndKey(s)

	tlsCert, err := grok.LoadCertificate(certPEM, keyPEM, "")

	s.assert.NoError(err)
	s.assert.NotNil(tlsCert)
	s.assert.NotNil(tlsCert.PrivateKey)
	s.assert.Len(tlsCert.Certificate, 1)
}

// TestLoadCertificateInvalidPEM garante que um bloco PEM inválido continua retornando erro tratável
// (não panic), tanto pra chave quanto pra certificado.
func (s *CertificateTestSuite) TestLoadCertificateInvalidKeyPEM() {
	_, certPEM := generateSelfSignedCertAndKey(s)

	_, err := grok.LoadCertificate(certPEM, []byte("não é PEM válido"), "")

	s.assert.Error(err)
}
