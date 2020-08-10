package dubbo

import (
	"context"

	"mosn.io/api"
	mosnctx "mosn.io/mosn/pkg/context"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/mosn/pkg/protocol/xprotocol/dubbo"
	"mosn.io/mosn/pkg/types"
	"mosn.io/pkg/buffer"
)

func init() {
	api.RegisterStream("dubbo_stream", CreateDubbo)
}

func CreateDubbo(conf map[string]interface{}) (api.StreamFilterChainFactory, error) {
	return &factory{}, nil
}

type factory struct{}

func (f *factory) CreateFilterChain(ctx context.Context, callbacks api.StreamFilterChainFactoryCallbacks) {
	filter := NewDubboFilter(ctx)
	callbacks.AddStreamReceiverFilter(filter, api.BeforeRoute)
	callbacks.AddStreamSenderFilter(filter)
}

type dubboFilter struct {
	handler api.StreamReceiverFilterHandler
}

func NewDubboFilter(ctx context.Context) *dubboFilter {
	return &dubboFilter{}
}

func (d *dubboFilter) OnDestroy() {}

func (d *dubboFilter) OnReceive(ctx context.Context, headers api.HeaderMap, buf buffer.IoBuffer, trailers api.HeaderMap) api.StreamFilterStatus {

	listener := mosnctx.Get(ctx, types.ContextKeyListenerName).(string)

	service, ok := headers.Get(dubbo.ServiceNameHeader)
	if ok {
		// adapte dubbo service to http host
		headers.Set(protocol.MosnHeaderHostKey, service)
	}
	// because use http rule, so should add default path
	headers.Set(protocol.MosnHeaderPathKey, "/")

	method, _ := headers.Get(dubbo.MethodNameHeader)
	stats := GetStatus(listener, service, method)
	if stats != nil {
		stats.RequestServiceInfo.Inc(1)
	}

	for k, v := range types.GetPodLabels() {
		if _, ok = headers.Get(k); !ok {
			headers.Set(k, v)
		}
	}

	ctx = mosnctx.WithValue(ctx, types.ContextKeyService, service)
	ctx = mosnctx.WithValue(ctx, types.ContextKeyMethod, method)

	return api.StreamFilterContinue
}

func (d *dubboFilter) SetReceiveFilterHandler(handler api.StreamReceiverFilterHandler) {
}

func (d *dubboFilter) Append(ctx context.Context, headers api.HeaderMap, buf buffer.IoBuffer, trailers api.HeaderMap) api.StreamFilterStatus {
	frame, ok := headers.(*dubbo.Frame)
	if ok {
		listener := mosnctx.Get(ctx, types.ContextKeyListenerName).(string)
		service := mosnctx.Get(ctx, types.ContextKeyService).(string)
		method := mosnctx.Get(ctx, types.ContextKeyMethod).(string)

		stats := GetStatus(listener, service, method)
		if stats != nil {
			if frame.GetStatusCode() == 20 {
				stats.ResponseSucc.Inc(1)
			} else {
				stats.ResponseFail.Inc(1)
			}
		}
	}

	return api.StreamFilterContinue
}

func (d *dubboFilter) SetSenderFilterHandler(handler api.StreamSenderFilterHandler) {
}
