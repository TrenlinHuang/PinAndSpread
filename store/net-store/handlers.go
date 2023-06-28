package netstore

import (
	"fmt"
	"github.com/deffusion/chunkstore/digest"
	"github.com/libp2p/go-libp2p/core/network"
	store2 "github.com/trenlinhuang/pin-spread/store"
	"go.uber.org/zap/buffer"
	"io"
	"os"
)

func addHandler(n *NetStore) network.StreamHandler {
	loggeer := n.logger.Named("addHandler")
	return func(s network.Stream) {
		defer s.Close()
		d, err := n.storage.Add(s)
		if err != nil {
			loggeer.Fatal(err.Error())
		}
		_, err = s.Write([]byte(d.String()))
		if err != nil {
			loggeer.Fatal(err.Error())
		}
	}
}

func extractHandler(n *NetStore) network.StreamHandler {
	loggeer := n.logger.Named("extractHandler").Sugar()
	return func(s network.Stream) {
		defer s.Close()
		var buff buffer.Buffer
		io.Copy(&buff, s)
		di, err := digest.New(buff.String())
		if err != nil {
			loggeer.Fatal("digest.New", err)
		}
		ds, err := n.storage.Get(di)
		for _, d := range ds {
			path := fmt.Sprint(store2.ChunkRoot, d.String())
			f, err := os.Open(path)
			if err != nil {
				loggeer.Fatalf("os.Open(%s): %s", path, err)
			}
			_, err = io.Copy(s, f)
			f.Close()
			if err != nil {
				loggeer.Fatal("io.Copy: %s", err)
			}
		}
	}
}
