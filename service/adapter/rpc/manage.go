package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"io/ioutil"
	"net"
	"path/filepath"
	"sync"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/xuperchain/xupercore/kernel/engines"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos"
	edef "github.com/xuperchain/xupercore/kernel/engines/xuperos/commom"
	"github.com/xuperchain/xupercore/lib/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	sconf "github.com/xuperchain/xuperos/common/config"
	def "github.com/xuperchain/xuperos/common/def"
	"github.com/xuperchain/xuperos/common/pb"
)

// rpc server启停控制管理
type RpcServMG struct {
	cfg      *sconf.ServConf
	log       logs.Logger
	engine    edef.Engine
	rpcServ   *Server
	servHD    *grpc.Server
	tlsServHD *grpc.Server
	isInit    bool
	exitOnce  *sync.Once
}

func NewRpcServMG(cfg *sconf.ServConf, engine engines.BCEngine) (*RpcServMG, error) {
	if cfg == nil || engine == nil {
		return nil, fmt.Errorf("param error")
	}
	xosEngine, err := xuperos.EngineConvert(engine)
	if err != nil {
		return nil, fmt.Errorf("not xuperos engine")
	}

	log, _ := logs.NewLogger("", def.SubModName)
	obj := &RpcServMG{
		cfg:      cfg,
		log:      log,
		engine:   xosEngine,
		rpcServ:  NewRpcServ(engine.(edef.Engine), log),
		isInit:   true,
		exitOnce: &sync.Once{},
	}

	return obj, nil
}

// 启动rpc服务
func (t *RpcServMG) Run() error {
	if !t.isInit {
		return errors.New("RpcServMG not init")
	}

	if t.cfg.EnableTls {
		err := t.runTlsServ()
		if err != nil {
			t.log.Error("grpc tls server abnormal exit.err: %v", err)
			return err
		}
	}

	// 启动rpc server，阻塞直到退出
	err := t.runRpcServ()
	if err != nil {
		t.log.Error("grpc server abnormal exit.err: %v", err)
		return err
	}

	t.log.Trace("grpc server exit")
	return nil
}

// 退出rpc服务，释放相关资源，需要幂等
func (t *RpcServMG) Exit() {
	if !t.isInit {
		return
	}

	t.exitOnce.Do(func() {
		t.stopRpcServ()
	})
}

// 启动rpc服务，阻塞直到退出
func (t *RpcServMG) runRpcServ() error {
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		t.rpcServ.UnaryInterceptor(),
		t.rpcServ.UnaryAccess(),
		grpc_prometheus.UnaryServerInterceptor,
	}

	rpcOptions := []grpc.ServerOption{
		middleware.WithUnaryServerChain(unaryInterceptors...),
		grpc.MaxRecvMsgSize(t.cfg.MaxMsgSize),
		grpc.ReadBufferSize(t.cfg.ReadBufSize),
		grpc.InitialWindowSize(t.cfg.InitWindowSize),
		grpc.InitialConnWindowSize(t.cfg.InitConnWindowSize),
		grpc.WriteBufferSize(t.cfg.WriteBufSize),
	}

	t.servHD = grpc.NewServer(rpcOptions...)
	pb.RegisterXchainServer(t.servHD, t.rpcServ)

	lis, err := net.Listen("tcp", fmt.Sprintf("%d", t.cfg.RpcPort))
	if err != nil {
		t.log.Error("failed to listen", "err", err)
		return fmt.Errorf("failed to listen")
	}

	reflection.Register(t.servHD)
	if err := t.servHD.Serve(lis); err != nil {
		t.log.Error("failed to serve", "err", err)
		return err
	}

	t.log.Trace("rpc server exit")
	return nil
}

func (t *RpcServMG) runTlsServ() error {
	t.log.Trace("start tls rpc server")
	tlsPath := t.cfg.TlsPath
	bs, err := ioutil.ReadFile(filepath.Join(tlsPath, "cert.crt"))
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return err
	}

	certificate, err := tls.LoadX509KeyPair(filepath.Join(tlsPath, "key.pem"), filepath.Join(tlsPath, "private.key"))
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   t.cfg.TlsServerName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		t.rpcServ.UnaryInterceptor(),
		t.rpcServ.UnaryAccess(),
		grpc_prometheus.UnaryServerInterceptor,
	}

	rpcOptions := []grpc.ServerOption{
		middleware.WithUnaryServerChain(unaryInterceptors...),
		grpc.MaxRecvMsgSize(t.cfg.MaxMsgSize),
		grpc.ReadBufferSize(t.cfg.ReadBufSize),
		grpc.InitialWindowSize(t.cfg.InitWindowSize),
		grpc.InitialConnWindowSize(t.cfg.InitConnWindowSize),
		grpc.WriteBufferSize(t.cfg.WriteBufSize),
		grpc.Creds(creds),
	}

	l, err := net.Listen("tcp", fmt.Sprintf("%d", t.cfg.TlsRpcPort))
	if err != nil {
		return err
	}

	t.tlsServHD = grpc.NewServer(rpcOptions...)
	pb.RegisterXchainServer(t.tlsServHD, t.rpcServ)
	reflection.Register(t.tlsServHD)

	go func() {
		if err = t.tlsServHD.Serve(l); err != nil {
			t.log.Error("failed to tls serve", "err", err)
			panic(err)
		}
	}()

	t.log.Trace("start tls rpc server")
	return nil
}

// 需要幂等
func (t *RpcServMG) stopRpcServ() {
	if t.servHD != nil {
		// 优雅关闭grpc server
		t.servHD.GracefulStop()
		t.tlsServHD.GracefulStop()
	}
}
