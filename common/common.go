package common

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"log"
	"net"
	"time"
)

func NextIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3)
}

func ReadCredentialsFromVault(vaultAddress, mountPath, path, roleId, secretId string) (string, string, *tls.Config) {
	ctx := context.Background()
	client, err := vault.New(vault.WithAddress(vaultAddress), vault.WithRequestTimeout(30*time.Second))

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Auth.AppRoleLogin(ctx, schema.AppRoleLoginRequest{RoleId: roleId, SecretId: secretId})

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		log.Fatal(err)
	}

	s, err := client.Secrets.KvV2Read(ctx, path, vault.WithMountPath(mountPath))
	if err != nil {
		log.Fatal(err)
	}

	cert, err := tls.X509KeyPair([]byte(s.Data.Data["cert"].(string)), []byte(s.Data.Data["key"].(string)))
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(s.Data.Data["ca"].(string)))

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
	}

	return s.Data.Data["username"].(string), s.Data.Data["password"].(string), tlsConfig
}
