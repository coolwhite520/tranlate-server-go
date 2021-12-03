package activation

import (
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"log"
)

func getMachineId() {
	id, err := machineid.ProtectedID("myAppName")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}