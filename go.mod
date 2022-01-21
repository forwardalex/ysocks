module github.com/forwardalex/ysocks

go 1.16

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/spf13/viper v1.8.1
	go.uber.org/zap v1.17.0
	google.golang.org/grpc v1.38.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	github.com/forwardalex/Ytool v0.0.5
)
replace (
	github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6
	google.golang.org/grpc v1.41.0 => google.golang.org/grpc v1.26.0
)
