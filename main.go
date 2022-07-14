package main

import (
	"context"
	"github.com/Nerzal/gocloak/v11"
)

func main() {
	client := gocloak.NewClient("https://kc.scb-monitor.ru")
	login, _ := client.Login(context.Background(), "node-auth", "280e97b8-447f-4492-8243-362fc1108b79", "master", "testuser", "testuser")
	print("\"" + login.AccessToken + "\"")
}
