package main

import (
	_ "translate-server/datamodels"
	"translate-server/http"
	_ "translate-server/logext"
	"translate-server/rpc"
)

func main()  {
	rpc.ImportAllImages()
	rpc.StopAllRunningDockers()
	// qidong
	http.StartMainServer()
}
