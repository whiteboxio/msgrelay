package corev1alpha1

import (
	"fmt"
	"sync"
)

type MsgStatus uint8

const (
	// MsgStatusNew represents a new message.
	MsgStatusNew MsgStatus = iota
	// MsgStatusDone represents a message that has left the pipeline.
	MsgStatusDone
	// MsgStatusPartialSend represents a partially sent message: most probably,
	// some branches of the pipeline succeeded to send it and some failed.
	MsgStatusPartialSend
	// MsgStatusInvalid represents a message recognised as invalid by the
	// pipeline components and therefore it's impossible to proceed forward.
	MsgStatusInvalid
	// MsgStatusFailed represents a message for which submission has failed.
	MsgStatusFailed
	// MsgStatusTimedOut represents a message for which one or more components
	// triggered a timeout watermark.
	MsgStatusTimedOut
	// MsgStatusUnroutable represents a message for which the submission
	// destination/branch is unknown. Most likely, a branch with the
	// corresponding name does not exist.
	MsgStatusUnroutable
	// MsgStatusThrottled represents a message which submission process was
	// cancelled due to a quota exhausting.
	MsgStatusThrottled
)

var (
	MsgCompletedBeforeErr = fmt.Errorf("message has been completed before")
)

type Message struct {
	body   []byte
	done   chan struct{}
	meta   map[interface{}]interface{}
	status MsgStatus
	mutex  sync.Mutex
}

func NewMessage(body []byte) *Message {
	cpbody := make([]byte, len(body))
	copy(cpbody, body)
	return &Message{
		body:   cpbody,
		done:   make(chan struct{}),
		meta:   make(map[interface{}]interface{}),
		status: MsgStatusNew,
	}
}

func (msg *Message) Await() MsgStatus {
	<-msg.done
	return msg.status
}

func (msg *Message) AwaitChan() <-chan MsgStatus {
	res := make(chan MsgStatus)
	go func() {
		<-msg.done
		res <- msg.status
		close(res)
	}()
	return res
}

func (msg *Message) Complete(status MsgStatus) error {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()
	if msg.status != MsgStatusNew {
		return MsgCompletedBeforeErr
	}
	msg.status = status
	close(msg.done)
	return nil
}

func (msg *Message) Body() []byte {
	return msg.body
}

func (msg *Message) SetBody(body []byte) {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()
	msg.body = body
}

func (msg *Message) MetaKeys() []interface{} {
	res := make([]interface{}, 0, len(msg.meta))
	for k := range msg.meta {
		res = append(res, k)
	}
	return res
}

func (msg *Message) Meta(key interface{}) (interface{}, bool) {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()
	return msg.unsafeMeta(key)
}

func (msg *Message) unsafeMeta(key interface{}) (interface{}, bool) {
	val, ok := msg.meta[key]
	return val, ok
}

func (msg *Message) SetMeta(key, val interface{}) {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()
	msg.unsafeSetMeta(key, val)
}

func (msg *Message) unsafeSetMeta(key, val interface{}) {
	msg.meta[key] = val
}

func (msg *Message) Copy() *Message {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()
	return msg.unsafeCopy()
}

func (msg *Message) unsafeCopy() *Message {
	cpmsg := NewMessage(msg.Body())
	for key, val := range msg.meta {
		cpmsg.unsafeSetMeta(key, val)
	}
	return cpmsg
}
