package tcp

import (
	"bufio"
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"github.com/forwardalex/ysocks/conf"
	"github.com/forwardalex/ysocks/handler/tcp"
	"github.com/forwardalex/Ytool/log"
	"strconv"
	"sync"
	"time"
)

const (
	controlAddr = "0.0.0.0:5005"
	tunnelAddr  = "0.0.0.0:5006"
)

var (
	clientConn         *net.TCPConn //连接 "0.0.0.0:5005" 发送心跳及服务内容专用通道
	connectionPool     map[string]*ConnMatch
	connectionPoolLock sync.Mutex
	closeChan          chan struct{}
)

type ConnMatch struct {
	addTime time.Time
	accept  *net.TCPConn
	SrcPort int //内网服务所在port
}

func ServerMain(ctx context.Context,userConnIP []int) {
	connectionPool = make(map[string]*ConnMatch, 32)
	f1 := createControlChannel
	f2 := acceptClientRequest
	go f1(ctx)
	go f2()
	for key, v := range userConnIP {
		ip := "0.0.0.0:" + strconv.Itoa(v)
		srcPort := conf.Conf.ServerPortConfg.LocaServerPort[key]
		log.Info(nil, "ip", zap.String("connect src ip ", ip))
		go acceptUserRequest(ip, srcPort)
	}
	cleanConnectionPool()
}

// 创建一个控制通道，用于传递控制消息，如：心跳，创建新连接
func createControlChannel(ctx context.Context) {
	tcpListener, err := CreateTCPListener(controlAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("[已监听]" + controlAddr)
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		log.Info(ctx, "[新连接]"+tcpConn.RemoteAddr().String()+"->"+tcpConn.LocalAddr().String(),"")
		// 如果当前已经有一个客户端存在，则丢弃这个链接
		if clientConn != nil {
			log.Info(nil, "[旧新连接丢弃]"+clientConn.RemoteAddr().String()+"->"+clientConn.LocalAddr().String(),"")
			_ = tcpConn.Close()
		} else {
			clientConn = tcpConn
			go keepAlive()
			//go listAlice()
			//go closeConnect()
		}
	}
}

// 和客户端保持一个心跳链接
func keepAlive() {
	go func() {
		for {
			if clientConn == nil {
				return
			}
			_, err := clientConn.Write(([]byte)(KeepAlive + "\n"))
			if err != nil {
				log.Info(nil, "[已断开客户端连接]", zap.String("ip", clientConn.RemoteAddr().String()))
				_ = clientConn.Close()
				clientConn = nil
				return
			}
			time.Sleep(time.Second * 3)
		}
	}()
}

//监听心跳回复
func listAlice() {
	reader := bufio.NewReader(clientConn)
	for {
		s, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			log.Info(nil, "[已断开客户端连接]", zap.String("ip", clientConn.RemoteAddr().String()))
			_ = clientConn.Close()
			break
		}
		fmt.Println(s)
		if s == KeepAlive+"\n" {
			<-closeChan
		}
	}
}

//心跳超时关闭
func closeConnect() {
	for {
		select {
		case <-closeChan:
			fmt.Println("server heat")
		case <-time.After(time.Second * 5):
			_ = clientConn.Close()
			log.Info(nil, "丢弃心跳超时连接","")
			return
		}
	}
}

// 监听来自用户的请求
func acceptUserRequest(userConn string, srcPort int) {
	tcpListener, err := CreateTCPListener(userConn)
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		addConn2Pool(tcpConn, srcPort)
		localServerPort := "0.0.0.0:" + strconv.Itoa(srcPort)
		sendMessage(NewConnection + "@" + localServerPort + "\n")

	}
}

// 将用户来的连接放入连接池中
func addConn2Pool(accept *net.TCPConn, srcPort int) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()

	now := time.Now()
	connectionPool[strconv.FormatInt(now.UnixNano(), 10)] = &ConnMatch{now, accept, srcPort}
}

// 发送给客户端新消息
func sendMessage(message string) {
	if clientConn == nil {
		log.Error(nil, "[无已连接的客户端]","")
		return
	}
	_, err := clientConn.Write([]byte(message))
	if err != nil {
		log.Error(nil, "[发送消息异常]: message: ", zap.String("", message))
	}
}

// 接收客户端来的请求并建立隧道
func acceptClientRequest() {
	tcpListener, err := CreateTCPListener(tunnelAddr)
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		go establishTunnel(tcpConn)
	}
}

func establishTunnel(tunnel *net.TCPConn) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()
	for key, connMatch := range connectionPool {
		if connMatch.accept != nil {
			go Join2Conn(connMatch.accept, tunnel, connMatch.accept.LocalAddr().String())
			delete(connectionPool, key)
			return
		}
	}
	_ = tunnel.Close()

}

func cleanConnectionPool() {
	for {
		connectionPoolLock.Lock()
		for key, connMatch := range connectionPool {
			if time.Now().Sub(connMatch.addTime) > time.Second*10 {
				_ = connMatch.accept.Close()
				delete(connectionPool, key)
			}
		}
		connectionPoolLock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

// ListenAndServeWithSignal binds port and handle requests, blocking until receive stop signal
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler, closeChan chan struct{}) error {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	//cfg.Address = listener.Addr().String()
	log.Info(nil, fmt.Sprintf("bind: %s, start listening...", cfg.Address),"")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe binds port and handle requests, blocking until close
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// listen signal
	go func() {
		<-closeChan
		log.Info(nil, "shutting down...","")
		_ = listener.Close() // listener.Accept() will return err immediately
		_ = handler.Close()  // close connections
	}()
	// listen port
	defer func() {
		// close during unexpected error
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("list accept err", err)
			break
		}
		// handle
		log.Info(nil, "accept link","")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
