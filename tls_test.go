package gohans

import (
	"context"
	"crypto/tls"
	"os"
	"testing"

	"github.com/madflojo/testcerts"
	"github.com/stretchr/testify/assert"
)

func TestTLSConfig(t *testing.T) {
	ctx := context.Background()
	// Generate Certificate Authority
	ca := testcerts.NewCA()
	ca.ToFile("/tmp/ca.crt", "/tmp/ca.key")

	// Create a signed Certificate and Key for "localhost"
	certs, err := ca.NewKeyPair("localhost")
	assert.NoError(t, err)

	// Write certificates to a file
	err = certs.ToFile("/tmp/cert.crt", "/tmp/key.key")
	assert.NoError(t, err)

	f, err := os.Create("/tmp/dummy-ca.key")
	assert.NoError(t, err)
	f.Close()

	f, err = os.Create("/tmp/dummy-cert.crt")
	assert.NoError(t, err)
	f.Close()

	f, err = os.Create("/tmp/dummy-key.key")
	assert.NoError(t, err)
	f.Close()

	type args struct {
		ctx                context.Context
		TLSCA              string
		TLSCert            string
		TLSKey             string
		InsecureSkipVerify bool
	}

	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Test TLSConfig non existent files",
			args: args{
				ctx:                ctx,
				TLSCA:              "ca.pem",
				TLSCert:            "cert.pem",
				TLSKey:             "key.pem",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "could not read certificate ca.pem: open ca.pem: no such file or directory",
		},
		{
			name: "Test TLSConfig wrong content",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/dummy-ca.key",
				TLSCert:            "/tmp/cert.crt",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "could not parse any PEM certificates /tmp/dummy-ca.key: %!s(<nil>)",
		},
		{
			name: "Test TLSConfig wrong cert:key content",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "/tmp/dummy-cert.crt",
				TLSKey:             "/tmp/dummy-key.key",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "could not load keypair /tmp/dummy-cert.crt:/tmp/dummy-key.key: tls: failed to find any PEM data in certificate input",
		},
		{
			name: "Test TLSConfig wrong cert content",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "/tmp/dummy-cert.crt",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "could not load keypair /tmp/dummy-cert.crt:/tmp/key.key: tls: failed to find any PEM data in certificate input",
		},
		{
			name: "Test TLSConfig wrong key content",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "/tmp/cert.crt",
				TLSKey:             "/tmp/dummy-key.key",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "could not load keypair /tmp/cert.crt:/tmp/dummy-key.key: tls: failed to find any PEM data in key input",
		},
		{
			name: "Test TLSConfig missing cert, no CA",
			args: args{
				ctx:                ctx,
				TLSCA:              "",
				TLSCert:            "",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "TLS Key and Cert file paths not provided, TLS not configured",
		},
		{
			name: "Test TLSConfig missing cert",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "TLS Key and Cert file paths not provided, TLS not configured",
		},
		{
			name: "Test TLSConfig missing key",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "/tmp/cert.crt",
				TLSKey:             "",
				InsecureSkipVerify: false,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "TLS Key and Cert file paths not provided, TLS not configured",
		},
		{
			name: "Test TLSConfig successful",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "/tmp/cert.crt",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			wantErr: false,
		},
		{
			name: "Test TLSConfig successful, optional CA missing",
			args: args{
				ctx:                ctx,
				TLSCA:              "",
				TLSCert:            "/tmp/cert.crt",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			wantErr: false,
		},
		{
			name: "Test TLSConfig successful, InsecureSkipVerify false",
			args: args{
				ctx:                ctx,
				TLSCA:              "/tmp/ca.crt",
				TLSCert:            "/tmp/cert.crt",
				TLSKey:             "/tmp/key.key",
				InsecureSkipVerify: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TLSConfig(tt.args.ctx, tt.args.TLSCA, tt.args.TLSCert, tt.args.TLSKey, tt.args.InsecureSkipVerify)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Nil(t, got)

				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, got)
			if len(tt.args.TLSCA) > 0 {
				assert.NotNil(t, got.RootCAs)

				// We only have one CA
				assert.Equal(t, 1, len(got.RootCAs.Subjects()))
			}

			assert.NotNil(t, got.Certificates)
			assert.Equal(t, 1, len(got.Certificates))
			assert.NotNil(t, got.InsecureSkipVerify)
			assert.Equal(t, tt.args.InsecureSkipVerify, got.InsecureSkipVerify)
		})
	}
}
