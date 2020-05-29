package dubbo

import (
	"context"
	"time"

	"mosn.io/api"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/pkg/buffer"
)

func init() {
	api.RegisterStream("dmall_dubbo", CreateDmallDubbo)
}

func CreateDmallDubbo(conf map[string]interface{}) (api.StreamFilterChainFactory, error) {
	return &factory{}, nil
}

type factory struct{}

func (f *factory) CreateFilterChain(ctx context.Context, callbacks api.StreamFilterChainFactoryCallbacks) {
	filter := NewDmallDubboFilter(ctx)
	callbacks.AddStreamReceiverFilter(filter, api.BeforeRoute)
}

type dmallDubboFilter struct {
	handler api.StreamReceiverFilterHandler
}

func NewDmallDubboFilter(ctx context.Context) *dmallDubboFilter {
	return &dmallDubboFilter{}
}

func (d *dmallDubboFilter) OnDestroy() {}

func (d *dmallDubboFilter) OnReceive(ctx context.Context, headers api.HeaderMap, buf buffer.IoBuffer, trailers api.HeaderMap) api.StreamFilterStatus {
	// spew.Dump(headers)

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

func (d *dmallDubboFilter) SetReceiveFilterHandler(handler api.StreamReceiverFilterHandler) {
}
