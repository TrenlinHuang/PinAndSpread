package netstore

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/deffusion/chunkstore/digest"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"log"
	"testing"
)

func TestID(t *testing.T) {
	d, err := digest.New("s816f1bac92c8d67ca6896086b259b14cd9c2d23dc7f86122d12280aa8936b1b14")
	if err != nil {
		log.Fatal(err)
	}

	pi := digestToPeerInfo(d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("id:", pi.ID)
}

func TestKeyFile(t *testing.T) {
	//var priv crypto.PrivKey
	//privFile, err := os.Open(store.ConfRoot + ".priv")
	//defer privFile.Close()
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		fmt.Println("not exist", err)
	//		privFile, err = os.Create(store.ConfRoot + ".priv")
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//	} else {
	//		log.Fatal(err)
	//	}
	//} else { // 文件存在
	//	var buff buffer.Buffer
	//	_, err = io.Copy(&buff, privFile)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	priv, err = crypto.UnmarshalPrivateKey(buff.Bytes())
	//	// 文件存在但私钥格式不正确
	//	if err != nil {
	//		fmt.Println("invalid private key detected", err)
	//		priv, err = generatePrivateKey()
	//		f, err := os.OpenFile(store.ConfRoot+".priv", os.O_RDWR, 0777)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//	}
	//}

}

func TestKey(t *testing.T) {

	//if _, err := os.Stat(store.ConfRoot + ".priv"); err != nil {
	//	if os.IsNotExist(err) {
	//		priv, err = generatePrivateKey()
	//		log.Fatal(err)
	//	}
	//} else {
	//
	//}
	r := rand.Reader
	priv, pub, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	//fmt.Println("pub", pub)
	//raw, _ := priv.Raw()
	raw, _ := crypto.MarshalPrivateKey(priv)
	//fmt.Printf("%x", raw)
	priv2, err := crypto.UnmarshalPrivateKey(raw)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("pub2", priv2.GetPublic().Equals(pub))
}

func TestDial(t *testing.T) {
	host, _ := New(3001, nil, nil)
	maddr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3000/p2p/QmQcfJGVmVvBgvQGVF7ss5JXqeWYxvQnNYaNykabBjssQi")

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatal("addr info:", err)
	}
	h := host.Host()
	fmt.Println("addrs:", info.Addrs)
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	s, err := h.NewStream(context.Background(), info.ID, "/chunkservice")
	if err != nil {
		log.Fatal("new stream:", err)
	}
	_, err = s.Write([]byte("hello world\n"))
}
