package grok

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewMongoConnection ...
func NewMongoConnection(connectionString string, caFilePath *string) *mongo.Client {
	var client *mongo.Client
	var err error

	if caFilePath != nil {
		tlsConfig := getCustomTLSConfig(*caFilePath)

		client, err = mongo.NewClient(options.Client().ApplyURI(connectionString).SetTLSConfig(tlsConfig))
	} else {
		client, err = mongo.NewClient(options.Client().ApplyURI(connectionString))
	}

	if err != nil {
		logrus.WithError(err).Panic("Error connecting to MongoDB")
	}

	client.Connect(context.Background())

	err = client.Ping(context.Background(), readpref.Primary())

	if err != nil {
		logrus.WithError(err).Panic("Error pinging MongoDB")
	}

	return client
}

func getCustomTLSConfig(caFile string) *tls.Config {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		logrus.WithError(err).
			Panic("error loading ca")
		return nil
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		logrus.WithError(err).
			Panic("error tls")
		return nil
	}

	return tlsConfig
}
