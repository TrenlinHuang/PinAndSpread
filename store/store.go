package store

import (
	"fmt"
	"log"
	"os"
	"os/user"
)

var storeRoot string
var ChunkRoot string
var NamingRoot string // maps filename to digest
var ConfRoot string

func init() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	storeRoot = fmt.Sprint(currentUser.HomeDir, "/pin-spread/")
	ChunkRoot = fmt.Sprint(storeRoot, "chunks/")
	NamingRoot = fmt.Sprint(storeRoot, "naming/")
	ConfRoot = fmt.Sprint(storeRoot, "conf/")
}

func init() {
	makeDirIfNotExist(ChunkRoot)
	makeDirIfNotExist(NamingRoot)
	makeDirIfNotExist(ConfRoot)
}

func makeDirIfNotExist(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("directory %s is not exist\n", path)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("directory %s was created\n", path)
	}
}
