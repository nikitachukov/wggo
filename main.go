package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"log"
	"text/template"
	"time"
	"wggo/common"
	"wggo/mikrotikgo"
)

var (
	ClientEndpointPort = 51820
	IfcPubKey          = "uOQzUkEBJAyQWH5LopDUmz3k95+oAddf+hHLQYzoLBo="
)

var username string
var password string
var tlsConfig *tls.Config
var quit chan struct{}

var Client mikrotikgo.MikrotikClient

var currentPeersChan chan []mikrotikgo.MikrotikPeer

func DeletePeer(c *fiber.Ctx) error {
	statusCode := Client.DeletePeer(Client.GetPeerById(c.Params("id")))
	if statusCode == 204 {
		return c.Status(fiber.StatusNoContent).SendString("")
	} else {
		return c.Status(fiber.StatusInternalServerError).SendString("")
	}
}

func DisablePeer(c *fiber.Ctx) error {
	statusCode := Client.SetPeerState(Client.GetPeerById(c.Params("id")), false)
	if statusCode == 200 {
		return c.Status(fiber.StatusNoContent).SendString("")
	} else {
		return c.Status(fiber.StatusInternalServerError).SendString("")
	}
}

func EnablePeer(c *fiber.Ctx) error {
	statusCode := Client.SetPeerState(Client.GetPeerById(c.Params("id")), true)
	if statusCode == 200 {
		return c.Status(fiber.StatusNoContent).SendString("")
	} else {
		return c.Status(fiber.StatusInternalServerError).SendString("")
	}
}

func AddPeer(c *fiber.Ctx) error {
	payload := struct {
		Name string `json:"name"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return err
	}
	comment := common.CreateNewComment(payload.Name)
	allowedAddress := common.GetNextPeerIp(<-currentPeersChan)
	Client.AddPeer("wg-in", "gopnik.win", "192.168.0.254", allowedAddress, comment)
	return c.JSON(payload)
}

func GetWebPeers(c *fiber.Ctx) error {
	mikrotikPeers := <-currentPeersChan
	var _result []common.WebPeer
	for _, t := range mikrotikPeers {
		Peer := common.CreateWebPeer(t)
		if (Peer.ID != "") && (Peer.Hide == false) {
			_result = append(_result, Peer)
		}
	}

	result, err := json.Marshal(_result)
	if err != nil {
		panic(err)
	}
	return c.SendString(string(result))

}

func Session(c *fiber.Ctx) error {
	mySession, err := json.Marshal(common.MySession{RequiresPassword: false, Authenticated: true})
	if err != nil {
		panic(err)
	}
	return c.SendString(string(mySession))
}

func Configuration(c *fiber.Ctx) error {
	ClientEndpointPort := 51820
	IfcPubKey := "uOQzUkEBJAyQWH5LopDUmz3k95+oAddf+hHLQYzoLBo="

	webpeer := common.CreateWebPeer(Client.GetPeerById(c.Params("id")))
	webpeer.ClientEndpointPort = ClientEndpointPort
	webpeer.IfcPubKey = IfcPubKey

	c.Append("content-disposition", "attachment; filename=\""+webpeer.Name+".conf\"")
	c.Append("content-type", "text/plain; charset=utf-8")

	return c.SendString(string(configFromPeer(webpeer)))
}

func configFromPeer(webpeer common.WebPeer) (config []byte) {
	templateStr := `[Interface]
ListenPort = {{.ClientEndpointPort}}
PrivateKey = {{.PrivateKey}}
Address = {{.Address}}
DNS = {{.ClientDNS}}

[Peer]
PublicKey = {{.IfcPubKey}}
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = {{.ClientEndpoint}}:{{.ClientEndpointPort}}
PresharedKey = {{.PresharedKey}}
`

	buf := new(bytes.Buffer)
	tmpl, err := template.New("test").Parse(templateStr)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, webpeer)
	if err != nil {
		panic(err)
	}

	config = buf.Bytes()

	return
}

func GetQRCode(c *fiber.Ctx) error {

	webpeer := common.CreateWebPeer(Client.GetPeerById(c.Params("id")))
	webpeer.ClientEndpointPort = ClientEndpointPort
	webpeer.IfcPubKey = IfcPubKey

	c.Append("content-disposition", "inline; filename=qrcode.svg")
	c.Append("content-type", "image/png; charset=utf-8")

	var png []byte
	png, err := qrcode.Encode(string(configFromPeer(webpeer)), qrcode.Highest, 512)
	if err != nil {
		panic(err)
	}

	return c.Send(png)

}

func main() {
	startApp()
}

func startApp() {

	var (
		roleId       = "697a6493-09a8-9a37-a9e3-ef8106b78507"
		secretId     = "200913ae-c711-00a8-cb94-3c1b8bca6a23"
		vaultAddress = "https://vault.gopnik.win"
		mountPoint   = "infra"
		path         = "mikrotik"
		ticker       = time.NewTicker(750 * time.Millisecond)
		quit         = make(chan struct{})
	)

	currentPeersChan = make(chan []mikrotikgo.MikrotikPeer)

	username, password, tlsConfig = common.ReadCredentialsFromVault(vaultAddress, mountPoint, path, roleId, secretId)

	Client = mikrotikgo.MikrotikClient{
		Url:       "https://router.gopnik.win/rest/",
		Login:     username,
		Password:  password,
		TlsConfig: tlsConfig,
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				currentPeersChan <- Client.GetPeers()

				log.Println("fetching data from mikrotik")

			case <-quit:
				log.Println("stop")
				ticker.Stop()
				//return
			}
		}
	}()

	app := fiber.New()
	app.Get("/api/session", Session)
	app.Get("/api/wireguard/client", GetWebPeers)
	app.Post("/api/wireguard/client", AddPeer)
	app.Post("/api/wireguard/client/:id/disable", DisablePeer)
	app.Post("/api/wireguard/client/:id/enable", EnablePeer)
	app.Delete("/api/wireguard/client/:id", DeletePeer)
	app.Get("/api/wireguard/client/:id/configuration", Configuration)
	app.Get("/api/wireguard/client/:id/qrcode.svg", GetQRCode)
	app.Static("/", "www")

	log.Fatal(app.Listen(":3000"))

}
