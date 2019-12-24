package  conf

import (
    "gopkg.in/ini.v1"
)
type  ConfInfo struct {
     KafkaInfo  `ini:"kafka"`
     FileInfo `ini:"file"`
     EtcdInfo  `ini:"etcd"`
}

type KafkaInfo struct {
    Address string `ini:"address"`
}

type FileInfo struct {
    FilePath  string `ini:"filePath"`
    Position  string `ini:"position"`
}


type EtcdInfo struct {
    CAFile       string `ini:"CAFile"`
    KeyFile      string `ini:"KeyFile"`
    CertFile     string `ini:"CertFile"`
    Address      string `init:"Address"`
}


func NewConfInfo()*ConfInfo{
    return &ConfInfo{}
}

func(c *ConfInfo) Init() (cfg *ini.File,err error){
    cfg, err = ini.Load("./conf/conf.ini")
    //cfg, err = ini.Load("/etc/eagle/conf.ini")
    if err   != nil {
        return
    }
    return
}
