package mikrotikgo

type Peer struct {
	ID                     string `json:".id"`
	AllowedAddress         string `json:"allowed-address"`
	ClientEndpoint         string `json:"client-endpoint"`
	CurrentEndpointAddress string `json:"current-endpoint-address"`
	CurrentEndpointPort    string `json:"current-endpoint-port"`
	Disabled               string `json:"disabled"`
	Dynamic                string `json:"dynamic"`
	EndpointAddress        string `json:"endpoint-address"`
	EndpointPort           string `json:"endpoint-port"`
	Interface              string `json:"interface"`
	LastHandshake          string `json:"last-handshake,omitempty"`
	PresharedKey           string `json:"preshared-key"`
	PrivateKey             string `json:"private-key"`
	PublicKey              string `json:"public-key"`
	Rx                     string `json:"rx"`
	Tx                     string `json:"tx"`
	ClientAddress          string `json:"client-address,omitempty"`
	Comment                string `json:"comment,omitempty"`
}
