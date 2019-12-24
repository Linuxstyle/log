package conf

import (
    "fmt"
    "go.uber.org/zap"
    "gopkg.in/ini.v1"
    "eagle/consume/log"
)

type ConfigInfo struct {
    EtcdInfo   `ini:"etcd"`
    KafkaInfo  `ini:"kafka"`
    ElasticInfo `ini:"elastic"`
    Log *zap.Logger
}

type EtcdInfo struct {
    CAFile  string `ini:"CAFile"`
    KeyFile string `ini:"KeyFile"`
    CertFile string `ini:"CertFile"`
}


type KafkaInfo struct {
    Address string `ini:"address"`
}
type ElasticInfo struct {
    Address string `ini:"address"`
}

func NewConfigInfo()*ConfigInfo{
    return &ConfigInfo{
        Log: log.Log(),
    }
}


func (c *ConfigInfo) Init()(cfg *ini.File, err error){
    //cfg, err = ini.Load("/etc/eagle/conf.ini")
    cfg, err = ini.Load("./conf/conf.ini")
    if err != nil {
        c.Log.Error("load config file  failed"+fmt.Sprintf("%s\n",err))
        return
    }
    return
}
