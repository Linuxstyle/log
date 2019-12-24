package kafkatool

import (
    "fmt"
    "github.com/Shopify/sarama"
    "go.uber.org/zap"
    "eagle/product/log"
)

var (
    producer sarama.SyncProducer
)
type KafkaInfo struct {
    Address string
    Log *zap.Logger
}

func NewKafkaInfo(address string )*KafkaInfo{
    return &KafkaInfo{
        Address: address,
        Log: log.Log(),
    }
}
//初始化kafka product
func(k *KafkaInfo) Init() (err error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Partitioner = sarama.NewRandomPartitioner
    config.Producer.Return.Successes = true
    producer, err = sarama.NewSyncProducer([]string{k.Address}, config)
    if err != nil {
        return
    }
    k.Log.Info(fmt.Sprintf("init connect to kafka successfully"))
    return
}

//调用product 发送mesage 到kafka
func(k *KafkaInfo) SendToMsg(message string, topic string) (err error) {
    //defer producer.Close()
    //log.Printf("topic: %s, msg:%s", topic, message)
    msg := &sarama.ProducerMessage{}
    msg.Topic = topic
    msg.Value = sarama.StringEncoder(message)
    _, _, err = producer.SendMessage(msg)
    if err != nil {
        return err
    }
    //log.Printf("send message to kafka, partition:%d, offset:%d \n", partition, offset)
    return
}
