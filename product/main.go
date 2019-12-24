package main

import (
    "eagle/product/conf"
    "eagle/product/etcdtool"
    "eagle/product/kafkatool"
    "eagle/product/tailftool"
    "eagle/product/watchfiletool"
    "eagle/product/log"
    "fmt"
    "os"
    "os/signal"
    "syscall"
)



func main() {
    //1.初始化config
    log := log.Log()
    configInfo := conf.NewConfInfo()
    cfg, err := configInfo.Init()
    if err != nil {
        log.Error(fmt.Sprintf("init config file failed: %s\n", err))
        return
    }
    log.Info("init config  success")
    cfg.MapTo(&configInfo)

    //2.初始化kafka
    kafkaInfo := kafkatool.NewKafkaInfo(configInfo.KafkaInfo.Address)
    err = kafkaInfo.Init()
    if err != nil {
        log.Error(fmt.Sprintf("init kafka connect failed: %s\n", err))
        return
    }
    log.Info("kafka init success")

    //3.初始化 fsnotify
    wathFile := watchfiletool.NewWatchFile()
    go func() {
        err = wathFile.Init(configInfo.FileInfo.FilePath)
        if err != nil {
            log.Error(fmt.Sprintf("init watch file failed: %s\n", err))
            return
        }
        log.Info("init  watch file successfully")
    }()

    //4.初始化etcd连接
    etcdInfo := etcdtool.NewEtcdInfo(configInfo.EtcdInfo.Address, configInfo.EtcdInfo.CAFile, configInfo.EtcdInfo.KeyFile, configInfo.EtcdInfo.CertFile)
    err = etcdInfo.Init()
    if err != nil {
        log.Error(fmt.Sprintf("init etcd  failed: %s\n ", err))
        return
    }
    log.Info("init etcd successfully")

    //5.捕捉系统信号推出
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt,syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

    //6.初始化tailf，打开文件
    tailFile := tailftool.NewFileTail(configInfo.FileInfo.Position)
    for {
        select {
        case fileName, ok := <-wathFile.Wach:
            if !ok {
                log.Error("get watchfile from chan failed")
                return
            }
            //6.初始化tail,读取文件内容发送到kafka
            //log.Println("###########",fileName)
            tailFile.Init(fileName)

            //7. 调用GetData方法获取指定的topic，如果不存在则put到etcd
            err := etcdInfo.GetData(fileName)
            if err != nil {
                log.Error(fmt.Sprintf("init etcd  failed: %s\n ", err))
                return
            }
            tailFile.ReadFileLine(kafkaInfo.SendToMsg)
        case <-signals:
            tailFile.Stop()
            log.Info("stop sucessfully")
            return
        }
    }
}
