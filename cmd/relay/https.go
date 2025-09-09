package main

import (
	"net/http"
)

// startHTTPSIfEnabled starts the HTTPS server on :9091 when TLS is enabled.
// If TLS is enabled but the HTTPS server cannot be started, it returns an error
// so the caller can fail fast instead of silently falling back to HTTP.
func (s *Server) startHTTPSIfEnabled(httpMux *http.ServeMux) (*http.Server, error) {
    tlsServer := NewTLSServer(TLSConfig{
        Enabled:       s.config.Security.TLSEnabled,
        CertFile:      s.config.Security.CertFile,
        KeyFile:       s.config.Security.KeyFile,
        CipherSuites:  s.config.Security.CipherSuites,
        MinTLSVersion: s.config.Security.MinTLSVersion,
        MTLS: MTLSConfig{
            Enabled:            s.config.Security.MTLS.Enabled,
            CAFile:             s.config.Security.MTLS.CACert,
            ClientCertRequired: true,
        },
    }, s.logger.GetZapLogger())

    if err := tlsServer.ValidateTLSConfig(); err != nil {
        return nil, err
    }

    // Construct the server with TLS config; actual ListenAndServeTLS is started by caller
    server, err := tlsServer.StartHTTPServer(":9091", httpMux)
    if err != nil {
        return nil, err
    }
    return server, nil
}


