package server

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"

	"github.com/akka/gms/common"
	"github.com/akka/gms/gmsContext"
	"github.com/akka/gms/protocol"
)

type server struct {
	// 整个服务级别的锁
	sync.RWMutex
	// 路由Map
	routerMap map[string]gmsContext.Controller
	// gms 服务
	gmsHandler *gmsHandler
}

/**
初始化GMS服务
*/
func NewServer() IServer {
	s := server{
		routerMap: make(map[string]gmsContext.Controller),
	}
	return &s
}

/**
准备启动服务的资源
*/
func (s *server) InitServe() {
	fmt.Println("[gmsServer] InitServe")

	pool := goroutine.Default()
	defer pool.Release()

	// codec := &protocol.MessagePack{}
	// 启动gnet
	s.gmsHandler = &gmsHandler{
		gmsServer: s,
		pool:      pool,
		// codec:     codec,
	}
	log.Fatal(gnet.Serve(
		s.gmsHandler,
		fmt.Sprintf("tcp://:%v", common.GMS_PORT),
		gnet.WithMulticore(true),
		gnet.WithTCPKeepAlive(time.Minute*5), // todo 需要确定是否对长连接有影响
		// gnet.WithCodec(codec)
	))
}

/**
启动服务
*/
func (s *server) Run() {
	fmt.Println("[gmsServer] start run gms gmsServer")
	// 准备启动服务的资源
	s.InitServe()

}

/**
停止服务 回收资源
*/
func (s *server) Stop() {
	fmt.Println("[gmsServer] stop")
}

/**
添加路由
*/
func (s *server) AddRouter(handlerName string, handlerFunc gmsContext.Controller) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.routerMap[handlerName]; ok {
		fmt.Println("[AddRouter] fail handlerName:", handlerName, " alread exist")
		return
	}
	s.routerMap[handlerName] = handlerFunc
}

/**
获取路由
*/
func (s *server) GetRouter(handlerName string) (gmsContext.Controller, error) {
	s.RLock()
	defer s.RUnlock()
	if controller, ok := s.routerMap[handlerName]; ok {
		return controller, nil
	}
	return nil, errors.New("[GetRouter] Router not found")
}

/**
处理方法
*/
func (s *server) HandlerMessage(message protocol.Imessage) {
	fmt.Println(string(message.GetExt()))
	controller, err := s.GetRouter(string(message.GetExt()))
	if err != nil {
		// 回写错误信息
		fmt.Println("[HandlerMessage] Router:", message.GetExt(), " not found", err)
		return
	}
	context := gmsContext.NewContext()
	context.Param(message.GetData())
	// 调用方法
	controller(context)
}
