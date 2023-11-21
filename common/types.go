package common

type MySession struct {
	RequiresPassword bool `json:"requiresPassword"`
	Authenticated    bool `json:"authenticated"`
}

type MyPeer struct {
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
}
