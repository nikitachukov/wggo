package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
	"wggo/mikrotikgo"
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

func CreateNewComment(name string) string {
	commentBuffer, _ := json.Marshal(mikrotikgo.Comment{ID: uuid.Must(uuid.NewRandom()).String(), Name: name, CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)})
	return string(commentBuffer)
}

func GetNextPeerIp(peers []mikrotikgo.MikrotikPeer) (allowedAddress string) {

	var ips []net.IP
	for _, peer := range peers {
		if peer.Interface == "wg-in" {
			ips = append(ips, net.ParseIP(strings.Split(peer.AllowedAddress, "/")[0]))
		}
	}
	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})
	allowedAddress = NextIP(ips[len(ips)-1], 1).String()

	return

}

func ParseComment(comment string) (CommentValue struct {
	Name      string `json:"name"`
	Hide      bool   `json:"hide,omitempty"`
	Easy      bool   `json:"easy,omitempty"`
	ID        string `json:"id"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}, err error) {

	err = json.Unmarshal([]byte(comment), &CommentValue)

	if err != nil {
	}

	return
}

func CreateWebPeer(MikrotikPeer mikrotikgo.MikrotikPeer) (Peer WebPeer) {
	Peer.PrivateKey = MikrotikPeer.PrivateKey
	Peer.PublicKey = MikrotikPeer.PublicKey
	Peer.PresharedKey = MikrotikPeer.PresharedKey
	Peer.Enabled, _ = strconv.ParseBool(MikrotikPeer.Disabled)
	Peer.Enabled = !Peer.Enabled
	Peer.Address = MikrotikPeer.AllowedAddress
	Peer.ClientEndpoint = MikrotikPeer.ClientEndpoint
	Peer.ClientDNS = MikrotikPeer.ClientDNS
	Peer.TransferRx = MikrotikPeer.Rx
	Peer.TransferTx = MikrotikPeer.Tx

	comment, err := ParseComment(MikrotikPeer.Comment)
	if err == nil {
		Peer.ID = comment.ID
		Peer.Name = comment.Name
		Peer.CreatedAt = comment.CreatedAt
		Peer.UpdatedAt = comment.UpdatedAt
		Peer.Hide = comment.Hide

	}

	return

}

func SplitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

func microtikAtToS(s string) int32 {
	if len(s) > 0 {
		words := SplitAny(s, "dhms")
		log.Println(s, "-->", strings.Join(words, " "))
	}
	return 0
}
