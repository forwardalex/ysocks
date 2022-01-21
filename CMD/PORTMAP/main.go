package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/forwardalex/Ytool/log"
	"github.com/forwardalex/Ytool/tool"
	"github.com/forwardalex/ysocks/conf"
	redisDB "github.com/forwardalex/ysocks/dao"
	"github.com/forwardalex/ysocks/tcp"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var defaultProperties = &conf.ServerProperties{
	Bind:           "0.0.0.0",
	Port:           6399,
	AppendOnly:     false,
	AppendFilename: "",
	MaxClients:     1000,
}

func main() {
	var filepath string
	closeChan := make(chan struct{})
	flag.StringVar(&filepath, "filepath", "./conf/config.yaml", "configfilepath")
	flag.Parse()
	err := conf.Init(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}
	tool.Init()
	log.Init("Dev")
	ctx := context.Background()
	log.Info(ctx,"sever start ","ok ")
	if err := redisDB.Init(conf.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed,err:=%v\n", err)
		return
	}
	if conf.Conf.MachineID == 1 {
		go tcp.ClientMain(ctx,conf.Conf.ServerPortConfg.LocaServerPort, conf.Conf.ServerPortConfg.Host)
	}
	if conf.Conf.MachineID == 2 {
		go tcp.ServerMain(ctx,conf.Conf.ServerPortConfg.EscServerPort)
		go tcp.ListenAndServeWithSignal(&tcp.Config{
			Address: fmt.Sprintf("%s:%d", conf.Properties.Bind, conf.Properties.Port),
		}, tcp.MakeEchoHandler(), closeChan)
	}
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT) // 此处不会阻塞
	<-quit                                                                                // 阻塞在此，当接收到上述两种信号时才会往下执行
	closeChan <- struct{}{}
	log.Info(ctx,"Shutdown Server ...","")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	log.Info(ctx,"Server exiting","")
}
