package main

import (
    "eagle/consume/conf"
    "eagle/consume/consumermsg"
    "eagle/consume/estool"
    "eagle/consume/etcdtool"
    "eagle/consume/log"
    "fmt"
    "os"
    "os/signal"
    "sync"
    "syscall"
)

var (
    wg     sync.WaitGroup
)

func main() {
    //初始化config配置文件
    log :=log.Log()
    config := conf.NewConfigInfo()
    cfg, err := config.Init()
    if err != nil {
        log.Error(fmt.Sprintf("init config file failed: %s\n",err))
        return
    }
    err = cfg.MapTo(&config)
    if err != nil {
        log.Error(fmt.Sprintf("load config data failed: %s\n", err))
        return
    }
    log.Info("load config data sucessfully")

    //初始化etcd 连接
    etcdTool := etcdtool.NewEtcdTool(config.EtcdInfo.CAFile, config.EtcdInfo.KeyFile, config.EtcdInfo.CertFile)
    err = etcdTool.Init()
    if err != nil {
        log.Error(fmt.Sprintf("init etcd  client failed: %s\n", err))
        return
    }
    log.Info("init etcd client successfully")

    //初始化elastic
    elastic := estool.NewElastic(config.ElasticInfo.Address)
    //log.Println("Elastic Address: ",config.ElasticInfo.Address)
    err = elastic.Init()
    if err != nil {
        log.Error(fmt.Sprintf("init elasticsearch  failed: %s\n ", err))
        return
    }
    log.Info("init elasticsearch  successfully")


    //ch1 := make(chan string,10)
    //ch2 := make(chan  string,2)
    //起协成，在后端watch、get etcd数据
    go etcdTool.GetEtcd()
    go etcdTool.WatchEtcd()


    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt,syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

    consumerMsg:=consumermsg.NewKafkaConsumer(config.KafkaInfo.Address)

    for {
        select {
        //根据get到etcd的数据进行消费
        case topic1 := <- etcdTool.Gch :
            go func() {
                err = consumerMsg.Init(elastic.SendToEs,topic1)
                if err != nil {
                    log.Error(fmt.Sprintf("init kafka consumer client failed: %s\n ", err))
                    return
                }
            }()
            //根据watch到的etcd数据进行消费
        case topic2 := <- etcdTool.Wch:
            go func() {
                err = consumerMsg.Init(elastic.SendToEs, topic2)
                if err != nil {
                   log.Error(fmt.Sprintf("init kafka consumer client failed: %s\n", err))
                    return
                }
            }()
        case <-signals:
            log.Info("stop successfully")
            return
        }
    }
}
