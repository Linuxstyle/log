package etcdtool

import (
    "context"
    "fmt"
    "github.com/coreos/etcd/clientv3"
    "github.com/coreos/etcd/pkg/transport"
    "go.uber.org/zap"
    "eagle/product/log"
    "strings"
    "time"
)

var (
    cli *clientv3.Client
)
type EtcdInfo struct {
    CAFile string
    KeyFile string
    CertFile string
    Address string
    Log *zap.Logger
}

func NewEtcdInfo(address string, ca string, key string, crt string) *EtcdInfo{
    return &EtcdInfo{
        CAFile:  ca,
        KeyFile:  key,
        CertFile: crt,
        Address:  address,
        Log: log.Log(),
    }
}
//初始化etcd 需要到基本信息，如证书，地址等
func(e *EtcdInfo) Init() (err error) {
    //添加证书信息
    tlsInfo := transport.TLSInfo{
        CAFile:   e.CAFile,
        KeyFile:  e.KeyFile,
        CertFile: e.CertFile,
    }
    //初始化证书
    tlsConfig, err := tlsInfo.ClientConfig()
    if err != nil {
        e.Log.Error(fmt.Sprintf("init etcd tls failed:  %s\n", err))
        return
    }
    config := clientv3.Config{
        Endpoints:            []string{e.Address},
        AutoSyncInterval:     0,
        DialTimeout:          time.Second * 5,
        DialKeepAliveTime:    0,
        DialKeepAliveTimeout: 0,
        MaxCallSendMsgSize:   0,
        MaxCallRecvMsgSize:   0,
        TLS:                  tlsConfig,
        Username:             "",
        Password:             "",
        RejectOldCluster:     false,
        DialOptions:          nil,
        LogConfig:            nil,
        Context:              nil,
        PermitWithoutStream:  false,
    }

    cli, err = clientv3.New(config)
    if err != nil {
        //log.Println("init etcd client failed: ", err)
        return
    }
    return
}

//提交数据
func(e *EtcdInfo)PutData(topic string) (err error) {
    //defer cli.Close()
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    _, err = cli.Put(ctx, "/eagle/topic/"+topic, topic)
    cancel()
    if err != nil {
        e.Log.Error(fmt.Sprintf("put data to etcd failed: %s\n ", err))
        return
    }
    return
}

//WatchData etcd数据
func(e *EtcdInfo) WatchData() {
    //defer cli.Close()
    rch := cli.Watch(context.Background(), "/eagle/topic/", nil)
    for wresp := range rch {
        for _, ev := range wresp.Events {
            e.Log.Error(fmt.Sprintf("Type:%s Key:%q Value:%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value))
        }
    }
}

//获取etcd 数据
func (e *EtcdInfo) GetData(fileName string) (err error) {
    //defer cli.Close()
    path := strings.Split(fileName, "/")
    topic := path[3] + "_" + path[4] + "_" + path[5]
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
    resp, err := cli.Get(ctx, "/eagle/topic/"+topic)
    cancel()
    if err != nil {
        e.Log.Error(fmt.Sprintf("get etcd value failed: %s\n", err))
        return
    }
    //判断topic是否已经存在，如果不存在则提交到etcd
    if len(resp.Kvs) == 1 {
        return
    } else {
        //log.Println("begin put etcd data")
        e.PutData(topic)
    }
    return
}
