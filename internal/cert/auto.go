// Copyright (C) 2022 Andrew Ayer
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// Except as contained in this notice, the name(s) of the above copyright
// holders shall not be used in advertising or otherwise to promote the
// sale, use or other dealings in this Software without prior written
// authorization.

package cert

import (
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"os"
	"path/filepath"
)

func getAutocertCache() autocert.Cache {
	if cacheDir := os.Getenv("AUTOCERT_CACHE_DIR"); cacheDir != "" {
		return autocert.DirCache(cacheDir)
	} else if os.Getuid() == 0 {
		return autocert.DirCache("/var/lib/autocert-cache")
	} else if dataDir := os.Getenv("XDG_DATA_HOME"); dataDir != "" {
		return autocert.DirCache(filepath.Join(dataDir, "autocert-cache"))
	} else if homeDir, err := os.UserHomeDir(); err == nil {
		return autocert.DirCache(filepath.Join(homeDir, ".local/share/autocert-cache"))
	} else {
		return nil
	}
}

func getCertificateAutomatically(hostPolicy autocert.HostPolicy) GetCertificateFunc {
	manager := &autocert.Manager{
		Client: &acme.Client{
			DirectoryURL: os.Getenv("AUTOCERT_ACME_SERVER"),
		},

		Prompt:     autocert.AcceptTOS,
		Cache:      getAutocertCache(),
		HostPolicy: hostPolicy,
		Email:      os.Getenv("AUTOCERT_EMAIL"),
	}
	return manager.GetCertificate
}

func GetCertificateAutomatically(hostnames []string) GetCertificateFunc {
	if hostnames == nil {
		return getCertificateAutomatically(nil)
	} else {
		return getCertificateAutomatically(autocert.HostWhitelist(hostnames...))
	}
}
