package mikrotikgo

type Peer struct {
	MikrotikID             string `json:".id,omitempty"`
	AllowedAddress         string `json:"allowed-address"`
	ClientEndpoint         string `json:"client-endpoint"`
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
	ClientAddress          string `json:"client-address,omitempty"`
	Comment                string `json:"comment,omitempty"`
}
