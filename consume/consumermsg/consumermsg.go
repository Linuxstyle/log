package consumermsg

import (
    "fmt"
    cluster "github.com/bsm/sarama-cluster"
    "eagle/consume/log"
    "go.uber.org/zap"
)

type KafkaConsumer struct {
    //consumers *cluster.Consumer
    Address   string
    Log   *zap.Logger

}

func NewKafkaConsumer(address string) *KafkaConsumer {
    return &KafkaConsumer{
        Address: address,
        Log: log.Log(),
    }
}
func (k *KafkaConsumer) Init(es func(topic string, msg string)(err error), topic string) (err error) {
    config := cluster.NewConfig()
    config.Consumer.Return.Errors = true
    config.Consumer.Offsets.CommitInterval.Milliseconds()
    config.Group.Return.Notifications = true
    brokes := []string{k.Address}
    topics := []string{topic}
    consumer, err := cluster.NewConsumer(brokes, topic, topics, config)
    if err != nil {
        k.Log.Error("kafka init consumer failed")
        return
    }
    //k.consumers = append(k.consumers, consumer)

    defer consumer.Close()

    go func() {
        for err := range consumer.Errors() {
            k.Log.Error("Error: " + err.Error())
        }
    }()

    go func() {
        for ntf := range consumer.Notifications() {
            k.Log.Info("Rebalanced: "+fmt.Sprintf("%v\n",ntf))
        }
    }()

    for {
        select {
        case msg, ok := <-consumer.Messages():
            if ok {
                //log.Printf( "topic:%s partition:%d offset:%d  Key:%s Value:%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
                consumer.MarkOffset(msg, "")
                err = es(msg.Topic, string(msg.Value))
                if err != nil {
                    k.Log.Error("Write To Elastic Failed: "+fmt.Sprintf("%s\n", err))
                    return
                }
            }
        }
    }

}
