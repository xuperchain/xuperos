package service

import (
	"fmt"

	"github.com/xuperchain/xupercore/kernel/engines"
	"github.com/xuperchain/xupercore/lib/logs"

	sconf "github.com/xuperchain/xuperos/common/config"
	def "github.com/xuperchain/xuperos/common/def"
	adpgw "github.com/xuperchain/xuperos/service/adapter/gateway"
	adprpc "github.com/xuperchain/xuperos/service/adapter/rpc"
	endrpc "github.com/xuperchain/xuperos/service/endorser"
	"github.com/xuperchain/xuperos/service/rpc"
)

// 由于需要同时启动多个服务组件，采用注册机制管理
type ServCom interface {
	Run() error
	Exit()
}

// 各server组件运行控制
type ServMG struct {
	scfg    *sconf.ServConf
	log     logs.Logger
	servers []ServCom
}

func NewServMG(scfg *sconf.ServConf, engine engines.BCEngine) (*ServMG, error) {
	if scfg == nil || engine == nil {
		return nil, fmt.Errorf("param error")
	}

	log, _ := logs.NewLogger("", def.SubModName)
	obj := &ServMG{
		scfg:    scfg,
		log:     log,
		servers: make([]ServCom, 0),
	}

	// 实例化rpc服务
	rpcServ, err := rpc.NewRpcServMG(scfg, engine)
	if err != nil {
		return nil, err
	}
	obj.servers = append(obj.servers, rpcServ)

	// 实例化老版本接口适配服务
	if scfg.EnableAdapter {
		adpServ, err := adprpc.NewRpcServMG(scfg, engine)
		if err != nil {
			return nil, err
		}
		adpGW, err := adpgw.NewGateway(scfg)
		if err != nil {
			return nil, err
		}

		obj.servers = append(obj.servers, adpServ, adpGW)
	}

	//实例化背书服务
	if scfg.EnableEndorser {
		endorserServ, err := endrpc.NewRpcServMG(scfg, engine)
		if err != nil {
			return nil, err
		}

		obj.servers = append(obj.servers, endorserServ)
	}

	return obj, nil
}

// 启动rpc服务
func (t *ServMG) Run() error {
	ch := make(chan error, 0)
	defer close(ch)

	for _, serv := range t.servers {
		// 启动各个service
		go func(s ServCom) {
			ch <- s.Run()
		}(serv)
	}

	// 监听各个service状态
	exitCnt := 0
	for {
		if exitCnt >= len(t.servers) {
			break
		}

		select {
		case err := <-ch:
			t.log.Warn("service exit", "err", err)
			exitCnt++
		}
	}

	return nil
}

// 退出rpc服务，释放相关资源，需要幂等
func (t *ServMG) Exit() {
	for _, serv := range t.servers {
		// 触发各service退出
		go func(s ServCom) {
			s.Exit()
		}(serv)
	}
}
