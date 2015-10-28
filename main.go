package main

import (
	_ "google.golang.org/api/drive/v2"
)

import (
	"flag"
	"log"
)

var (
	confFile = flag.String("conf", "config.yml", "config file path")
)

func main() {
	flag.Parse()

	config, err := ParseConfig(*confFile)
	if err != nil {
		log.Println(err.Error())
		return
	}

	uploader, err := NewUploader(config)
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = uploader.Check()
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = uploader.Prepare()
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("start upload")
	err = uploader.Run()
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("done")
}
