package main

import (
	"log"

	"github.com/HYY-yu/seckill.pkg/pkg/shutdown"

	sailclient "github.com/HYY-yu/sail-client"
)

func main() {
	sail := sailclient.NewWithToml("./cfg.toml")
	if sail.Err() != nil {
		log.Fatalln(sail.Err())
	}

	err := sail.Pull()
	if err != nil {
		log.Fatalln(err)
	}

	// 监听信号
	shutdown.NewHook().Close(
		func() {
			err := sail.Close()
			if err != nil {
				log.Println(err)
			}
		},
	)
}
