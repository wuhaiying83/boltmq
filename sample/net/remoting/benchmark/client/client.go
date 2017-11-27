package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/boltmq/boltmq/net/remoting"
	"github.com/boltmq/common/protocol"
	"github.com/boltmq/common/protocol/header/namesrv"
)

var (
	remotingClient remoting.RemotingClient
)

func main() {
	//debug.SetMaxThreads(100000)
	host := flag.String("h", "10.122.1.200", "host")
	port := flag.Int("p", 10911, "port")
	gonum := flag.Int("n", 100, "thread num")
	sendnum := flag.Int("c", 10000, "thread/per send count")
	sendsize := flag.Int("s", 100, "send data size")
	flag.Parse()

	initClient()
	addr := net.JoinHostPort(*host, strconv.Itoa(*port))
	synctest(addr, *gonum, *sendnum, *sendsize)
}

func newbytes(size int) []byte {
	bs := make([]byte, size)
	for i := 0; i < size; i++ {
		bs[i] = 92
	}

	return bs
}

func synctest(addr string, gonum, sendnum, sendsize int) {
	var (
		wg      sync.WaitGroup
		success int64
		failed  int64
		total   int
	)

	// 请求的custom header
	topicStatisInfoRequestHeader := &namesrv.GetTopicStatisInfoRequestHeader{}
	topicStatisInfoRequestHeader.Topic = "testTopic"
	body := newbytes(sendsize)

	// 同步消息
	total = gonum * sendnum
	wg.Add(gonum)
	start := time.Now()
	for ii := 0; ii < gonum; ii++ {
		go func() {
			for i := 0; i < sendnum; i++ {
				request := protocol.CreateRequestCommand(protocol.GET_TOPIC_STATS_INFO, topicStatisInfoRequestHeader)
				request.Body = body
				response, err := remotingClient.InvokeSync(addr, request, 3000)
				if err != nil {
					failed++
					//log.Printf("Send Mssage[Sync] failed: %s\n", err)
				} else {
					if response.Code == protocol.SUCCESS {
						atomic.AddInt64(&success, 1)
						//log.Printf("Send Mssage[Sync] success. response: body[%s]\n", string(response.Body))
					} else {
						atomic.AddInt64(&failed, 1)
						//log.Printf("Send Mssage[Sync] failed: protocol[%d] err[%s]\n", response.Code, response.Remark)
					}
				}
			}

			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Now()
	spend := end.Sub(start)
	spendTime := int(end.UnixNano() - start.UnixNano())
	tps := total * 1000000000 / spendTime

	log.Printf("Send Mssage[Sync]. Time: %v, Total: %d, Success: %d, Failed: %d, Tps: %d\n", spend, total, success, failed, tps)
}

func initClient() {
	// 初始化客户端
	remotingClient = remoting.NewNMRemotingClient()
	//remotingClient.RegisterContextListener(&ClientContextListener{})
	remotingClient.Start()
}
