package mikrotikgo

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io"
	"log"
	"net/http"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GetPeers(username, password string, tlsConfig *tls.Config) []MikrotikPeer {
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", "https://router.gopnik.win/rest/interface/wireguard/peers", nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))

	q := req.URL.Query()
	q.Add("interface", "wg-in")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	data, err := io.ReadAll(resp.Body)

	var targets []MikrotikPeer

	err = json.Unmarshal(data, &targets)
	if err != nil {
		log.Fatal(err)
	}

	return targets

}

func AddPeers(username string, password string, tlsConfig *tls.Config, ifc string, clientEndpoint string, clientDns string, allowedAddress, comment string) string {

	privateKey, _ := wgtypes.GeneratePrivateKey()
	presharedKey, _ := wgtypes.GenerateKey()
	pubKey := privateKey.PublicKey()

	peer := &MikrotikPeer{
		ClientAddress:  allowedAddress + "/32",
		PublicKey:      pubKey.String(),
		PrivateKey:     privateKey.String(),
		PresharedKey:   presharedKey.String(),
		Comment:        comment,
		AllowedAddress: allowedAddress + "/32",
		Interface:      ifc,
		ClientEndpoint: clientEndpoint,
		ClientDNS:      clientDns,
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(peer)
	if err != nil {
		return ""
	}

	req, _ := http.NewRequest("POST", "https://router.gopnik.win/rest/interface/wireguard/peers/add", body)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	resp, _ := client.Do(req)

	data, _ := io.ReadAll(resp.Body)

	println(string(data[:]))

	return username
}

func SetPeerState(username, password string, tlsConfig *tls.Config, peer MikrotikPeer, enable bool) int {
	var verb string
	if enable == true {
		verb = "enable"
	} else {
		verb = "disable"
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(struct {
		Numbers string `json:"numbers"`
	}{Numbers: peer.ID})
	if err != nil {
	}

	req, _ := http.NewRequest("POST", "https://router.gopnik.win/rest/interface/wireguard/peers/"+verb, payload)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	resp, _ := client.Do(req)
	return resp.StatusCode

}

func DeletePeer(username, password string, tlsConfig *tls.Config, peer MikrotikPeer) int {
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("DELETE", "https://router.gopnik.win/rest/interface/wireguard/peers/"+peer.ID, nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	resp, _ := client.Do(req)
	return resp.StatusCode

}
