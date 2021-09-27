package etcdv3

import (
	"context"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"github.com/xqk/ox/pkg/client/etcdv3"
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/util/ogo"
	"time"
)

type etcdv3DataSource struct {
	propertyKey         string
	lastUpdatedRevision int64
	client              *etcdv3.Client
	// cancel is the func, call cancel will stop watching on the propertyKey
	cancel context.CancelFunc
	// closed indicate whether continuing to watch on the propertyKey
	// closed util.AtomicBool

	logger *olog.Logger

	changed chan struct{}
}

// NewDataSource new a etcdv3DataSource instance.
// client is the etcdv3 client, it must be useful and should be release by User.
func NewDataSource(client *etcdv3.Client, key string) conf.DataSource {
	ds := &etcdv3DataSource{
		client:      client,
		propertyKey: key,
	}
	ogo.Go(ds.watch)
	return ds
}

// ReadConfig ...
func (s *etcdv3DataSource) ReadConfig() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := s.client.Get(ctx, s.propertyKey)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, errors.New("empty response")
	}
	s.lastUpdatedRevision = resp.Header.GetRevision()
	return resp.Kvs[0].Value, nil
}

// IsConfigChanged ...
func (s *etcdv3DataSource) IsConfigChanged() <-chan struct{} {
	return s.changed
}

func (s *etcdv3DataSource) handle(resp *clientv3.WatchResponse) {
	if resp.CompactRevision > s.lastUpdatedRevision {
		s.lastUpdatedRevision = resp.CompactRevision
	}
	if resp.Header.GetRevision() > s.lastUpdatedRevision {
		s.lastUpdatedRevision = resp.Header.GetRevision()
	}

	if err := resp.Err(); err != nil {
		return
	}

	for _, ev := range resp.Events {
		if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
			select {
			case s.changed <- struct{}{}:
			default:
			}
		}
	}
}

func (s *etcdv3DataSource) watch() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	rch := s.client.Watch(ctx, s.propertyKey, clientv3.WithCreatedNotify(), clientv3.WithRev(s.lastUpdatedRevision))
	for {
		for resp := range rch {
			s.handle(&resp)
		}
		time.Sleep(time.Second)

		ctx, cancel = context.WithCancel(context.Background())
		if s.lastUpdatedRevision > 0 {
			rch = s.client.Watch(ctx, s.propertyKey, clientv3.WithCreatedNotify(), clientv3.WithRev(s.lastUpdatedRevision))
		} else {
			rch = s.client.Watch(ctx, s.propertyKey, clientv3.WithCreatedNotify())
		}
		s.cancel = cancel
	}
}

// Close ...
func (s *etcdv3DataSource) Close() error {
	s.cancel()
	return nil
}
