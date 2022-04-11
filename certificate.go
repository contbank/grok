package grok

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"hash"
	"net/http"
	"os"

	"github.com/xdg-go/pbkdf2"
)

func appendOID(b asn1.ObjectIdentifier, v ...int) asn1.ObjectIdentifier {
	n := make(asn1.ObjectIdentifier, len(b), len(b)+len(v))
	copy(n, b)
	return append(n, v...)
}

var (
	oidRSADSI              = asn1.ObjectIdentifier{1, 2, 840, 113549}
	oidPKCS5               = appendOID(oidRSADSI, 1, 5)
	oidPBKDF2              = appendOID(oidPKCS5, 12)
	oidPBES2               = appendOID(oidPKCS5, 13)
	oidDigestAlgorithm     = appendOID(oidRSADSI, 2)
	oidHMACWithSHA1        = appendOID(oidDigestAlgorithm, 7)
	oidHMACWithSHA224      = appendOID(oidDigestAlgorithm, 8)
	oidHMACWithSHA256      = appendOID(oidDigestAlgorithm, 9)
	oidHMACWithSHA384      = appendOID(oidDigestAlgorithm, 10)
	oidHMACWithSHA512      = appendOID(oidDigestAlgorithm, 11)
	oidHMACWithSHA512_224  = appendOID(oidDigestAlgorithm, 12)
	oidHMACWithSHA512_256  = appendOID(oidDigestAlgorithm, 13)
	oidEncryptionAlgorithm = appendOID(oidRSADSI, 3)
	oidDESCBC              = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 7}
	oidDESEDE3CBC          = appendOID(oidEncryptionAlgorithm, 7)
	oidAES                 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1}
	oidAES128CBCPAD        = appendOID(oidAES, 2)
	oidAES192CBCPAD        = appendOID(oidAES, 22)
	oidAES256CBCPAD        = appendOID(oidAES, 42)
)

func prfByOID(oid asn1.ObjectIdentifier) func() hash.Hash {
	if len(oid) == 0 {
		return sha1.New
	}
	if oid.Equal(oidHMACWithSHA1) {
		return sha1.New
	}
	if oid.Equal(oidHMACWithSHA224) {
		return sha256.New224
	}
	if oid.Equal(oidHMACWithSHA256) {
		return sha256.New
	}
	if oid.Equal(oidHMACWithSHA384) {
		return sha512.New384
	}
	if oid.Equal(oidHMACWithSHA512) {
		return sha512.New
	}
	if oid.Equal(oidHMACWithSHA512_224) {
		return sha512.New512_224
	}
	if oid.Equal(oidHMACWithSHA512_256) {
		return sha512.New512_256
	}
	return nil
}

func encsByOID(oid asn1.ObjectIdentifier) (func([]byte) (cipher.Block, error), func(cipher.Block, []byte) cipher.BlockMode, int) {
	if oid.Equal(oidDESCBC) {
		return des.NewCipher, cipher.NewCBCDecrypter, 8
	}
	if oid.Equal(oidDESEDE3CBC) {
		return des.NewTripleDESCipher, cipher.NewCBCDecrypter, 24
	}
	if oid.Equal(oidAES128CBCPAD) {
		return aes.NewCipher, cipher.NewCBCDecrypter, 16
	}
	if oid.Equal(oidAES192CBCPAD) {
		return aes.NewCipher, cipher.NewCBCDecrypter, 24
	}
	if oid.Equal(oidAES256CBCPAD) {
		return aes.NewCipher, cipher.NewCBCDecrypter, 32
	}
	return nil, nil, 0
}

func decryptPBES2(b, password []byte, maxIter int) (data, rest []byte, err error) {
	var p struct {
		ES struct {
			ID     asn1.ObjectIdentifier
			Params struct {
				KDF struct {
					ID     asn1.ObjectIdentifier
					Params struct {
						Salt      []byte
						Iter      int
						KeyLength int `asn1:"optional"`
						PRF       struct {
							ID     asn1.ObjectIdentifier
							Params asn1.RawValue
						} `asn1:"optional"`
					}
				}
				EncS struct {
					ID     asn1.ObjectIdentifier
					Params []byte
				}
			}
		}
		Data []byte
	}
	rest, err = asn1.Unmarshal(b, &p)
	if err != nil {
		return
	}
	if !p.ES.ID.Equal(oidPBES2) {
		return
	}
	if !p.ES.Params.KDF.ID.Equal(oidPBKDF2) {
		return
	}
	if p.ES.Params.KDF.Params.Iter < 1 {
		return
	}
	prf := prfByOID(p.ES.Params.KDF.Params.PRF.ID)
	if prf == nil {
		return
	}
	bcf, bmf, kl := encsByOID(p.ES.Params.EncS.ID)
	if bcf == nil || bmf == nil {
		return
	}
	if len(p.Data) == 0 {
		return
	}
	if maxIter > 0 && p.ES.Params.KDF.Params.Iter > maxIter {
		return
	}
	key := pbkdf2.Key(password, p.ES.Params.KDF.Params.Salt, p.ES.Params.KDF.Params.Iter, kl, prf)
	var bc cipher.Block
	bc, err = bcf(key)
	if err != nil {
		return
	}
	if len(p.ES.Params.EncS.Params) != bc.BlockSize() {
		return
	}
	bm := bmf(bc, p.ES.Params.EncS.Params)
	if len(p.Data)%bm.BlockSize() != 0 {
		return
	}
	data = make([]byte, len(p.Data))
	bm.CryptBlocks(data, p.Data)
	pl := data[len(data)-1]
	if pl == 0 || int(pl) > bm.BlockSize() {
		return
	}
	dl := len(data) - int(pl)
	for _, b := range data[dl:] {
		if b != pl {
			return
		}
	}
	data = data[:dl]
	return
}

func LoadCertificate(cert []byte, key []byte, passphrase string) (*tls.Certificate, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, NewError(http.StatusInternalServerError, "failed to decode PEM block")
	}

	var derKey []byte
	var err error
	password := []byte(passphrase)
	if block.Type == "ENCRYPTED PRIVATE KEY" {
		derKey, _, err = decryptPBES2(block.Bytes, password, 1000000)
		if err != nil {
			return nil, NewError(http.StatusInternalServerError, err.Error())
		}
	}

	privKey, err := x509.ParsePKCS8PrivateKey(derKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse PKCS #8 private key:", err)
		os.Exit(1)
	}

	certDERBlock, _ := pem.Decode([]byte(cert))

	var certificate tls.Certificate

	certificate.Certificate = append(certificate.Certificate, certDERBlock.Bytes)
	certificate.PrivateKey = privKey.(crypto.PrivateKey)

	return &certificate, nil
}
