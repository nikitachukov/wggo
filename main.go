package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
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
	app.Get("/api/session", session)
	app.Get("/api/wireguard/client", GetPeers)
	app.Post("/api/wireguard/client", AddPeer)
	app.Post("/api/wireguard/client/:id/disable", DisablePeer)
	app.Post("/api/wireguard/client/:id/enable", EnablePeer)
	app.Delete("/api/wireguard/client/:id", DeletePeer)
	app.Static("/", "www")

	log.Fatal(app.Listen(":3000"))

}

func DeletePeer(c *fiber.Ctx) error {
	statusCode := mikrotikgo.DeletePeer(username, password, tlsConfig, c.Params("id"))
	if statusCode == 204 {
		return c.Status(fiber.StatusNoContent).SendString("")
	} else {
		return c.Status(fiber.StatusInternalServerError).SendString("")
	}
}
func DisablePeer(c *fiber.Ctx) error {
	statusCode := mikrotikgo.SetPeerState(username, password, tlsConfig, c.Params("id"), false)
	if statusCode == 200 {
		return c.Status(fiber.StatusNoContent).SendString("")
	} else {
		return c.Status(fiber.StatusInternalServerError).SendString("")
	}
}

func EnablePeer(c *fiber.Ctx) error {
	statusCode := mikrotikgo.SetPeerState(username, password, tlsConfig, c.Params("id"), true)
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

	mikrotikgo.AddPeers(username, password, tlsConfig, payload.Name, "wg-in")
	return c.JSON(payload)
}

func main() {
	startApp()
}

func GetPeers(c *fiber.Ctx) error {
	mikrotikPeers := mikrotikgo.GetPeers(username, password, tlsConfig)
	var _result []common.MyPeer

	for _, t := range mikrotikPeers {
		mypeer := common.ParseComment(t.Comment)
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
