package macospkg

import (
	_ "embed"
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

//go:embed files/DeveloperIDCA.cer
var certDeveloperID []byte
