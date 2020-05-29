package dubbo

import (
	"context"
	"time"

	"mosn.io/api"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/pkg/buffer"
)

func init() {
	api.RegisterStream("dubbo_demo", CreateDubbo)
}

func CreateDubbo(conf map[string]interface{}) (api.StreamFilterChainFactory, error) {
	return &factory{}, nil
}

type factory struct{}

func (f *factory) CreateFilterChain(ctx context.Context, callbacks api.StreamFilterChainFactoryCallbacks) {
	filter := NewDubboFilter(ctx)
	callbacks.AddStreamReceiverFilter(filter, api.BeforeRoute)
}

type DubboFilter struct {
	handler api.StreamReceiverFilterHandler
}

func NewDubboFilter(ctx context.Context) *DubboFilter {
	return &DubboFilter{}
}

func (d *DubboFilter) OnDestroy() {}

func (d *DubboFilter) OnReceive(ctx context.Context, headers api.HeaderMap, buf buffer.IoBuffer, trailers api.HeaderMap) api.StreamFilterStatus {
	service, ok := headers.Get("service")
	if ok {
		headers.Set(protocol.MosnHeaderHostKey, service)
	}

	headers.Set(protocol.MosnHeaderPathKey, "/")

	if time.Now().Minute()%2 == 0 {
		headers.Set("zone", "gz01")
	}

	return api.StreamFilterContinue
}

func (d *DubboFilter) SetReceiveFilterHandler(handler api.StreamReceiverFilterHandler) {
}
