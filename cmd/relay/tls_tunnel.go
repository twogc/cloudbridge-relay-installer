package main

import (
	"crypto/tls"
	"net"

	"go.uber.org/zap"
)

// startTLSTunnelListener starts a TLS-wrapped TCP listener on :9091 and
// processes connections with the same tunnel pipeline as the plain TCP listener.
func (s *Server) startTLSTunnelListener() error {
	tlsServer := NewTLSServer(TLSConfig{
		Enabled:       s.config.Security.TLSEnabled,
		CertFile:      s.config.Security.CertFile,
		KeyFile:       s.config.Security.KeyFile,
		CipherSuites:  s.config.Security.CipherSuites,
		MinTLSVersion: s.config.Security.MinTLSVersion,
		MTLS: MTLSConfig{
			Enabled:            s.config.Security.MTLS.Enabled,
			CAFile:             s.config.Security.MTLS.CACert,
			ClientCertRequired: s.config.Security.MTLS.Enabled,
		},
	}, s.logger.GetZapLogger())

	if err := tlsServer.ValidateTLSConfig(); err != nil {
		return err
	}

	tlsConfig, err := tlsServer.CreateTLSConfig()
	if err != nil {
		return err
	}

	baseListener, err := net.Listen("tcp", "127.0.0.1:9091")
	if err != nil {
		s.logger.Error("Failed to bind TCP listener for TLS", zap.Error(err))
		return err
	}
	s.listener = baseListener

	tlsListener := tls.NewListener(baseListener, tlsConfig)

	go func() {
		for {
			conn, err := tlsListener.Accept()
			if err != nil {
				if s.ctx.Err() != nil {
					return
				}
				s.logger.Error("Failed to accept TLS connection", zap.Error(err))
				continue
			}
			go s.handleConnection(conn)
		}
	}()

	s.logger.Info("TLS tunnel listener started",
		zap.Int("tls_port", 9091),
		zap.Bool("mtls_enabled", s.config.Security.MTLS.Enabled),
		zap.String("cert_file", s.config.Security.CertFile),
		zap.String("key_file", s.config.Security.KeyFile),
	)
	return nil
}
