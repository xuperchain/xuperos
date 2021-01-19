package rpc

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"

	edef "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	sctx "github.com/xuperchain/xuperos/common/context"
	"github.com/xuperchain/xuperos/common/pb"
)

type Server struct {
	engine edef.Engine
	log    logs.Logger
}

func NewRpcServ(engine edef.Engine, log logs.Logger) *Server {
	return &Server{
		engine: engine,
		log:    log,
	}
}

// set request header
type HeaderInterface interface {
	GetHeader() *pb.Header
}

// UnaryInterceptor provides a hook to intercept the execution of a unary RPC on the server.
func (s *Server) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		// panic recover
		defer func() {
			if e := recover(); e != nil {
				s.log.Error("Rpc server happen panic", "error", e, "rpc_method", info.FullMethod)
			}
		}()

		if req.(HeaderInterface).GetHeader() == nil {
			header := reflect.ValueOf(req).Elem().FieldByName("Header")
			if header.IsValid() && header.IsNil() && header.CanSet() {
				header.Set(reflect.ValueOf(defReqHeader()))
			}
		}
		if req.(HeaderInterface).GetHeader().GetLogid() == "" {
			req.(HeaderInterface).GetHeader().Logid = utils.GenLogId()
		}

		// handle request
		return handler(ctx, req)
	}
}

func (s *Server) UnaryAccess() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 获取客户端ip
		clientIp, err := s.getClientIP(ctx)
		if err != nil {
			s.log.Error("access proc failed because get client ip failed", "error", err)
			return nil, fmt.Errorf("get client ip failed")
		}

		obj, ok := req.(HeaderInterface)
		if !ok {
			s.log.Error("access proc failed because req no header")
			return nil, fmt.Errorf("req no header")
		}
		reqHeader := obj.GetHeader()

		// 创建请求上下文
		reqCtx, err := sctx.NewReqCtx(s.engine, reqHeader.GetLogid(), clientIp)
		if err != nil {
			s.log.Error("access proc failed because create request context failed", "error", err)
			return nil, fmt.Errorf("create request context failed")
		}

		ctx = sctx.ContextWithReqCtx(ctx, reqCtx)
		resp, err = handler(ctx, req)
		obj, ok = resp.(HeaderInterface)
		if !ok {
			s.log.Error("access proc failed because resp no header")
			return nil, fmt.Errorf("resp no header")
		}

		respHeader := obj.GetHeader()
		logFields := make([]interface{}, 0)
		logFields = append(logFields, "method", info.FullMethod, "from", reqHeader.GetFromNode(), "client_ip", clientIp)
		logFields = append(logFields, "cost_time", reqCtx.GetTimer().Print(), "stable", respHeader.GetError(), "error", err)
		reqCtx.GetLog().Info("access", logFields...)
		return resp, err
	}
}

func defReqHeader() *pb.Header {
	return &pb.Header{
		Logid:    utils.GenLogId(),
		FromNode: "unknown",
	}
}

func defRespHeader(header *pb.Header) *pb.Header {
	return &pb.Header{
		Logid:   header.GetLogid(),
		Error:   pb.XChainErrorEnum_SUCCESS,
		FromNode: utils.GetHostName(),
	}
}

// 请求处理前处理，考虑到各接口个性化记录日志，没有使用拦截器
// others必须是KV格式，K为string
func (s *Server) access(ctx context.Context, header *pb.Header,
	others ...interface{}) (sctx.ReqCtx, error) {
	// 获取客户端ip
	clientIp, err := s.getClientIP(ctx)
	if err != nil {
		s.log.Error("access proc failed because get client ip failed", "error", err)
		return nil, fmt.Errorf("get client ip failed")
	}

	// 创建请求上下文
	rctx, err := sctx.NewReqCtx(s.engine, header.GetLogid(), clientIp)
	if err != nil {
		s.log.Error("access proc failed because create request context failed", "error", err)
		return nil, fmt.Errorf("create request context failed")
	}

	// 输出access log
	logFields := make([]interface{}, 0)
	logFields = append(logFields, "from", header.GetFromNode(), "client_ip", clientIp)
	logFields = append(logFields, others...)
	rctx.GetLog().Trace("received request", logFields...)

	return rctx, nil
}

// 请求完成后处理
// others必须是KV格式，K为string
func (s *Server) ending(rctx sctx.ReqCtx, header *pb.Header, others ...interface{}) {
	// 输出ending log
	logFields := make([]interface{}, 0)
	logFields = append(logFields, "error", header.GetError(), "cost_time", rctx.GetTimer().Print())
	logFields = append(logFields, others...)
	rctx.GetLog().Info("request done", logFields...)
}

func (s *Server) getClientIP(gctx context.Context) (string, error) {
	pr, ok := peer.FromContext(gctx)
	if !ok {
		return "", fmt.Errorf("create peer form context failed")
	}

	if pr.Addr == nil || pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("get client_ip failed because peer.Addr is nil")
	}

	addrSlice := strings.Split(pr.Addr.String(), ":")
	return addrSlice[0], nil
}
