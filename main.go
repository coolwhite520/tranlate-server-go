package main

import (
	_ "translate-server/datamodels"
	"translate-server/http"
	_ "translate-server/logext"
)

func main()  {
	http.StartIntServer()
}