package mux

import (
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/onsi/gomega"
	"testing"
)

func getMockConsumerFactory(t *testing.T) ConsumerFactory {
	return func(topics []string, name string) (*client.Consumer, error) {
		return client.GetConsumerMock(t), nil
	}
}

func getMultiplexerMock(t *testing.T) *Multiplexer {
	asyncP, _ := client.GetAsyncProducerMock(t)
	syncP, _ := client.GetSyncProducerMock(t)
	return NewMultiplexer(getMockConsumerFactory(t), syncP, asyncP, "name", logroot.Logger())
}

func TestMultiplexer(t *testing.T) {
	gomega.RegisterTestingT(t)
	mux := getMultiplexerMock(t)
	gomega.Expect(mux).NotTo(gomega.BeNil())

	c1 := mux.NewConnection("c1")
	gomega.Expect(c1).NotTo(gomega.BeNil())
	c2 := mux.NewConnection("c2")
	gomega.Expect(c2).NotTo(gomega.BeNil())

	ch1 := make(chan *client.ConsumerMessage)
	ch2 := make(chan *client.ConsumerMessage)

	err := c1.ConsumeTopic(ch1, "topic1")
	gomega.Expect(err).To(gomega.BeNil())
	err = c2.ConsumeTopic(ch2, "topic2", "topic3")
	gomega.Expect(err).To(gomega.BeNil())

	mux.Start()
	gomega.Expect(mux.started).To(gomega.BeTrue())

	// once the multiplexer is start an attempt to subscribe returns an error
	err = c1.ConsumeTopic(ch1, "anotherTopic1")
	gomega.Expect(err).NotTo(gomega.BeNil())

	mux.Close()
	close(ch1)
	close(ch2)

}
