package main

import (
	"translate-server/http"
	_ "translate-server/logext"
	_ "translate-server/models"
)

func main()  {
	http.StartIntServer()
}