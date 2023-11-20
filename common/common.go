package common

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"log"
	"time"
)

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
