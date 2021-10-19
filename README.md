[![pkg.go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/korylprince/go-macos-pkg)

# About

`go-macos-pkg` was built out of a desire to generate and sign macOS pkgs on non-macOS OSs.

# Caveats

* Right now most of the heavy lifting is done by calling the `xar` utility, as there aren't currently (2021) Go libraries that write xar archives. The good news is `xar` is available on other OSs, unlike productsign/productbuild
* Right now only payload-free (e.g. just runs `postinstall`) packages can be generated as that's all I currently need. It shouldn't be too hard for someone to add payload capabilities, though.

# Usage

```go
script := []byte("#!/bin/bash\necho 'hello, world!'\n")

pkg, err := macospkg.GeneratePkg("com.github.korylprince.go-macos-pkg", "1.0.0", script)
if err != nil {
    log.Fatalln("could not generate pkg:", err)
}

// use Apple Developer ID Installer cert and key to sign
signed, err := macospkg.SignPkg(pkg, cert, key)
if err != nil {
    log.Fatalln("could not sign pkg:", err)
}

if err = os.WriteFile("signed.pkg", signed, 0644); err != nil {
    log.Fatalln("could not write signed pkg:", err)
}
```
