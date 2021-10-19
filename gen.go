package macospkg

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/korylprince/go-cpio-odc"
)

// GeneratePkg creates a Distribution style, payload free pkg with identifier, version, and postinstall.
// See http://bomutils.dyndns.org/tutorial.html
func GeneratePkg(identifier, version string, postinstall []byte) ([]byte, error) {
	temp, err := os.MkdirTemp("", "macospkg-")
	if err != nil {
		return nil, fmt.Errorf("could not create temporary directory: %w", err)
	}
	defer os.RemoveAll(temp)

	// write Distribution
	buf := new(bytes.Buffer)
	if err = tmplDistribution.Execute(buf, struct {
		Identifier string
		Version    string
	}{identifier, version}); err != nil {
		return nil, fmt.Errorf("could not execute Distribution template: %w", err)
	}

	if err = os.WriteFile(path.Join(temp, "Distribution"), buf.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("could not write Distribution to %s: %w", temp, err)
	}

	// create flat package
	if err = os.Mkdir(path.Join(temp, "payload.pkg"), 0755); err != nil {
		return nil, fmt.Errorf("could not write create %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	// write PackageInfo
	buf = new(bytes.Buffer)
	if err = tmplPackageInfo.Execute(buf, struct {
		Identifier string
		Version    string
	}{identifier, version}); err != nil {
		return nil, fmt.Errorf("could not execute PackageInfo template: %w", err)
	}

	if err = os.WriteFile(path.Join(temp, "payload.pkg", "PackageInfo"), buf.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("could not write PackageInfo to %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	// create Scripts payload
	f, err := os.OpenFile(path.Join(temp, "payload.pkg", "Scripts"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not create Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	gzw := gzip.NewWriter(f)

	// write payload into cpio archive
	cw := cpio.NewWriter(gzw, 0)
	postFile := &cpio.File{FileMode: 0755, UID: 0, GID: 80, Path: "./postinstall", Body: postinstall}
	if err = cw.WriteFile(postFile); err != nil {
		return nil, fmt.Errorf("could not write postinstall: %w", err)
	}
	if _, err = cw.Close(); err != nil {
		return nil, fmt.Errorf("could not close Scripts cpio archive: %w", err)
	}

	if err = gzw.Close(); err != nil {
		f.Close()
		return nil, fmt.Errorf("could not finish writing to Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("could not close Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	// package everything with xar
	cmd := exec.Command("xar", "--compression", "none", "-cf", "-", "Distribution", "payload.pkg")
	cmd.Dir = temp

	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not create xar archive: %w", err)
	}

	return b, nil
}
