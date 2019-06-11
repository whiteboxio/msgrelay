package actor

import (
	"sync"

	core "github.com/awesome-flow/flow/pkg/corev1alpha1"
)

type Router struct {
	name  string
	ctx   *core.Context
	rtmap map[string]chan *core.Message
	lock  sync.Mutex
}

var _ core.Actor = (*Router)(nil)

func NewRouter(name string, ctx *core.Context, params core.Params) (core.Actor, error) {
	return &Router{
		name: name,
		ctx:  ctx,
		lock: sync.Mutex{},
	}, nil
}

func (r *Router) Name() string {
	return r.name
}

func (r *Router) Connect(nthreads int, peer core.Receiver) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	peername := peer.(core.Namer).Name()
	if _, ok := r.rtmap[peername]; !ok {
		r.rtmap[peername] = make(chan *core.Message)
	}
	queue := r.rtmap[peername]
	for i := 0; i < nthreads; i++ {
		go func() {
			for msg := range queue {
				if err := peer.Receive(msg); err != nil {
					r.ctx.Logger().Error(err.Error())
				}
			}
		}()
	}
	return nil
}

func (r *Router) Receive(msg *core.Message) error {
	if rtkey, ok := msg.Meta("sendto"); ok {
		if queue, ok := r.rtmap[rtkey.(string)]; ok {
			queue <- msg
			return nil
		}
	}
	msg.Complete(core.MsgStatusUnroutable)
	return nil
}

func (r *Router) Start() error {
	return nil
}

func (r *Router) Stop() error {
	return nil
}