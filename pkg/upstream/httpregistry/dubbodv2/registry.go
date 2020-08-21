package dubbodv2

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/symcn/registry/dubbo/common"
	"mosn.io/mosn/pkg/log"
)

var (
	eventQueue = make(chan event, 100)
)

func init() {
	go loopReceiveEvent()
}

func entryQueue(list []event, cur, old map[string]ServiceRegistryInfo, role Role) []event {
	var (
		registryList   = make(map[string]ServiceRegistryInfo, len(cur))
		unRegistryList = make(map[string]ServiceRegistryInfo, len(old))
	)
	// copy cur & old
	for k, v := range cur {
		registryList[k] = v
	}
	for k, v := range old {
		unRegistryList[k] = v
	}

	// need registry
	for s := range old {
		delete(registryList, s)
	}
	for _, req := range registryList {
		list = append(list, event{
			Role:        role,
			Operat:      OpRegistry,
			ServiceInfo: req,
		})
	}

	// need unregistry
	for s := range cur {
		delete(unRegistryList, s)
	}
	for _, req := range unRegistryList {
		list = append(list, event{
			Role:        role,
			Operat:      OpUnRegistry,
			ServiceInfo: req,
		})
	}

	return list
}

func registryReq(req *ServiceRegistrySnap) {
	// build pubList & subList
	pubList := make(map[string]ServiceRegistryInfo, len(req.ProviderList))
	for _, sr := range req.ProviderList {
		pubList[sr.Service.Interface] = sr
	}
	subList := make(map[string]ServiceRegistryInfo, len(req.ConsumerList))
	for _, sr := range req.ConsumerList {
		subList[sr.Service.Interface] = sr
	}

	l.RLock()
	oldVersion := snapVersion
	oldPubList := snapPubList
	oldSubList := snapSubList
	l.RUnlock()

	eventList := []event{}
	eventList = entryQueue(eventList, pubList, oldPubList, RoleProvider)
	eventList = entryQueue(eventList, subList, oldSubList, RoleConsmmer)

	if len(eventList) == 0 {
		return
	}

	log.DefaultLogger.Infof("snap change, should update version")
	if oldVersion != atomic.LoadUint64(&snapVersion) {
		// version change, should drop this modify
		return
	}

	l.Lock()
	if oldVersion != atomic.LoadUint64(&snapVersion) {
		l.Unlock()
		return
	}
	snapPubList = pubList
	snapSubList = subList
	version := atomic.AddUint64(&snapVersion, 1)
	l.Unlock()

	for _, evt := range eventList {
		evt.Version = version
		eventQueue <- evt
	}
}

func loopReceiveEvent() {

	var (
		succ bool
	)

	for {
		select {
		case evt, ok := <-eventQueue:
			if !ok {
				return
			}
			succ = false
			log.DefaultLogger.Infof("%s %s service {%s}", evt.Operat, evt.Role, evt.ServiceInfo.Service.Interface)

			for {
				_, err := getRegistryWithCheck(common.PROVIDER)
				if err != nil {
					log.DefaultLogger.Warnf("zk connect failed: err:%+v", err)
					time.Sleep(time.Second * 1)
					continue
				}
				break
			}

			err := eventHandler(evt)
			if err == nil {
				succ = true
			} else if evt.Operat == OpRegistry && strings.Contains(err.Error(), zkNodeHasBeenRegisteredErr) {
				log.DefaultLogger.Infof("%s %s service {%s} succ: %+v", evt.Operat, evt.Role, evt.ServiceInfo.Service.Interface, err)
				succ = true
			} else if evt.Operat == OpUnRegistry && strings.Contains(err.Error(), zkNodeHasNotRegisteredErr) {
				log.DefaultLogger.Infof("%s %s service {%s} succ: %+v", evt.Operat, evt.Role, evt.ServiceInfo.Service.Interface, err)
				succ = true
			} else {
				log.DefaultLogger.Errorf("%s %s service {%s} failed: %+v, err: %+v", evt.Operat, evt.Role, evt.ServiceInfo.Service.Interface, evt, err)
			}

			if succ {
				l.Lock()
				switch {
				case evt.Role == RoleProvider && evt.Operat == OpRegistry:
					alreadyPubList[evt.ServiceInfo.Service.Interface] = struct{}{}

				case evt.Role == RoleProvider && evt.Operat == OpUnRegistry:
					delete(alreadyPubList, evt.ServiceInfo.Service.Interface)

				case evt.Role == RoleConsmmer && evt.Operat == OpRegistry:
					alreadySubList[evt.ServiceInfo.Service.Interface] = struct{}{}

				case evt.Role == RoleConsmmer && evt.Operat == OpUnRegistry:
					delete(alreadySubList, evt.ServiceInfo.Service.Interface)

				default:
					log.DefaultLogger.Errorf("[loopReceiveEvent] not define evt:%+v", evt)
				}
				l.Unlock()

				// reduce cpu calc
				arrangeAlreadySlice()
				continue
			}

			// exec fail, judge need re-entry
			time.Sleep(GetZkInterval())
			afterEventHandler(evt)
		}
	}
}

func arrangeAlreadySlice() {
	l.Lock()
	defer l.Unlock()

	plist := make([]string, 0, len(alreadyPubList))
	slist := make([]string, 0, len(alreadySubList))

	for sn := range alreadyPubList {
		plist = append(plist, sn)
	}
	for sn := range alreadySubList {
		slist = append(slist, sn)
	}

	snapAlreadyRegistryPubList = plist
	snapAlreadyRegistrySubList = slist
}

func eventHandler(evt event) (err error) {
	switch evt.Role {
	case RoleProvider:
		return doPubUnPub(evt.ServiceInfo, evt.Operat == OpRegistry)
	case RoleConsmmer:
		return doSubUnsub(evt.ServiceInfo, evt.Operat == OpRegistry)
	default:
		return fmt.Errorf("not define event role:%s", evt.Role)
	}
}

func afterEventHandler(evt event) {

	evt.Version = atomic.LoadUint64(&snapVersion)

	// do event failed, should re-entry queue
	var list map[string]ServiceRegistryInfo

	l.RLock()
	switch evt.Role {
	case RoleProvider:
		list = snapPubList
	case RoleConsmmer:
		list = snapSubList
	default:
		log.DefaultLogger.Errorf("[afterEventHandler] not define role:", evt.Role)
	}
	l.RUnlock()

	switch evt.Operat {

	case OpRegistry:
		// registry failed, but new snap exist should re-registry
		log.DefaultLogger.Infof("registry service:%s failed, should re-registry", evt.ServiceInfo.Service.Interface)
		if _, exist := list[evt.ServiceInfo.Service.Interface]; exist {
			log.DefaultLogger.Infof("registry service:%s failed, should re-registry", evt.ServiceInfo.Service.Interface)
			eventQueue <- evt
			return
		}

	case OpUnRegistry:
		// unregistry failed, but new snap not exist, should re-unregistry
		log.DefaultLogger.Infof("unregistry service:%s failed, should re-unregistry", evt.ServiceInfo.Service.Interface)
		if _, exist := list[evt.ServiceInfo.Service.Interface]; !exist {
			log.DefaultLogger.Infof("unregistry service:%s failed, should re-unregistry", evt.ServiceInfo.Service.Interface)
			eventQueue <- evt
			return
		}

	default:
		log.DefaultLogger.Errorf("[afterEventHandler] not define operat:%s", evt.Operat)
	}
}
