package main

import (
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgclient/consumer"
	"git.oschina.net/cloudzone/smartgo/stgclient/consumer/listener"
	"git.oschina.net/cloudzone/smartgo/stgclient/process"
	"git.oschina.net/cloudzone/smartgo/stgcommon/message"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/heartbeat"
	"time"
	"sync/atomic"
)

type MessageListenerImpl struct {
	MsgCount int64
	StartTime int64
}

func (listenerImpl *MessageListenerImpl) ConsumeMessage(msgs []*message.MessageExt, context *consumer.ConsumeConcurrentlyContext) listener.ConsumeConcurrentlyStatus {
	for _, msg := range msgs {
		atomic.AddInt64(&listenerImpl.MsgCount, 1)
		if listenerImpl.MsgCount==500000{
			fmt.Println(500000/(time.Now().Unix()-listenerImpl.StartTime),"______________________________________")

		}
		fmt.Println(listenerImpl.MsgCount)
		fmt.Println(msg.ToString())
	}
	return listener.CONSUME_SUCCESS
}

func taskC() {
	t := time.NewTicker(time.Second * 1000)
	for {
		select {
		case <-t.C:
		}

	}
}

func main() {
	defaultMQPushConsumer := process.NewDefaultMQPushConsumer("consume9s991a2a11")
	defaultMQPushConsumer.SetConsumeFromWhere(heartbeat.CONSUME_FROM_LAST_OFFSET)
	defaultMQPushConsumer.SetMessageModel(heartbeat.CLUSTERING)
	defaultMQPushConsumer.SetNamesrvAddr("10.112.68.189:9876")
	defaultMQPushConsumer.Subscribe("cloudzone1", "tagA")
	defaultMQPushConsumer.RegisterMessageListener(&MessageListenerImpl{StartTime:time.Now().Unix()})
	defaultMQPushConsumer.Start()
	go taskC()
	select {}
}
