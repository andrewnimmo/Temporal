package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/api/rtfs_cluster"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtswarm"
)

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		log.Fatal("idiot")
	}
	switch os.Args[1] {
	case "cluster":
		cm := rtfs_cluster.Initialize()
		cm.GenRestAPIConfig()
		cm.GenClient()
		cm.ParseLocalStatusAllAndSync()
		cid := cm.DecodeHashString("QmXXSSQpbYhGRMPqqZ4gF1SjqBkBjpnb44JuR1frwL1RiA")
		err := cm.Pin(cid)
		if err != nil {
			log.Fatal(err)
		}
	case "api":
		router := api.Setup()
		router.Run(":6767")
	case "swarm":
		sm, err := rtswarm.NewSwarmManager()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", sm)
	case "queue-dpa":
		qm, err := queue.Initialize(queue.DatabasePinAddQueue)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("")
		if err != nil {
			log.Fatal(err)
		}
	case "queue-dfa":
		qm, err := queue.Initialize(queue.DatabaseFileAddQueue)
		if err != nil {
			log.Fatal(err)
		}
		qm.ConsumeMessage("")
	case "queue-ipfs":
		qm, err := queue.Initialize(queue.IpfsQueue)
		if err != nil {
			log.Fatal(err)
		}
		qm.ConsumeMessage("")
	case "migrate":
		dbm := database.Initialize()
		dbm.RunMigrations()
	default:
		fmt.Println("idiot")
	}

}
