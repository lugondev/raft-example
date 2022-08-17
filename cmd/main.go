package main

import (
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
	"log"
	"os"
	"ysf/raftsample/fsm"
	"ysf/raftsample/server"
	"ysf/raftsample/types"
)

const (
	serverPort = "SERVER_PORT"

	raftNodeId = "RAFT_NODE_ID"
	raftPort   = "RAFT_PORT"
	raftVolDir = "RAFT_VOL_DIR"
)

var confKeys = []string{
	serverPort,

	raftNodeId,
	raftPort,
	raftVolDir,
}

// main entry point of application start
// run using CONFIG=config.yaml ./program
func main() {

	var v = viper.New()
	v.AutomaticEnv()
	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	conf := types.Config{
		Server: types.ConfigServer{
			Port: v.GetInt(serverPort),
		},
		Raft: types.ConfigRaft{
			NodeId:    v.GetString(raftNodeId),
			Port:      v.GetInt(raftPort),
			VolumeDir: v.GetString(raftVolDir),
		},
	}

	log.Printf("%+v\n", conf)

	// Preparing badgerDB
	badgerOpt := badger.DefaultOptions(conf.Raft.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()

	raftServer, _ := fsm.NewRaft(badgerDB, &conf)

	srv := server.New(fmt.Sprintf(":%d", conf.Server.Port), badgerDB, raftServer)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	return
}
