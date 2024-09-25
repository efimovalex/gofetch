package gofetch

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

// TLSConfig returns a tls.Config object based on the provided paths to TLS CA, Cert and Key files
func TLSConfig(ctx context.Context, TLSCA, TLSCert, TLSKey string, insecureSkipVerify bool) (*tls.Config, error) {
	logger := slog.Default()
	if TLSKey == "" || TLSCert == "" {
		logger.Error("TLS Key and Cert file paths not provided, TLS not configured")

		return nil, errors.New("TLS Key and Cert file paths not provided, TLS not configured")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if TLSCA != "" {
		pool, err := makeCertPool([]string{TLSCA})
		if err != nil {
			logger.Error("Error loading CA certificate", "error", err)

			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if TLSCert != "" && TLSKey != "" {
		err := loadCertificate(tlsConfig, TLSCert, TLSKey)
		if err != nil {
			logger.Error("Error loading certificate and key", "error", err)

			return nil, err
		}
	}

	return tlsConfig, nil
}

// makeCertPool - make Cert pool and append them
func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("could not read certificate %s: %s", certFile, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf("could not parse any PEM certificates %s: %s", certFile, err)
		}
	}
	return pool, nil
}

// loadCertificate - Load the certificates for SSL connection
func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf(
			"could not load keypair %s:%s: %s", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	return nil
}
