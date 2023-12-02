package common

type MySession struct {
	RequiresPassword bool `json:"requiresPassword"`
	Authenticated    bool `json:"authenticated"`
}

type WebPeer struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Enabled             bool   `json:"enabled"`
	Address             string `json:"address"`
	PublicKey           string `json:"publicKey"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
	PersistentKeepalive any    `json:"persistentKeepalive"`
	LatestHandshakeAt   any    `json:"latestHandshakeAt"`
	TransferRx          any    `json:"transferRx"`
	TransferTx          any    `json:"transferTx"`
	PrivateKey          string `json:"privateKey"`
	PresharedKey        string `json:"presharedKey"`
	Hide                bool   `json:"hide"`
	ClientEndpointPort  int    `json:"ClientEndpointPort,omitempty"`
	IfcPubKey           string `json:"IfcPubKey,omitempty"`
	ClientEndpoint      string `json:"ClientEndpoint,omitempty"`
	ClientDNS           string `json:"ClientDNS,omitempty"`
}

type Comment struct {
	Name      string `json:"name"`
	Hide      bool   `json:"hide,omitempty"`
	Easy      bool   `json:"easy,omitempty"`
	ID        string `json:"id"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}
