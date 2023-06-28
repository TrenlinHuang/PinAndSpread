package storage

import (
	"bufio"
	"github.com/deffusion/chunkstore/digest"
	"github.com/deffusion/chunkstore/store"
	"github.com/deffusion/chunkstore/store/kv"
	"github.com/deffusion/chunkstore/store/kv/level_kv"
	"github.com/pkg/errors"
	store2 "github.com/trenlinhuang/pin-spread/store"
	netstore "github.com/trenlinhuang/pin-spread/store/net-store"
	"go.uber.org/zap"
	"os"
)

type Service struct {
	localStore store.Store
	netStore   *netstore.NetStore
	naming     kv.KV // map the filename to hash
	logger     *zap.Logger
}

func NewService() *Service {
	zaplogger, _ := zap.NewDevelopment()
	l := zaplogger.Named("Service")
	logger := l.Named("NewService")
	dbLocal, err := level_kv.New(store.KVRoot)
	if err != nil {
		logger.Info(err.Error())
		return nil
	}
	lcs := store.New(dbLocal, store.ChunkRoot, l.Named("ChunkStore(local)"))
	dbNet, err := level_kv.New(store2.NamingRoot)
	if err != nil {
		logger.Info(err.Error())
		return nil
	}
	ncs := store.New(dbNet, store2.ChunkRoot, l.Named("ChunkStore(net)"))
	ns, err := netstore.New(3000, ncs, l.Named("NetStore"))
	if err != nil {
		logger.Info(err.Error())
		return nil
	}
	ns.Service()
	return &Service{
		lcs,
		ns,
		dbNet,
		l,
	}
}

func (s *Service) AddFile(f *os.File) error {
	errMessage := "Service.AddFile"
	d, err := s.localStore.Add(bufio.NewReader(f))
	if err != nil {
		return errors.WithMessage(err, errMessage)
	}
	err = s.naming.Put([]byte(f.Name()), []byte(d.String()))
	if err != nil {
		return errors.WithMessage(err, errMessage)
	}
	s.netStore.Add(d, bufio.NewReader(f))
	return nil
}

func (s *Service) ExtractFile(filename, path string) error {
	errMessage := "Service.ExtractFile"
	digestBytes, err := s.naming.Get([]byte(filename))
	if err != nil {
		return errors.WithMessage(err, errMessage)
	}
	d, err := digest.New(string(digestBytes))
	if err != nil {
		return errors.WithMessage(err, errMessage)
	}
	err = s.localStore.Extract(d, path)
	if err != nil {
		return errors.WithMessage(err, errMessage)
	}
	return nil
}

func (s *Service) Close() error {
	errMessage := "Service.Close"
	err := s.localStore.Close()
	if err != nil {
		s.logger.Error(errors.WithMessage(err, errMessage).Error())
	}
	err = s.naming.Close()
	if err != nil {
		s.logger.Error(errors.WithMessage(err, errMessage).Error())
	}
	err = s.netStore.Close()
	if err != nil {
		s.logger.Error(errors.WithMessage(err, errMessage).Error())
	}
	return err
}
