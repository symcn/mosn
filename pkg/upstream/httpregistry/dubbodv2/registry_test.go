package dubbodv2

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
)

func Test_entryQueue(t *testing.T) {
	type args struct {
		list []event
		cur  map[string]ServiceRegistryInfo
		old  map[string]ServiceRegistryInfo
		role Role
	}
	tests := []struct {
		name string
		args args
		want []event
	}{
		// TODO: Add test cases.
		{
			name: "registry",
			args: args{
				list: []event{},
				cur: map[string]ServiceRegistryInfo{
					"a": {},
				},
				old:  map[string]ServiceRegistryInfo{},
				role: RoleProvider,
			},
			want: []event{
				{
					Role:        RoleProvider,
					Operat:      OpRegistry,
					ServiceInfo: ServiceRegistryInfo{},
				},
			},
		},
		{
			name: "unregistry",
			args: args{
				list: []event{},
				cur:  map[string]ServiceRegistryInfo{},
				old: map[string]ServiceRegistryInfo{
					"a": {},
				},
				role: RoleProvider,
			},
			want: []event{
				{
					Role:        RoleProvider,
					Operat:      OpUnRegistry,
					ServiceInfo: ServiceRegistryInfo{},
				},
			},
		},
		{
			name: "registry unregistry",
			args: args{
				list: []event{},
				cur: map[string]ServiceRegistryInfo{
					"a": {},
				},
				old: map[string]ServiceRegistryInfo{
					"b": {},
				},
				role: RoleProvider,
			},
			want: []event{
				{
					Role:        RoleProvider,
					Operat:      OpRegistry,
					ServiceInfo: ServiceRegistryInfo{},
				},
				{
					Role:        RoleProvider,
					Operat:      OpUnRegistry,
					ServiceInfo: ServiceRegistryInfo{},
				},
			},
		},
		{
			name: "not change",
			args: args{
				list: []event{},
				cur: map[string]ServiceRegistryInfo{
					"a": {},
				},
				old: map[string]ServiceRegistryInfo{
					"a": {},
				},
				role: RoleProvider,
			},
			want: []event{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entryQueue(tt.args.list, tt.args.cur, tt.args.old, tt.args.role); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entryQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_registryReq(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	type args struct {
		req *ServiceRegistrySnap
	}
	tests := []struct {
		name string
		args args
	}{}

	for i := 0; i < 500; i++ {
		req := generatReqistryReq()
		tests = append(tests, struct {
			name string
			args args
		}{
			// TODO: Add test cases.
			name: fmt.Sprintf("generat test case %d", i),
			args: args{&req},
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registryReq(tt.args.req)
			time.Sleep(time.Second * 1)
			if !judge(tt.args.req) {
				time.Sleep(time.Second * 3)
				if !judge(tt.args.req) {
					t.Errorf("generat test err:req:%+v, list:%+v", tt.args.req, getRegistryInterfaceList())
				}
			}
		})
	}
}

func judge(req *ServiceRegistrySnap) bool {
	serviceList := getRegistryInterfaceList()
	for _, p := range serviceList.PubInterfaceList {
		isExist := false
		for _, ps := range req.ProviderList {
			if p == ps.Service.Interface {
				isExist = true
				break
			}
		}
		if !isExist {
			return false
		}
	}

	for _, s := range serviceList.SubInterfaceList {
		isExist := false
		for _, ss := range req.ConsumerList {
			if s == ss.Service.Interface {
				isExist = true
				break
			}
		}
		if !isExist {
			return false
		}
	}

	return true
}

func generatReqistryReq() ServiceRegistrySnap {

	pn := rand.Intn(30)
	cn := rand.Intn(30)

	snap := ServiceRegistrySnap{
		ProviderList: make([]ServiceRegistryInfo, 0, pn),
		ConsumerList: make([]ServiceRegistryInfo, 0, cn),
	}

	index := rand.Intn(100)
	for i := 0; i < pn; i++ {
		snap.ProviderList = append(snap.ProviderList, generatReqInfo(index))
		index++
	}

	index = rand.Intn(100)
	for i := 0; i < cn; i++ {
		snap.ConsumerList = append(snap.ConsumerList, generatReqInfo(index))
		index++
	}

	fmt.Println(snap)

	return snap
}

var (
	service = []string{"a", "b", "c", "d", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "y", "aa", "bb", "cc", "dd", "ee", "ff", "gg", "hh", "ii", "jj", "kk", "ll"}
	method  = []string{"get", "post", "delete", "option", "put"}
)

func generatReqInfo(i int) ServiceRegistryInfo {
	return ServiceRegistryInfo{
		Service: Service{
			Interface: service[i%len(service)],
			Methods:   []string{method[(i+1)%len(method)], method[(i+2)%len(method)], method[(i+3)%len(method)]},
		},
		Host: fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(254)+1, rand.Intn(254)+1, rand.Intn(254)+1),
		Port: 20882,
	}
}
