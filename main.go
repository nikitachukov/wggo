package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"wggo/common"
	"wggo/mikrotikgo"
)

var username string
var password string
var tlsConfig *tls.Config

func startApp() {
	var (
		roleId       = "697a6493-09a8-9a37-a9e3-ef8106b78507"
		secretId     = "200913ae-c711-00a8-cb94-3c1b8bca6a23"
		vaultAddress = "https://vault.gopnik.win"
		mountPoint   = "infra"
		path         = "mikrotik"
	)

	username, password, tlsConfig = common.ReadCredentialsFromVault(vaultAddress, mountPoint, path, roleId, secretId)

	app := fiber.New()
	app.Get("/hello", hello)
	app.Get("/api/session", session)
	app.Get("/api/wireguard/client", GetPeers)
	app.Post("/api/wireguard/client", AddPeer)
	app.Static("/", "www") // http://localhost:3000

	log.Fatal(app.Listen(":3000"))

}
func AddPeer(c *fiber.Ctx) error {
	payload := struct {
		Name string `json:"name"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	log.Println(payload.Name)
	peers := mikrotikgo.GetPeers(username, password, tlsConfig)
	var ips []net.IP
	for _, peer := range peers {
		if peer.Interface == "wg-in" {
			ips = append(ips, net.ParseIP(strings.Split(peer.AllowedAddress, "/")[0]))
			println(peer.AllowedAddress, peer.Interface)
		}
	}

	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})

	nextIP := common.NextIP(ips[len(ips)-1], 1)
	println(nextIP.String())

	return c.JSON(payload)
}

func main() {
	startApp()
	//privateKey, _ := wgtypes.GeneratePrivateKey()
	//presharedKey, _ := wgtypes.GenerateKey()
	//println("PresharedKey")
	//println(presharedKey.String())
	//println("PrivateKey:")
	//println(privateKey.String())
	//println("PublicKey:")
	//print(privateKey.PublicKey().String())
}

func ParseComment(commnet string) (peer common.MyPeer, err error) {
	peer = common.MyPeer{}
	err = json.Unmarshal([]byte(commnet), &peer)

	if err != nil {
		log.Println(err)
	}

	return
}

func GetPeers(c *fiber.Ctx) error {
	mikrotikPeers := mikrotikgo.GetPeers(username, password, tlsConfig)
	var _result []common.MyPeer

	for _, t := range mikrotikPeers {
		mypeer, err := ParseComment(t.Comment)
		if err != nil {
			continue
		}
		mypeer.PublicKey = t.PublicKey
		mypeer.PrivateKey = t.PrivateKey
		mypeer.PresharedKey = t.PresharedKey
		mypeer.Address = t.AllowedAddress
		disabled, _ := strconv.ParseBool(t.Disabled)
		mypeer.Enabled = !disabled

		_result = append(_result, mypeer)
	}

	result, err := json.Marshal(_result)
	if err != nil {
		panic(err)
	}
	return c.SendString(string(result))

}

func session(c *fiber.Ctx) error {
	mySession, err := json.Marshal(common.MySession{RequiresPassword: false, Authenticated: true})
	if err != nil {
		panic(err)
	}
	return c.SendString(string(mySession))
}

func hello(c *fiber.Ctx) error {

	return c.SendString("[{" +
		"\"id\":\"94924658-f969-4f4f-b70c-05bb0d370faf\"," +
		"\"name\":\"01_nikitos\"," +
		"\"enabled\":true," +
		"\"address\":\"10.8.0.1\"," +
		"\"publicKey\":\"lKAsNXJdPcrKgkM5bALoethhP8JccCkk7sBJZ0BFojg=\"," +
		"\"createdAt\":\"2023-08-20T19:32:45.497Z\"," +
		"\"updatedAt\":\"2023-09-19T20:55:26.362Z\"," +
		"\"persistentKeepalive\":null," +
		"\"latestHandshakeAt\":null," +
		"\"transferRx\":null," +
		"\"transferTx\":null" +
		//"\".id\":\"*45\"," +
		//"\"privateKey\":\"QA6MPn34BLO+70RqGB2K64D4Xivsq+rbsgIaydIvMWQ=\"," +
		//"\"presharedKey\":\"+eOW46fT37henjVVU7IK38/PJ40qMgmLS9ces3RDrdA=\"," +
		//"\"hide\":false" +
		"}]")
}
