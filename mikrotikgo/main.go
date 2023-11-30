package mikrotikgo

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
	"wggo/common"
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

	q := req.URL.Query()
	q.Add("interface", "wg-in")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	data, err := io.ReadAll(resp.Body)

	var targets []mikrotikgo.Peer

	err = json.Unmarshal(data, &targets)
	if err != nil {
		log.Fatal(err)
	}

	return targets

}

func AddPeers(username string, password string, tlsConfig *tls.Config, name string, ifc string) string {

	var ips []net.IP
	for _, peer := range GetPeers(username, password, tlsConfig) {
		if peer.Interface == "wg-in" {
			ips = append(ips, net.ParseIP(strings.Split(peer.AllowedAddress, "/")[0]))
		}
	}
	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})
	allowedAddress := common.NextIP(ips[len(ips)-1], 1).String()

	privateKey, _ := wgtypes.GeneratePrivateKey()
	presharedKey, _ := wgtypes.GenerateKey()
	pubKey := privateKey.PublicKey()

	type Comment struct {
		Name      string `json:"name"`
		Hide      bool   `json:"hide,omitempty"`
		Easy      bool   `json:"easy,omitempty"`
		ID        string `json:"id"`
		UpdatedAt string `json:"updatedAt,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
	}

	commentBuffer, _ := json.Marshal(Comment{ID: uuid.Must(uuid.NewRandom()).String(), Name: name, CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)})

	peer := &mikrotikgo.Peer{
		ClientAddress:  allowedAddress + "/32",
		PublicKey:      pubKey.String(),
		PrivateKey:     privateKey.String(),
		PresharedKey:   presharedKey.String(),
		Comment:        string(commentBuffer),
		AllowedAddress: allowedAddress + "/32",
		Interface:      ifc,
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

func SetPeerState(username, password string, tlsConfig *tls.Config, id string, enable bool) int {
	var verb string
	if enable == true {
		verb = "enable"
	} else {
		verb = "disable"
	}

	peers := make(map[string]mikrotikgo.Peer)

	for _, peer := range GetPeers(username, password, tlsConfig) {

		var comment struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}

		err := json.Unmarshal([]byte(peer.Comment), &comment)
		if err != nil {
			continue
		}

		peers[comment.Id] = peer
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(struct {
		Numbers string `json:"numbers"`
	}{Numbers: peers[id].MikrotikID})
	if err != nil {
	}

	req, _ := http.NewRequest("POST", "https://router.gopnik.win/rest/interface/wireguard/peers/"+verb, payload)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	resp, _ := client.Do(req)
	return resp.StatusCode

}

func DeletePeer(username, password string, tlsConfig *tls.Config, id string) int {

	peers := make(map[string]mikrotikgo.Peer)

	for _, peer := range GetPeers(username, password, tlsConfig) {

		var comment struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}

		err := json.Unmarshal([]byte(peer.Comment), &comment)
		if err != nil {
			continue
		}

		peers[comment.Id] = peer
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(struct {
		Numbers string `json:"numbers"`
	}{Numbers: peers[id].MikrotikID})
	if err != nil {
	}

	req, _ := http.NewRequest("DELETE", "https://router.gopnik.win/rest/interface/wireguard/peers/"+peers[id].MikrotikID, nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	resp, _ := client.Do(req)
	return resp.StatusCode

}
