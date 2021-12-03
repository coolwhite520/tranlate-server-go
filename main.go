package main

import (
	"translate-server/http"
	_ "translate-server/logext"
	_ "translate-server/datamodels"
)

func main()  {
	http.StartIntServer()
}