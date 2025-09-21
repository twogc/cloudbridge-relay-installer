package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/netip"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

const (
	capsuleTypeAddressAssign      http3.CapsuleType = 1
	capsuleTypeAddressRequest     http3.CapsuleType = 2
	capsuleTypeRouteAdvertisement http3.CapsuleType = 3
)

// IPCapsule представляет IP Capsule для CONNECT-IP (из connect-ip-go)
type IPCapsule struct {
	IPPacket []byte
	Checksum uint16
}

// Validate проверяет целостность IP пакета
func (c *IPCapsule) Validate() error {
	if len(c.IPPacket) < 20 {
		return errors.New("IP packet too short")
	}

	// Проверяем версию IP
	version := c.IPPacket[0] >> 4
	if version != 4 && version != 6 {
		return fmt.Errorf("invalid IP version: %d", version)
	}

	// Проверяем checksum для IPv4
	if version == 4 {
		calculatedChecksum := c.calculateIPv4Checksum()
		if calculatedChecksum != c.Checksum {
			return fmt.Errorf("checksum mismatch: expected %d, got %d", calculatedChecksum, c.Checksum)
		}
	}

	return nil
}

// calculateIPv4Checksum вычисляет checksum для IPv4 пакета
func (c *IPCapsule) calculateIPv4Checksum() uint16 {
	if len(c.IPPacket) < 20 {
		return 0
	}

	var sum uint32
	// Суммируем заголовок по 16-бит словам
	for i := 0; i < 20; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(c.IPPacket[i : i+2]))
	}

	// Добавляем carry bits
	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	// Инвертируем результат
	return ^uint16(sum)
}

// Encode кодирует IP пакет в capsule формат
func (c *IPCapsule) Encode() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid IP packet: %w", err)
	}

	// Создаем capsule с IP пакетом
	capsule := make([]byte, 0, len(c.IPPacket)+4)

	// Добавляем тип capsule (IP packet) - используем кастомный тип
	capsule = quicvarint.Append(capsule, uint64(capsuleTypeAddressAssign))

	// Добавляем длину
	capsule = quicvarint.Append(capsule, uint64(len(c.IPPacket)))

	// Добавляем IP пакет
	capsule = append(capsule, c.IPPacket...)

	return capsule, nil
}

// Decode декодирует capsule в IP пакет
func (c *IPCapsule) Decode(data []byte) error {
	if len(data) < 2 {
		return errors.New("capsule too short")
	}

	// Парсим тип capsule
	capsuleType, n, err := quicvarint.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse capsule type: %w", err)
	}

	if capsuleType != uint64(capsuleTypeAddressAssign) {
		return fmt.Errorf("unexpected capsule type: %d", capsuleType)
	}

	// Парсим длину
	length, m, err := quicvarint.Parse(data[n:])
	if err != nil {
		return fmt.Errorf("failed to parse capsule length: %w", err)
	}

	// Извлекаем IP пакет
	start := n + m
	if len(data) < start+int(length) {
		return errors.New("capsule truncated")
	}

	c.IPPacket = make([]byte, length)
	copy(c.IPPacket, data[start:start+int(length)])

	// Вычисляем checksum
	c.Checksum = c.calculateIPv4Checksum()

	return nil
}

// addressAssignCapsule представляет ADDRESS_ASSIGN capsule (из connect-ip-go)
type addressAssignCapsule struct {
	AssignedAddresses []AssignedAddress
}

// AssignedAddress представляет назначенный адрес
type AssignedAddress struct {
	RequestID uint64
	IPPrefix  netip.Prefix
}

func (a AssignedAddress) len() int {
	return quicvarint.Len(a.RequestID) + 1 + a.IPPrefix.Addr().BitLen()/8 + 1
}

// addressRequestCapsule представляет ADDRESS_REQUEST capsule
type addressRequestCapsule struct {
	RequestedAddresses []RequestedAddress
}

// RequestedAddress представляет запрашиваемый адрес
type RequestedAddress struct {
	RequestID uint64
	IPPrefix  netip.Prefix
}

func (r RequestedAddress) len() int {
	return quicvarint.Len(r.RequestID) + 1 + r.IPPrefix.Addr().BitLen()/8 + 1
}

// routeAdvertisementCapsule представляет ROUTE_ADVERTISEMENT capsule
type routeAdvertisementCapsule struct {
	IPAddressRanges []IPRoute
}

// IPRoute представляет IP маршрут
type IPRoute struct {
	StartIP    netip.Addr
	EndIP      netip.Addr
	IPProtocol uint8
}

func (r IPRoute) len() int {
	return 1 + r.StartIP.BitLen()/8 + r.EndIP.BitLen()/8 + 1
}

