package macospkg

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
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

	// write temporary payload
	if err = os.WriteFile(path.Join(temp, "postinstall"), postinstall, 0755); err != nil {
		return nil, fmt.Errorf("could not write payload to %s: %w", temp, err)
	}

	// create Scripts cpio/gz payload
	cmd := exec.Command("cpio", "-o", "--format", "odc", "--owner", "0:80")
	cmd.Dir = temp

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not get cpio stdin: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not get cpio stdout: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start cpio: %w", err)
	}

	if _, err = stdin.Write([]byte("./postinstall\n")); err != nil {
		return nil, fmt.Errorf("could write to cpio: %w", err)
	}

	if err = stdin.Close(); err != nil {
		return nil, fmt.Errorf("could not close cpio stdin: %w", err)
	}

	cpio, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("could not read cpio stdout: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("could not stop cpio successfully: %w", err)
	}

	f, err := os.OpenFile(path.Join(temp, "payload.pkg", "Scripts"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not create Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	gzw := gzip.NewWriter(f)
	if _, err = gzw.Write(cpio); err != nil {
		f.Close()
		return nil, fmt.Errorf("could not write to Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	if err = gzw.Close(); err != nil {
		f.Close()
		return nil, fmt.Errorf("could not finish writing to Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("could not close Scripts in %s: %w", path.Join(temp, "payload.pkg"), err)
	}

	// clean postinstall file
	if err = os.Remove(path.Join(temp, "postinstall")); err != nil {
		return nil, fmt.Errorf("could not remove payload from %s: %w", temp, err)
	}

	// package everything with xar
	cmd = exec.Command("xar", "--compression", "none", "-cf", "-", "Distribution", "payload.pkg")
	cmd.Dir = temp

	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not create xar archive: %w", err)
	}

	return b, nil
}
