package controller

import (
	"github.com/kataras/neffos"
	"log"
)

type WebsocketController struct {
	*neffos.NSConn `stateless:"true"`
	Namespace string
}

func (c *WebsocketController) OnNamespaceConnected(msg neffos.Message) error {
	log.Println(msg)
	return nil
}

func (c *WebsocketController) OnNamespaceDisconnect(msg neffos.Message) error {
	log.Println(msg)
	return nil
}

func (c *WebsocketController) OnChat(msg neffos.Message) error {
	log.Println(msg)
	return nil
}
