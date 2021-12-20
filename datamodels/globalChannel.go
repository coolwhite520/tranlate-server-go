package datamodels

import "os"

var GlobalChannel = make(chan os.Signal, 1)
