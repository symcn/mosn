package registry

import (
	"context"
	"net"

	"mosn.io/api"
)

func init() {
	api.RegisterNetwork("registry_zk", CreateRegistryFactory)
}

type zkfilterConfigFactory struct {
	upStream net.Conn
}

func (z *zkfilterConfigFactory) CreateFilterChain(context context.Context, callbacks api.NetWorkFilterChainFactoryCallbacks) {
	zkr := NewZkRegistry()
	callbacks.AddReadFilter(zkr)
	// callbacks.AddWriteFilter(wf api.WriteFilter)
}

func CreateRegistryFactory(conf map[string]interface{}) (api.NetworkFilterChainFactory, error) {

	zkr := &zkfilterConfigFactory{}
	return zkr, nil
}
