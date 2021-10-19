package macospkg

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type ManifestHash int

const (
	ManifestHashMD5 ManifestHash = iota
	ManifestHashSHA256
)

type Asset struct {
	Kind       string   `plist:"kind" json:"kind"`
	MD5Size    int      `plist:"md5-size,omitempty" json:"md5-size,omitempty"`
	MD5s       []string `plist:"md5s,omitempty" json:"md5s,omitempty"`
	SHA256Size int      `plist:"sha256-size,omitempty" json:"sha256-size,omitempty"`
	SHA256s    []string `plist:"sha256s,omitempty" json:"sha256s,omitempty"`
	URL        string   `plist:"url" json:"url"`
}

type Item struct {
	Assets []*Asset `plist:"assets" json:"assets"`
}

// Manifest is used by a MDM's InstallApplication or InstallEnterpiseApplication's command
type Manifest struct {
	Items []*Item `plist:"items" json:"items"`
}

// NewManifest generates a Manifest by hashing the given package (which should be signed) and setting the URL.
// h should be ManifestHashMD5 for InstallApplication commands and ManifestHashSHA256 for InstallEnterpiseApplication commands
func NewManifest(pkg []byte, url string, h ManifestHash) *Manifest {
	a := &Asset{
		Kind: "software-package",
		URL:  url,
	}
	if h == ManifestHashMD5 {
		hash := md5.Sum(pkg)
		a.MD5Size = len(pkg)
		a.MD5s = []string{hex.EncodeToString(hash[:])}
	} else if h == ManifestHashSHA256 {
		hash := sha256.Sum256(pkg)
		a.SHA256Size = len(pkg)
		a.SHA256s = []string{hex.EncodeToString(hash[:])}
	} else {
		panic(fmt.Errorf("unknown hash type: %v", h))
	}

	return &Manifest{Items: []*Item{{Assets: []*Asset{a}}}}
}
