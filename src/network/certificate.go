package network

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"ows/ledger"
)

func makeTLSCertificate(keyPair ledger.KeyPair) (*tls.Certificate, error) {
	// Create an X.509 certificate template
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "OWS Peer Certificate",
			Organization: []string{"OWS"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1-year validity
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, ed25519.PublicKey(keyPair.Public), ed25519.PrivateKey(keyPair.Private))
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate (%v)", err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  ed25519.PrivateKey(keyPair.Private),
	}, nil
}

func makeClientTLSConfig(cert *tls.Certificate, isValidPeer func(k ledger.PublicKey) bool) *tls.Config {
	return &tls.Config{
		Certificates:          []tls.Certificate{*cert},
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: makeVerifyPeerCertificate(isValidPeer),
	}
}

func makeServerTLSConfig(cert *tls.Certificate, isValidPeer func(k ledger.PublicKey) bool) *tls.Config {
	return &tls.Config{
		Certificates:          []tls.Certificate{*cert},
		ClientAuth:            tls.RequestClientCert,
		VerifyPeerCertificate: makeVerifyPeerCertificate(isValidPeer),
	}
}

func makeVerifyPeerCertificate(isValidPeer func(k ledger.PublicKey) bool) func(rawPeerCerts [][]byte, chain [][]*x509.Certificate) error {
	return func(rawPeerCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawPeerCerts) == 0 {
			return fmt.Errorf("no peer certificates provided")
		}

		peerCert, err := x509.ParseCertificate(rawPeerCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse peer certificate (%v)", err)
		}

		pk, ok := peerCert.PublicKey.(ed25519.PublicKey)
		if !ok {
			return fmt.Errorf("peer certificate is not using Ed25519")
		}

		if !isValidPeer(ledger.PublicKey(pk)) {
			return fmt.Errorf("invalid peer")
		}

		return nil
	}
}
