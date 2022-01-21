package tcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"github.com/forwardalex/ysocks/conf"
	"time"
	"github.com/forwardalex/Ytool/log"
	"net"
	"strings"
)

var ReConnect chan time.Time

func ClientMain(ctx context.Context,localServer []int, remoteIP string) {
	// 开始重连守护进程
	ReConnect = make(chan time.Time)
	go ReConnectToServer(ctx)
	cmd := make([]string, 2)
	remoteControlAddr := remoteIP + ":5005"
	tcpConn, err := CreateTCPConn(remoteControlAddr)
	if err != nil {
		log.Error(ctx, "[连接失败]",remoteControlAddr+err.Error())
		return
	}
	log.Error(ctx, "[已连接]",remoteControlAddr)
	defer tcpConn.Close()
	reader := bufio.NewReader(tcpConn)
	writer := bufio.NewWriter(tcpConn)
	for {
		s, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			fmt.Println(err)
			break
		}
		if s == KeepAlive+"\n" {
			now := time.Now()
			//回复心跳
			writer.WriteString(KeepAlive + "\n")
			ReConnect <- now
		}
		// 当有新连接信号出现时，新建一个tcp连接
		if strings.Contains(s, NewConnection) {
			cmd = strings.Split(s, "@")
			cmd[1] = strings.Replace(cmd[1], "\n", "", -1)
			ip := cmd[1]
			go connectLocalAndRemote(ctx,ip, remoteIP)
		}
	}
	log.Info(ctx, "[已断开]",remoteControlAddr)
}

func connectLocalAndRemote(ctx context.Context,ip string, remoteIP string) {
	remote := connectRemote(ctx,remoteIP)
	local := connectLocal(ctx,ip)
	if local != nil && remote != nil {
		Join2Conn(local, remote, ip)
	} else {
		if local != nil {
			_ = local.Close()
		}
		if remote != nil {
			_ = remote.Close()
		}
	}
	//local := connectLocal(localServer)
}

func connectLocal(ctx context.Context,localServer string) *net.TCPConn {
	fmt.Println(localServer, "local server")
	conn, err := CreateTCPConn(localServer)
	if err != nil {
		log.Error(ctx, "[连接本地服务失败]",err.Error())
	}
	return conn
}

//通信端口
func connectRemote(ctx context.Context,remoteIP string) *net.TCPConn {
	remoteServerAddr := remoteIP + ":5006"
	conn, err := CreateTCPConn(remoteServerAddr)
	if err != nil {
		log.Info(ctx, "[连接远端服务失败]",err.Error())
	}
	return conn
}

func ReConnectToServer(ctx context.Context,) {
	log.Info(ctx, "重连机制开启","")
	for {
		select {
		case <-ReConnect:
		case <-time.After(time.Second * 5):
			ClientMain(ctx,conf.Conf.ServerPortConfg.LocaServerPort, conf.Conf.ServerPortConfg.Host)
			log.Info(ctx , "执行了一次重连","")
			return
		}
	}
}
