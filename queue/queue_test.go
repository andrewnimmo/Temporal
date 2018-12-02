package queue_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/queue"
)

const (
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	testRabbitAddress = "amqp://127.0.0.1:5672"
	logFilePath       = "/tmp/%s_%s.log"
)

func TestInitialize(t *testing.T) {
	type args struct {
		queueName string
		publish   bool
		service   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"DFAQ", args{queue.DatabaseFileAddQueue, false, false}},
		{"IPQ", args{queue.IpfsPinQueue, false, false}},
		{"IFQ", args{queue.IpfsFileQueue, false, false}},
		{"ESQ", args{queue.EmailSendQueue, false, false}},
		{"IEQ", args{queue.IpnsEntryQueue, false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := queue.Initialize(tt.args.queueName,
				testRabbitAddress, tt.args.publish, tt.args.service, fmt.Sprintf(logFilePath, tt.name, tt.args.publish)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQueues(t *testing.T) {
	qm, err := queue.Initialize(queue.IpfsPinQueue, testRabbitAddress, true, false)
	if err != nil {
		t.Fatal(err)
	}

	pin := queue.IPFSPin{
		CID:              testCID,
		NetworkName:      "public",
		HoldTimeInMonths: 10,
	}

	if err = qm.PublishMessageWithExchange(pin, queue.PinExchange); err != nil {
		t.Fatal(err)
	}
	qm, err = queue.Initialize(queue.EmailSendQueue, testRabbitAddress, true, false)
	if err != nil {
		t.Fatal(err)
	}
	es := queue.EmailSend{
		Subject:     "test email",
		Content:     "this is a test email",
		ContentType: "text/html",
		UserNames:   []string{"postables"},
		Emails:      []string{"postables@rtradetechnologies.com"},
	}
	if err = qm.PublishMessage(es); err != nil {
		t.Fatal(err)
	}
}
