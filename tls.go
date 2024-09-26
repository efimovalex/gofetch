package gohans

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

	tlsConf := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		Renegotiation:      tls.RenegotiateOnceAsClient,
		MinVersion:         tls.VersionTLS12,
	}

	if TLSCA != "" {
		pool := x509.NewCertPool()
		pem, err := os.ReadFile(TLSCA)
		if err != nil {
			return nil, fmt.Errorf("could not read certificate %s: %s", TLSCA, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf("could not parse any PEM certificates %s: %s", TLSCA, err)
		}
		tlsConf.RootCAs = pool
	}

	if TLSCert != "" && TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(TLSCert, TLSKey)
		if err != nil {
			return nil, fmt.Errorf(
				"could not load keypair %s:%s: %s", TLSCert, TLSKey, err)
		}

		tlsConf.Certificates = []tls.Certificate{cert}
	}

	return tlsConf, nil
}
