package macospkg

import (
	"crypto/x509"
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed files/Distribution
var distribution string
var tmplDistribution = template.Must(template.New("Distribution").Parse(distribution))

//go:embed files/PackageInfo
var packageInfo string
var tmplPackageInfo = template.Must(template.New("PackageInfo").Parse(packageInfo))

//go:embed files/AppleIncRootCertificate.cer
var certAppleRoot []byte
var certAppleRootParsed = mustParseCertificate(certAppleRoot)

//go:embed files/DeveloperIDCA.cer
var certDeveloperID []byte

func mustParseCertificate(der []byte) *x509.Certificate {
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic(fmt.Errorf("could not parse certificate: %w", err))
	}
	return cert
}
