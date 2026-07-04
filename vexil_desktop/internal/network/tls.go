package network

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TLSConfigBuilder struct {
	certFile           string
	keyFile            string
	insecureSkipVerify bool
	sessionCacheSize   int
	isServer           bool
}

func NewServerTLSBuilder(certFile, keyFile string) *TLSConfigBuilder {
	return &TLSConfigBuilder{
		certFile: certFile,
		keyFile:  keyFile,
		isServer: true,
	}
}

func NewClientTLSBuilder(insecureSkipVerify bool, sessionCacheSize int) *TLSConfigBuilder {
	return &TLSConfigBuilder{
		insecureSkipVerify: insecureSkipVerify,
		sessionCacheSize:   sessionCacheSize,
		isServer:           false,
	}
}

var (
	autoGenCertCache       *tls.Certificate
	autoGenCertCacheMu     sync.Mutex
	autoGenCertFingerprint string
)

func (b *TLSConfigBuilder) Build() (*tls.Config, error) {
	if b.isServer {
		return b.buildServerConfig()
	}
	return b.buildClientConfig()
}

func (b *TLSConfigBuilder) buildServerConfig() (*tls.Config, error) {
	var cert tls.Certificate
	var err error

	if b.certFile != "" && b.keyFile != "" {
		cert, err = tls.LoadX509KeyPair(b.certFile, b.keyFile)
		if err != nil {
			return nil, fmt.Errorf("加载 TLS 证书失败: %w", err)
		}
		fp := certFingerprint(&cert)
		fmt.Printf("  [TLS] 已加载自定义证书\n")
		fmt.Printf("  [TLS] 证书指纹 (SHA-256): %s\n", fp)
	} else {
		cert, err = b.generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("生成自签名证书失败: %w", err)
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		ClientAuth:   tls.NoClientCert,
	}, nil
}

func (b *TLSConfigBuilder) buildClientConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, // 自签名证书跳过系统CA验证
	}

	if b.sessionCacheSize > 0 {
		cfg.ClientSessionCache = tls.NewLRUClientSessionCache(b.sessionCacheSize)
	}

	// 基本验证：证书非空、未过期
	cfg.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("未收到服务器证书")
		}
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("解析证书失败: %w", err)
		}
		now := time.Now()
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			return fmt.Errorf("证书已过期或尚未生效 (有效: %s ~ %s)",
				cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
		}
		return nil
	}

	return cfg, nil
}

func (b *TLSConfigBuilder) generateSelfSignedCert() (tls.Certificate, error) {
	autoGenCertCacheMu.Lock()
	defer autoGenCertCacheMu.Unlock()

	if autoGenCertCache != nil {
		return *autoGenCertCache, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	cacheDir := filepath.Join(homeDir, ".vexil")
	certPath := filepath.Join(cacheDir, "auto_tls.crt")
	keyPath := filepath.Join(cacheDir, "auto_tls.key")

	if cachedCert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
		autoGenCertCache = &cachedCert
		autoGenCertFingerprint = certFingerprint(&cachedCert)
		fmt.Printf("  [TLS] 已加载缓存的自动生成证书\n")
		fmt.Printf("  [TLS] 证书指纹 (SHA-256): %s\n", autoGenCertFingerprint)
		return cachedCert, nil
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("生成私钥失败: %w", err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "vexil-local"
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("生成序列号失败: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("Vexil-%s", hostname),
			Organization: []string{"Vexil Auto-Generated"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           getLocalIPs(),
		DNSNames:              []string{hostname, "localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("创建证书失败: %w", err)
	}

	os.MkdirAll(cacheDir, 0755)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("编码私钥失败: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0600)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}

	autoGenCertCache = &cert
	autoGenCertFingerprint = certFingerprint(&cert)

	fmt.Printf("  [TLS] 已自动生成自签名证书\n")
	fmt.Printf("  [TLS] 证书指纹 (SHA-256): %s\n", autoGenCertFingerprint)

	return cert, nil
}

func certFingerprint(cert *tls.Certificate) string {
	if len(cert.Certificate) == 0 {
		return ""
	}
	hash := sha256.Sum256(cert.Certificate[0])
	var formatted string
	for i, b := range hash {
		if i > 0 && i%2 == 0 {
			formatted += ":"
		}
		formatted += fmt.Sprintf("%02X", b)
	}
	return formatted
}

func GetAutoCertFingerprint() string {
	autoGenCertCacheMu.Lock()
	defer autoGenCertCacheMu.Unlock()
	return autoGenCertFingerprint
}

func getLocalIPs() []net.IP {
	var ips []net.IP
	ips = append(ips, net.IPv4(127, 0, 0, 1), net.IPv6loopback)

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					ips = append(ips, ip4)
				}
			}
		}
	}
	return ips
}