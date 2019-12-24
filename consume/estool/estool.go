package estool

import (
    "context"
    "encoding/json"
    "fmt"
    "go.uber.org/zap"
    "math/rand"
    "strconv"
    "time"

    //"gopkg.in/olivere/elastic.v7"
    "github.com/olivere/elastic/v7"
    "eagle/consume/log"
)

var (
    cli *elastic.Client
)

type MsgData struct {
    Message string
}
type Elastic struct {
    Address string
    Log *zap.Logger
}
func NewElastic(address string )*Elastic{
    return &Elastic{
        Address: address,
        Log:  log.Log(),
    }
}
//初始化es连接
func (e *Elastic) Init() (err error) {
    //初始化elastic连接
    //cli, err = elastic.NewClient(elastic.SetURL("http://127.0.0.1:9200"))
    cli, err = elastic.NewClient(elastic.SetURL(e.Address))
    if err != nil {
        e.Log.Error("connect elastic failed: "+fmt.Sprintf("%s\n", err))
        return
    }
    //测试elastic连接
    //info, code, err := cli.Ping("http://127.0.0.1:9200").Do(context.Background())
    info, code, err := cli.Ping(e.Address).Do(context.Background())
    if err != nil {
        e.Log.Error("get elastic info failed: "+fmt.Sprintf("%s\n", err))
        return
    }
    e.Log.Info(fmt.Sprintf("elasticsearch version: %s   response code:%d\n", info.Version.Number, code))
    //e.SendToEs()
    return
}

//把信息写入es
func(e *Elastic) SendToEs(topic string, msg string ) (err error) {
    //tweet2 := `{"user" : "olivere", "message" : "It's a Raggy Waltz"}`
    //生成随机id，用于elastic的id字段
    num := rand.Int63()
    str:= strconv.FormatInt(num,10)

    //json序列化数据，写入es只能是json格式的字符串或者json数据
    data:=&MsgData{Message:msg}
    message,err:=json.Marshal(data)
    if err != nil {
       e.Log.Error(fmt.Sprintf("json Marshal failed: %s ",err))
        return
    }
    //写入es
    indexDate := time.Now().Format("2006-01-02")
    _, err = cli.Index().Index(topic+"-"+indexDate).Id(str).BodyString(string(message)).Do(context.Background())
    if err != nil {
        e.Log.Error(fmt.Sprintf("insert to elasticsearch data failed: %s ",err))
        return
    }
    //log.Printf("Indexed tweet: %s  to index %s, type %s\n",put2.Id,put2.Index,put2.Type)
    return
}


