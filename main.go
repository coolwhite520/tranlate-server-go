package main

import (
	_ "translate-server/datamodels"
	"translate-server/http"
	_ "translate-server/logext"
	"translate-server/rpc"
)

func main()  {
	rpc.StopAllServer()
	rpc.StartTikaServer()
	http.StartMainServer()
}
