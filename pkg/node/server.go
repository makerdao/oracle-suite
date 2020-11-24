package node

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/makerdao/gofer/internal/transport"
)

type Server struct {
	api      *Api
	rpc      *rpc.Server
	listener net.Listener
	network  string
	address  string
}

func NewServer(datastore Datastore, transport transport.Transport, network, address string) (*Server, error) {
	server := &Server{
		api: &Api{
			datastore: datastore,
			transport: transport,
		},
		rpc:     rpc.NewServer(),
		network: network,
		address: address,
	}
	err := server.rpc.Register(server.api)
	if err != nil {
		return nil, err
	}
	server.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	return server, nil
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	err = s.api.datastore.Start()
	if err != nil {
		return err
	}

	go http.Serve(s.listener, nil)
	return nil
}

func (s *Server) Stop() error {
	err := s.api.datastore.Start()
	if err != nil {
		return err
	}
	return s.listener.Close()
}
