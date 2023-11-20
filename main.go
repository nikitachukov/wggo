package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"wggo/common"
	"wggo/mikrotikgo"
)

var username string
var password string
var tlsConfig *tls.Config

func main() {

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
	app.Get("/api/wireguard/client", client)
	app.Static("/", "www") // http://localhost:3000

	log.Fatal(app.Listen(":3000"))
}

func hello(c *fiber.Ctx) error {

	for _, t := range mikrotikgo.GetPeers(username, password, tlsConfig) {
		fmt.Println(t.ID, "-", t.AllowedAddress)
	}

	return c.SendString("ok")
}

func session(c *fiber.Ctx) error {
	mySession, err := json.Marshal(common.MySession{RequiresPassword: false, Authenticated: true})

	if err != nil {
		panic(err)
	}

	return c.SendString(string(mySession))
}

func client(c *fiber.Ctx) error {

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
