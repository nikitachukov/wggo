package common

type MySession struct {
	RequiresPassword bool `json:"requiresPassword"`
	Authenticated    bool `json:"authenticated"`
}
