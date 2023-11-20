package mikrotikgo

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"wggo/mikrotikgo/types"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GetPeers(username, password string, tlsConfig *tls.Config) []mikrotikgo.Peer {
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", "https://router.gopnik.win/rest/interface/wireguard/peers", nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))

	resp, err := client.Do(req)

	data, err := io.ReadAll(resp.Body)

	targets := []mikrotikgo.Peer{}

	err = json.Unmarshal(data, &targets)
	if err != nil {
		log.Fatal(err)
	}

	return targets

}
