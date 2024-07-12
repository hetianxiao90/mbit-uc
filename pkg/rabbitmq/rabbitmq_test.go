package rabbitmq

import (
	"fmt"
	"testing"
	"time"
)

var AMQPT = new(AMQPConnectionPool)

func TestNewAMQPConnectionPool(t *testing.T) {
	AMQPT = NewAMQPConnectionPool(&options{
		maxOpen:     1000,
		maxIdle:     500,
		maxAttempts: 5,
		url: fmt.Sprintf("amqp://%s:%s@%s:%d/",
			"guest",
			"guest",
			"127.0.0.1",
			5672,
		),
	})
	//wg := &sync.WaitGroup{}
	//for i := 0; i < 1000; i++ {
	//	go func() {
	//		defer wg.Done()
	//		err := AMQPT.Publish("", "test_queue", []byte("Hello, World!"))
	//		if err != nil {
	//			fmt.Println("Failed to publish message:", err)
	//		} else {
	//			fmt.Println("success to publish message:", err)
	//		}
	//	}()
	//}
	//
	//wg.Wait()
	start := time.Now()
	for i := 0; i < 10000; i++ {
		func() {
			err := AMQPT.Publish("", "test_queue", []byte("Hello, World!"))
			if err != nil {
				fmt.Println("Failed to publish message:", err)
			}
		}()

	}
	end := time.Now()
	duration := end.Sub(start)
	fmt.Println("执行时间：", duration)
	defer AMQPT.Close()

}
