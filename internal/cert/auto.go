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
