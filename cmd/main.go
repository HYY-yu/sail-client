package main

import (
	"log"

	"github.com/HYY-yu/seckill.pkg/pkg/shutdown"

	sailclient "github.com/HYY-yu/sail-client"
)

func main() {
	sail := sailclient.NewWithToml("./cfg.toml",
		sailclient.WithOnConfigChange(func(configFileKey string, s *sailclient.Sail) {
			log.Println("find key change - ", configFileKey)

			log.Println("new value: ", s.MustGetString("test.log"))
		}))
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