// IPCapsuleManager управляет IP Capsules
type IPCapsuleManager struct {
	logger *logger.Logger
}

// NewIPCapsuleManager создает новый IP Capsule Manager
func NewIPCapsuleManager(logger *logger.Logger) *IPCapsuleManager {
	return &IPCapsuleManager{
		logger: logger,
	}
}

// ProcessIPPacket обрабатывает IP пакет и создает capsule
func (icm *IPCapsuleManager) ProcessIPPacket(ipPacket []byte) (*IPCapsule, error) {
	capsule := &IPCapsule{
		IPPacket: ipPacket,
	}

	if err := capsule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid IP packet: %w", err)
	}

	icm.logger.Debug("Processed IP packet",
		zap.Int("packet_len", len(ipPacket)),
		zap.Uint16("checksum", capsule.Checksum))

	return capsule, nil
}

// ProcessCapsule обрабатывает capsule и извлекает IP пакет
func (icm *IPCapsuleManager) ProcessCapsule(data []byte) (*IPCapsule, error) {
	capsule := &IPCapsule{}

	if err := capsule.Decode(data); err != nil {
		return nil, fmt.Errorf("failed to decode capsule: %w", err)
	}

	icm.logger.Debug("Processed capsule",
		zap.Int("packet_len", len(capsule.IPPacket)),
		zap.Uint16("checksum", capsule.Checksum))

	return capsule, nil
}

// parseAddress парсит адрес из reader (из connect-ip-go)
func parseAddress(r io.Reader) (requestID uint64, prefix netip.Prefix, _ error) {
	vr := quicvarint.NewReader(r)
	requestID, err := quicvarint.Read(vr)
	if err != nil {
		return 0, netip.Prefix{}, err
	}
	ipVersion, err := vr.ReadByte()
	if err != nil {
		return 0, netip.Prefix{}, err
	}
	var ip netip.Addr
	switch ipVersion {
	case 4:
		var ipv4 [4]byte
		if _, err := io.ReadFull(r, ipv4[:]); err != nil {
			return 0, netip.Prefix{}, err
		}
		ip = netip.AddrFrom4(ipv4)
	case 6:
		var ipv6 [16]byte
		if _, err := io.ReadFull(r, ipv6[:]); err != nil {
			return 0, netip.Prefix{}, err
		}
		ip = netip.AddrFrom16(ipv6)
	default:
		return 0, netip.Prefix{}, fmt.Errorf("invalid IP version: %d", ipVersion)
	}
	prefixLen, err := vr.ReadByte()
	if err != nil {
		return 0, netip.Prefix{}, err
	}
	if int(prefixLen) > ip.BitLen() {
		return 0, netip.Prefix{}, fmt.Errorf("prefix length %d exceeds IP address length (%d)", prefixLen, ip.BitLen())
	}
	prefix = netip.PrefixFrom(ip, int(prefixLen))
	if prefix != prefix.Masked() {
		return 0, netip.Prefix{}, errors.New("lower bits not covered by prefix length are not all zero")
	}
	return requestID, prefix, nil
}

// parseIPAddressRange парсит IP диапазон (из connect-ip-go)
func parseIPAddressRange(r io.Reader) (IPRoute, error) {
	var ipVersion uint8
	if err := binary.Read(r, binary.LittleEndian, &ipVersion); err != nil {
		return IPRoute{}, err
	}

	var startIP, endIP netip.Addr
	switch ipVersion {
	case 4:
		var start, end [4]byte
		if _, err := io.ReadFull(r, start[:]); err != nil {
			return IPRoute{}, err
		}
		if _, err := io.ReadFull(r, end[:]); err != nil {
			return IPRoute{}, err
		}
		startIP = netip.AddrFrom4(start)
		endIP = netip.AddrFrom4(end)
	case 6:
		var start, end [16]byte
		if _, err := io.ReadFull(r, start[:]); err != nil {
			return IPRoute{}, err
		}
		if _, err := io.ReadFull(r, end[:]); err != nil {
			return IPRoute{}, err
		}
		startIP = netip.AddrFrom16(start)
		endIP = netip.AddrFrom16(end)
	default:
		return IPRoute{}, fmt.Errorf("invalid IP version: %d", ipVersion)
	}

	if startIP.Compare(endIP) > 0 {
		return IPRoute{}, errors.New("start IP is greater than end IP")
	}

	var ipProtocol uint8
	if err := binary.Read(r, binary.LittleEndian, &ipProtocol); err != nil {
		return IPRoute{}, err
	}
	return IPRoute{
		StartIP:    startIP,
		EndIP:      endIP,
		IPProtocol: ipProtocol,
	}, nil
}
