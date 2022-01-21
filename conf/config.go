package conf

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(AppConfig)

type ServerProperties struct {
	Bind           string `cfg:"bind"`
	Port           int    `cfg:"port"`
	AppendOnly     bool   `cfg:"appendOnly"`
	AppendFilename string `cfg:"appendFilename"`
	MaxClients     int    `cfg:"maxclients"`
	RequirePass    string `cfg:"requirepass"`

	Peers []string `cfg:"peers"`
	Self  string   `cfg:"self"`
}

var Properties *ServerProperties

func makeServerConfig(port int) *ServerProperties {
	return &ServerProperties{
		Bind:           "0.0.0.0",
		Port:           port,
		AppendOnly:     false,
		AppendFilename: "",
		MaxClients:     1000,
	}
}

type AppConfig struct {
	Name      string `mapstructure:"name" json:"name"`
	Mode      string `mapstructure:"mode" json:"mode"`
	Version   string `mapstructure:"version" json:"version"`
	StartTime string `mapstructure:"start_time" json:"start_time"`
	MachineID int64  `mapstructure:"machine_id" json:"machine_id"`

	*LogConfig       `mapstructure:"log" json:"Log"`
	*RedisConfig     `mapstructure:"redis" json:"Redis"`
	*ServerPortConfg `mapstructure:"serverport" json:"Serverport"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host" json:"host"`
	Password     string `mapstructure:"password" json:"password"`
	Port         int    `mapstructure:"port" json:"port"`
	DB           int    `mapstructure:"db" json:"db"`
	PoolSize     int    `mapstructure:"pool_size" json:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" json:"min_idle_conns"`
}

type LogConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	Filename   string `mapstructure:"filename" json:"filename"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups"`
}

type ServerPortConfg struct {
	Host           string `mapstructure:"host" json:"host"`
	EscServerPort  []int  `mapstructure:"escServerPort" json:"escServerPort"`
	LocaServerPort []int  `mapstructure:"locaServerPort" json:"locaServerPort"`
}

func Init(fp string) (err error) {
	viper.SetConfigFile(fp)

	//viper.SetConfigName("config")
	//viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("viper.ReadInConfig failed,err:%v\n", err)
		return err
	}
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarsh failed,err:%v\n", err)

	}
	go func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			fmt.Println("conf was update")
			if err := viper.Unmarshal(Conf); err != nil {
				fmt.Printf("viper.Unmarshal failed,err:%v\n", err)
			}
		})
	}()
	return
}
