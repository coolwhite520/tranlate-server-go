package controller

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/neffos"
	"log"
	"os"
)

type WebsocketController struct {
	Conn *neffos.NSConn
	Namespace string
	FilePoint *os.File
}

type ReqStruct struct {
	FileName string `json:"file_name"`
	Data []byte `json:"data"`
	Size int64 `json:"size"`
}

func (c *WebsocketController) OnNamespaceConnected(msg neffos.Message) error {
	log.Println(c.Conn.String())
	log.Println(string(msg.Body))
	return nil
}

func (c *WebsocketController) OnNamespaceDisconnect(msg neffos.Message) error {
	log.Println(string(msg.Body))
	return nil
}


func (c *WebsocketController) NewFile(msg neffos.Message) error {
	var req ReqStruct
	err := json.Unmarshal(msg.Body, &req)
	if err != nil {
		return err
	}
	c.FilePoint, err = os.Create(req.FileName)
	if err != nil {
		neffos.Reply([]byte(err.Error()))
	}
	return neffos.Reply([]byte("success"))
}

func (c *WebsocketController) WriteFile(msg neffos.Message) error {
	//bytes, _ := base64.StdEncoding.DecodeString(string(msg.Body))
	n, err := c.FilePoint.Write(msg.Body)
	if err != nil {
		return err
	}
	return neffos.Reply([]byte(fmt.Sprintf("success write %d byte", n)))
}
func (c *WebsocketController) CloseFile(msg neffos.Message) error {
	err := c.FilePoint.Close()
	if err != nil {
		return err
	}
	return neffos.Reply([]byte(fmt.Sprintf("success save")))
}