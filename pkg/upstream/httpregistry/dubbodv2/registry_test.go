package dubbodv2

import (
	"reflect"
	"testing"
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
