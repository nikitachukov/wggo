package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/viper"
	"log"
	"strings"
	"text/template"
	"time"
	"wggo/common"
	"wggo/mikrotikgo"
)

var ClientConfig common.MyClientConfig

//var (
//	ClientEndpointPort    = "51820"
//	ClientEndpointAddress = "gopnik.win"
//	ClientDns             = "192.168.0.254"
//)

var bindAddress string
var bindPort string
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
	Client.AddPeer(viper.GetString("wg_ifc.name"), ClientConfig.EndpointAddress, ClientConfig.Dns, allowedAddress, comment)
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
	webpeer := common.CreateWebPeer(Client.GetPeerById(c.Params("id")))
	webpeer.ClientEndpointPort = ClientConfig.EndpointPort
	webpeer.IfcPubKey = viper.GetString("wg_ifc.pub_key")

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
	webpeer.ClientEndpointPort = ClientConfig.EndpointPort
	webpeer.IfcPubKey = viper.GetString("wg_ifc.pub_key")

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

func readConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yml")    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("app.bind_address", "127.0.0.1")
	viper.SetDefault("app.bind_port", "3000")

	viper.SetDefault("client_config.endpoint_port", "51820")

	viper.SetEnvPrefix("wg")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	bindAddress = viper.GetString("app.bind_address")
	bindPort = viper.GetString("app.bind_port")

	ClientConfig.EndpointPort = viper.GetString("client_config.endpoint_port")
	ClientConfig.EndpointAddress = viper.GetString("client_config.endpoint_address")
	ClientConfig.Dns = viper.GetString("client_config.dns")

	log.Println("ready!")

}

func startApp() {
	readConfig()

	var ticker = time.NewTicker(750 * time.Millisecond)
	var quit = make(chan struct{})

	currentPeersChan = make(chan []mikrotikgo.MikrotikPeer)

	username, password, tlsConfig = common.ReadCredentialsFromVault(
		viper.GetString("router.vault_address"),
		viper.GetString("router.vault_mount_point"),
		viper.GetString("router.vault_path"),
		viper.GetString("router.vault_role_id"),
		viper.GetString("router.vault_secret_id"),
	)

	Client = mikrotikgo.MikrotikClient{
		Url:       viper.GetString("router.rest_url"),
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

	log.Fatal(app.Listen(bindAddress + ":" + bindPort))

}
