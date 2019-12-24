package etcdtool

import (
    "context"
    "github.com/coreos/etcd/clientv3"
    "github.com/coreos/etcd/pkg/transport"
    "log"
    "sync"
    "time"
)

var (
    cli     *clientv3.Client
    wg sync.WaitGroup
)
type EtcdTool struct {
    CA  string
    ClietKey string
    ClientCRT string
    //Watch Etcd传输数据
    Wch chan string
    //Get Etcd传输数据
    Gch chan string
}

func NewEtcdTool(ca string, clienkey string, clientcrt string)*EtcdTool{
    return &EtcdTool{
        CA: ca,
        ClietKey: clienkey,
        ClientCRT: clientcrt,
        Wch: make(chan string,10),
        Gch: make(chan string,2),
    }
}
func (e *EtcdTool) Init() (err error) {
    tlsconf := transport.TLSInfo{
        CertFile:            e.ClientCRT,
        KeyFile:             e.ClietKey,
        CAFile:              e.CA,
        TrustedCAFile:       "",
        ClientCertAuth:      false,
        CRLFile:             "",
        InsecureSkipVerify:  false,
        SkipClientSANVerify: false,
        ServerName:          "",
        HandshakeFailure:    nil,
        CipherSuites:        nil,
        AllowedCN:           "",
    }
    tlsconfig, err := tlsconf.ClientConfig()
    if err != nil {
        log.Println("load cert faild:  ", err)
        return
    }
    config := clientv3.Config{
        Endpoints:            []string{"172.17.1.119:2379"},
        AutoSyncInterval:     0,
        DialTimeout:          time.Second * 5,
        DialKeepAliveTime:    0,
        DialKeepAliveTimeout: 0,
        MaxCallSendMsgSize:   0,
        MaxCallRecvMsgSize:   0,
        TLS:                  tlsconfig,
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
        log.Println("init etcd client failed:", err)
        return
    }
    return
}

//watch 数据变化
func (e *EtcdTool)WatchEtcd(){
   //defer cli.Close()
   // defer  wg.Done()
    wrch := cli.Watch(context.Background(), "/eagle/topic/", clientv3.WithPrefix())
    for wresp := range wrch {
        for _, resp := range wresp.Events {
            e.Wch <- string(resp.Kv.Value)
        }
    }
}

//获取topic
func(e *EtcdTool) GetEtcd(){
    //defer cli.Close()
    //defer wg.Done()
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
    resp, err := cli.Get(ctx, "/eagle/topic/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
    defer cancel()
    if err != nil {
        log.Println("get from etcd data filed: ", err)
        return
    }
    for _, ev := range resp.Kvs {
         e.Gch<- string(ev.Value)
    }
}
