# Quantum Service - DISABLED

**Status**: Not implemented (stub only)
**MVP Phase**: Excluded
**Future**: May be implemented for quantum-safe key exchange

## Why Disabled

This service currently contains only a 44-line stub with mock endpoints that return empty/hardcoded data:
- `GET /health` - Returns static health status
- `GET /keys` - Returns empty array
- `POST /generate` - Returns mock success message
- `GET /metrics` - Returns placeholder metrics

**No real quantum key distribution functionality exists.**

It should not be deployed until real quantum key distribution is implemented.

## Current Implementation

See [main.go](main.go) - 44 lines of stub code with Gin server and mock endpoints.

## Re-enabling

To implement this service properly:

1. **Research**: Study quantum key distribution protocols (QKD):
   - BB84 protocol
   - E91 protocol
   - Modern implementations (ID-QKD, CV-QKD)

2. **Integration Options**:
   - Connect to quantum hardware (if available)
   - Use quantum simulation libraries
   - Integrate with cloud quantum services (IBM Q, AWS Braket)

3. **Implementation**:
   - Real key generation using quantum entropy
   - Key distribution protocol implementation
   - Secure key storage and management
   - Authentication and authorization

4. **Testing**:
   - Unit tests for key generation
   - Integration tests with relay
   - Security audits
   - Performance benchmarks

5. **Deployment**:
   - Create Kubernetes deployment manifests
   - Add monitoring and alerting
   - Document operational procedures

## Dependencies (When Implemented)

- Quantum random number generator (QRNG)
- Secure key storage (HSM or encrypted database)
- Authentication service (Zitadel)
- Metrics collection (Prometheus)

## References

- [Quantum Key Distribution - Wikipedia](https://en.wikipedia.org/wiki/Quantum_key_distribution)
- [ETSI QKD Standards](https://www.etsi.org/technologies/quantum-key-distribution)
- [Post-Quantum Cryptography](https://csrc.nist.gov/projects/post-quantum-cryptography)

---

Last Updated: 2025-10-13
Decision: Keep stub for future implementation, do not deploy in MVP
