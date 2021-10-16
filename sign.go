package macospkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
)

// SignPkg signs and returns the given pkg with the given certificate and key.
// The certificate should be an "Apple Developer ID Installer" certificate.
// See https://mackyle.github.io/xar/howtosign.html
func SignPkg(pkg []byte, cert *x509.Certificate, key *rsa.PrivateKey) ([]byte, error) {
	temp, err := os.MkdirTemp("", "macospkg-")
	if err != nil {
		return nil, fmt.Errorf("could not create temporary directory: %w", err)
	}
	defer os.RemoveAll(temp)

	if err = os.WriteFile(path.Join(temp, "archive.pkg"), pkg, 0600); err != nil {
		return nil, fmt.Errorf("could not write archive.pkg to %s: %w", temp, err)
	}

	if err = os.WriteFile(path.Join(temp, "cert.cer"), cert.Raw, 0600); err != nil {
		return nil, fmt.Errorf("could not write cert.cer to %s: %w", temp, err)
	}

	if err = os.WriteFile(path.Join(temp, "inter.cer"), certDeveloperID, 0600); err != nil {
		return nil, fmt.Errorf("could not write inter.cer to %s: %w", temp, err)
	}

	if err = os.WriteFile(path.Join(temp, "root.cer"), certAppleRoot, 0600); err != nil {
		return nil, fmt.Errorf("could not write root.cer to %s: %w", temp, err)
	}

	cmd := exec.Command("xar", "--sign", "-f", path.Join(temp, "archive.pkg"), "--digestinfo-to-sign", path.Join(temp, "digest.dat"), "--sig-size", strconv.Itoa(len(cert.Signature)), "--cert-loc", path.Join(temp, "cert.cer"), "--cert-loc", path.Join(temp, "inter.cer"), "--cert-loc", path.Join(temp, "root.cer"))
	if b, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("could not write prepare archive.pkg for signing: %w: %s", err, string(b))
	}

	digest, err := os.ReadFile(path.Join(temp, "digest.dat"))
	if err != nil {
		return nil, fmt.Errorf("could not read digest.dat: %w", err)
	}

	// sign data directly without hashing
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, 0, digest)
	if err != nil {
		return nil, fmt.Errorf("could not sign digest info: %w", err)
	}

	if err = os.WriteFile(path.Join(temp, "digest.sig"), sig, 0600); err != nil {
		return nil, fmt.Errorf("could not write digest.sig to %s: %w", temp, err)
	}

	cmd = exec.Command("xar", "--inject-sig", path.Join(temp, "digest.sig"), "-f", path.Join(temp, "archive.pkg"))
	if b, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("could not inject signature in archive.pkg: %w: %s", err, string(b))
	}

	signed, err := os.ReadFile(path.Join(temp, "archive.pkg"))
	if err != nil {
		return nil, fmt.Errorf("could not read signed archive.pkg: %w", err)
	}

	return signed, nil
}
