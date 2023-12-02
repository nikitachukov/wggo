package mikrotikgo

type MikrotikPeer struct {
	ID                     string `json:".id,omitempty"`
	AllowedAddress         string `json:"allowed-address"`
	ClientAddress          string `json:"client-address"`
	ClientDNS              string `json:"client-dns,omitempty"`
	ClientEndpoint         string `json:"client-endpoint,omitempty"`
	Comment                string `json:"comment,omitempty"`
	CurrentEndpointAddress string `json:"current-endpoint-address,omitempty"`
	CurrentEndpointPort    string `json:"current-endpoint-port,omitempty"`
	Disabled               string `json:"disabled,omitempty"`
	Dynamic                string `json:"dynamic,omitempty"`
	EndpointAddress        string `json:"endpoint-address,omitempty"`
	EndpointPort           string `json:"endpoint-port,omitempty"`
	Interface              string `json:"interface"`
	LastHandshake          string `json:"last-handshake,omitempty"`
	PresharedKey           string `json:"preshared-key"`
	PrivateKey             string `json:"private-key"`
	PublicKey              string `json:"public-key"`
	Rx                     string `json:"rx,omitempty"`
	Tx                     string `json:"tx,omitempty"`
}
