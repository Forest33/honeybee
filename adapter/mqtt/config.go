package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Host                 string
	Port                 int
	ClientID             string
	User                 string
	Password             string
	UseTLS               bool
	ServerTLS            bool
	CACert               string
	Cert                 string
	Key                  string
	InsecureSkipVerify   bool
	ConnectRetryInterval time.Duration
	Timeout              time.Duration
}

func (cfg Config) getTLSConfig() (*tls.Config, error) {
	if !cfg.UseTLS {
		return nil, nil
	}

	ca, err := os.ReadFile(cfg.CACert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	cert, err := os.ReadFile(cfg.Cert)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	key, err := os.ReadFile(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate key: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	serverCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{serverCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		ClientCAs:          certPool,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}, nil
}
