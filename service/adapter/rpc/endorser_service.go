package rpc

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"

	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	sconf "github.com/xuperchain/xuperos/common/config"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
)

type endorserService struct {
	engine      ecom.Engine
	clientCache sync.Map
	mutex       sync.Mutex
	conf        *sconf.ServConf
}

func newEndorserService(cfg *sconf.ServConf, engine ecom.Engine) *endorserService {
	return &endorserService{
		engine: engine,
		conf:   cfg,
	}
}

func (t *endorserService) EndorserCall(gctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error) {
	resp := &pb.EndorserResponse{}
	rctx := sctx.ValueReqCtx(gctx)
	endc, err := t.getClient(t.getHost())
	if err != nil {
		return resp, err
	}
	res, err := endc.EndorserCall(gctx, req)
	if err != nil {
		return resp, err
	}
	resp.EndorserAddress = res.EndorserAddress
	resp.ResponseName = res.ResponseName
	resp.ResponseData = res.ResponseData
	resp.EndorserSign = res.EndorserSign
	rctx.GetLog().SetInfoField("bc_name", req.GetBcName())
	rctx.GetLog().SetInfoField("request_name", req.GetBcName())
	return resp, nil
}

func (t *endorserService) getHost() string {
	host := ""
	hostCnt := len(t.conf.EndorserHosts)
	if hostCnt > 0 {
		rand.Seed(time.Now().Unix())
		index := rand.Intn(hostCnt)
		host = t.conf.EndorserHosts[index]
	}
	return host
}

func (t *endorserService) getClient(host string) (pb.XendorserClient, error) {
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	if c, ok := t.clientCache.Load(host); ok {
		return c.(pb.XendorserClient), nil
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	if c, ok := t.clientCache.Load(host); ok {
		return c.(pb.XendorserClient), nil
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := pb.NewXendorserClient(conn)
	t.clientCache.Store(host, c)
	return c, nil
}