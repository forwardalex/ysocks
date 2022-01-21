package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/forwardalex/Ytool/log"
	"github.com/forwardalex/Ytool/tool"
	redisDB "github.com/forwardalex/ysocks/dao"
	"github.com/forwardalex/ysocks/tcp"
	"github.com/tal-tech/go-zero/core/conf"
	"os"
	"os/signal"
	"syscall"
	"time"
)


func main() {
	closeChan := make(chan struct{})
	tool.Init()
	log.Init("Dev")
	ctx := context.Background()
	log.Info(ctx,"sever start ","ok ")

	go tcp.ClientMain(ctx,conf.Conf.ServerPortConfg.LocaServerPort, conf.Conf.ServerPortConfg.Host)



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
