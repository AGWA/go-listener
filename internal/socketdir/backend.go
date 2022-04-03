package socketdir

type BackendType struct {
	socketFilename string
	TLS            bool
	ProxyProto     bool
}

var backendTypes = []BackendType{
	BackendType{"cleartext", false, false},
	BackendType{"cleartext+proxy", false, true},
	BackendType{"tls", true, false},
	BackendType{"tls+proxy", true, true},
}

type Backend struct {
	Hostname string
	Service  string
	Type     BackendType
}
