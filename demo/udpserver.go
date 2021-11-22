package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

var m = make(map[string]time.Time, 0) // 计数器

// 监听心跳的Router
type LHBRouter struct {
	enet.BaseRouter //一定要先基础BaseRouter
}

// 心跳监控
func (this *LHBRouter) Handle(request iface.IRequest) {
	// 先更新
	locker.Lock()
	m[string(request.GetData())] = time.Now()
	locker.Unlock()
}

// 广播心跳的Router
type BHBRouter struct {
	enet.BaseRouter
}

func (this *BHBRouter) Handle(request iface.IRequest) {
	// 访问这个Router的都是API server
	var s string
	locker.RLock()
	for k := range m {
		s += k + " " // A.A.A.A:a B.B.B.B:b
	}
	locker.RUnlock()

	err := request.GetConnection().SendBuffUdpMsg(10,
		[]byte(s), request.GetRemoteAddr())
	if err != nil {
		fmt.Println(err)
	}
}

// broadcast file location
type BFLRouter struct {
	enet.BaseRouter
}

func (this *BFLRouter) Handle(request iface.IRequest) {
	file := string(request.GetData())

	// TODO 将file广播转发出去
	// <- ch
	err := request.GetConnection().SendBuffUdpMsg(20, // BFL
		[]byte(s), request.GetRemoteAddr())
	if err != nil {
		fmt.Println(err)
	}
}

func ListenHeartBeat() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("udp")

	s.AddRouter(10, &LHBRouter{})
	s.AddRouter(11, &BHBRouter{})
	s.AddRouter(20, &BHBRouter{})

	//2 开启服务
	s.Serve()
}

var locker sync.RWMutex

func main() {
	go ListenHeartBeat()

	for {
		locker.Lock()
		for k, t := range m {
			if t.Add(2 * time.Second).Before(time.Now()) {
				delete(m, k)
				log.Printf("<%s>失效\n", k)
			}
		}
		locker.Unlock()
		time.Sleep(2 * time.Second)
	}

}
