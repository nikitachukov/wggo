package mikrotikgo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io"
	"log"
	"net/http"
)

func (c *MikrotikClient) GetPeers() []MikrotikPeer {
	transport := &http.Transport{TLSClientConfig: c.TlsConfig}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", c.Url+"interface/wireguard/peers", nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Login+":"+c.Password)))

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

func (c *MikrotikClient) AddPeer(ifc string, clientEndpoint string, clientDns string, allowedAddress, comment string) string {

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

	transport := &http.Transport{TLSClientConfig: c.TlsConfig}
	client := &http.Client{Transport: transport}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(peer)
	if err != nil {
		return ""
	}

	req, _ := http.NewRequest("POST", c.Url+"interface/wireguard/peers/add", body)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Login+":"+c.Password)))
	resp, _ := client.Do(req)

	data, _ := io.ReadAll(resp.Body)

	return string(data[:])
}

func (c *MikrotikClient) SetPeerState(peer MikrotikPeer, enable bool) int {
	var verb string
	if enable == true {
		verb = "enable"
	} else {
		verb = "disable"
	}

	transport := &http.Transport{TLSClientConfig: c.TlsConfig}
	client := &http.Client{Transport: transport}

	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(struct {
		Numbers string `json:"numbers"`
	}{Numbers: peer.ID})
	if err != nil {
	}

	req, _ := http.NewRequest("POST", c.Url+"interface/wireguard/peers/"+verb, payload)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Login+":"+c.Password)))
	resp, _ := client.Do(req)
	return resp.StatusCode

}

func (c *MikrotikClient) DeletePeer(peer MikrotikPeer) int {
	transport := &http.Transport{TLSClientConfig: c.TlsConfig}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest("DELETE", c.Url+"interface/wireguard/peers/"+peer.ID, nil)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Login+":"+c.Password)))
	resp, _ := client.Do(req)
	return resp.StatusCode
}

func (c *MikrotikClient) ParseComment(comment string) (CommentValue Comment, err error) {
	err = json.Unmarshal([]byte(comment), &CommentValue)
	if err != nil {
	}
	return
}

func (c *MikrotikClient) GetPeerById(id string) MikrotikPeer {

	peers := c.GetPeers()

	peersBuf := make(map[string]MikrotikPeer)
	for _, peer := range peers {
		comment, _ := c.ParseComment(peer.Comment)
		peersBuf[comment.ID] = peer
	}
	return peersBuf[id]
}
