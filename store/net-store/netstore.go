package netstore

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/deffusion/chunkstore/digest"
	"github.com/deffusion/chunkstore/store"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"io"
	"log"
	"os"
)

const (
	service        = "/chunkservice"
	serviceAdd     = service + "/add"
	serviceExtract = service + "/extract"
)

type NetStore struct {
	p2p     host.Host
	storage *store.ChunkStore
	logger  *zap.Logger
}

func New(port int, s *store.ChunkStore, l *zap.Logger) (*NetStore, error) {
	h := makeHost(port)
	addr := h.Addrs()[0]
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", h.ID().String()))
	fmt.Println(addr.Encapsulate(hostAddr))
	return &NetStore{h, s, l.Named("netstore")}, nil
}

func generatePrivateKey() (priv crypto.PrivKey, err error) {
	r := rand.Reader
	priv, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	return
}

func makeHost(port int) host.Host {
	priv, err := generatePrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	raw, _ := priv.Raw()
	fmt.Printf("%x", raw)
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprint("/ip4/0.0.0.0/tcp/", port)),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err := dht.New(context.Background(), h)
			return idht, err
		}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return host
}

// services for other peers
func (n *NetStore) Service() {
	n.p2p.SetStreamHandler(serviceAdd, addHandler(n))
	n.p2p.SetStreamHandler(serviceExtract, extractHandler(n))
}

func digestToPeerInfo(d digest.Digest) peer.AddrInfo {
	mhbytes, err := multihash.Encode(d.Bytes(), multihash.SHA2_256)
	if err != nil {
		log.Fatal(err)
	}
	_, mh, err := multihash.MHFromBytes(mhbytes)
	if err != nil {
		log.Fatal(err)
	}
	c := cid.NewCidV1(cid.Libp2pKey, mh)
	id, _ := peer.FromCid(c)
	return peer.AddrInfo{
		ID: id,
	}
}

func (n *NetStore) Close() error {
	return n.p2p.Close()
}

func (n *NetStore) streamToPeer(d digest.Digest, service protocol.ID) (network.Stream, error) {
	pi := digestToPeerInfo(d)
	n.p2p.Connect(context.Background(), pi)
	return n.p2p.NewStream(context.Background(), pi.ID, service)
}

func (n *NetStore) Get(d digest.Digest) []digest.Digest {
	return nil
}

func (n *NetStore) Add(d digest.Digest, r io.Reader) {
	logger := n.logger.Named(fmt.Sprintf("Add(%s): ", d))
	s, err := n.streamToPeer(d, serviceAdd)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	_, err = io.Copy(s, r)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	var buff buffer.Buffer
	_, err = io.Copy(&buff, s)
	if buff.String() == d.String() {
		logger.Info("succeed")
	} else {
		logger.Sugar().Warnf(`digest mismatched: local(%s), remote(%s)`, d.String(), buff.String())
	}
}

func (n *NetStore) Extract(d digest.Digest, path string) {
	logger := n.logger.Named(fmt.Sprintf("Extract(%s, %s): ", d, path))
	s, err := n.streamToPeer(d, serviceExtract)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	_, err = s.Write([]byte(d.String()))
	if err != nil {
		logger.Error(err.Error())
		return
	}
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	_, err = io.Copy(file, s)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	logger.Info("succeed")
}

func (n *NetStore) Host() host.Host {
	return n.p2p
}
