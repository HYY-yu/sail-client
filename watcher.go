package sail

import (
	"context"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Watcher interface {
	Run()
}

type etcdWatcher struct {
	s *Sail

	ctx    context.Context
	cancel context.CancelFunc
}

func NewWatcher(ctx context.Context, s *Sail) Watcher {
	ctx, cancel := context.WithCancel(ctx)

	etcdW := &etcdWatcher{
		s:      s,
		ctx:    ctx,
		cancel: cancel,
	}
	return etcdW
}

func (e *etcdWatcher) Run() {
	if e.s.etcdClient.Watcher == nil {
		return
	}

	wc := e.s.etcdClient.Watch(
		e.ctx,
		e.s.getETCDKeyPrefix(),
		clientv3.WithPrefix(),
	)

	go func() {
		for {
			select {
			case we := <-wc:
				for _, ev := range we.Events {
					switch ev.Type {
					case mvccpb.PUT:
						isPublish, _ := e.s.checkPublish(ev.Kv.Value)
						if isPublish {
							// 忽略 Publish 消息推送
							continue
						}

						e.dealETCDMsg(string(ev.Kv.Key), ev.Kv.Value)
					case mvccpb.DELETE:
						//do nothing with delete event
					}
				}
			case <-e.ctx.Done():
				e.s.l.Info("close etcd watch, bye~ ")
				return
			}
		}
	}()
}

func (e *etcdWatcher) dealETCDMsg(key string, value []byte) {
	e.s.l.Debug("got a event by: ", "key", key)
	if len(value) == 0 {
		return
	}
	configFileKey := getConfigFileKeyFrom(key)

	viperETCD, err := e.s.newViperWithETCDValue(configFileKey, value)
	if err != nil {
		e.s.l.Error("deal msg error: ", "err", err, "key", configFileKey, "value", string(value))
		return
	}

	e.s.lock.Lock()
	e.s.vipers[configFileKey] = viperETCD
	e.s.lock.Unlock()

	e.s.fm.asyncWriteConfigFile(configFileKey)

	if e.s.changeFunc != nil {
		e.s.changeFunc(key, e.s)
	}
}
