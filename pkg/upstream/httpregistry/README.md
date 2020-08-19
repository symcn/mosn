# HttpRegistry

## 服务注册信息同步

### route

`$host:$port/registry/info/sync`

### method

`POST`

### request struct

```golang
type ServiceRegistrySnap struct {
	ProviderList []ServiceRegistryInfo `json:"provider_list"`
	ConsumerList []ServiceRegistryInfo `json:"consumer_list"`
}

type ServiceRegistryInfo struct {
	Service Service `json:"service"`
	Host    string  `json:"host,omitempty"`
	Port    int     `json:"port,omitempty"`
}

type Service struct {
	Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
	Methods   []string               `json:"methods"`                      // eg. GetUser,GetProfile,UpdateName
	Params    map[string]interface{} `json:"params"`
}

```

### example

- request

```json
{
    "provider_list": [
        {
            "service": {
                "interface": "a",
                "methods": [
                    "a",
                    "b",
                    "c"
                ]
            },
            "host": "1.2.3.4",
            "port": 8080
        },
        {
            "service": {
                "interface": "b",
                "methods": [
                    "a",
                    "b",
                    "c"
                ]
            },
            "host": "1.2.3.4",
            "port": 8080
        }
    ],
    "consumer_list": [
        {
            "service": {
                "interface": "c",
                "methods": [
                    "a",
                    "b",
                    "c"
                ]
            },
            "host": "1.2.3.4",
            "port": 8080
        },
        {
            "service": {
                "interface": "d",
                "methods": [
                    "a",
                    "b",
                    "c"
                ]
            },
            "host": "1.2.3.4",
            "port": 8080
        }
    ]
}
```

- succ

```json
{"code":0,"msg":"registry service success","service_list":{"pub_interface_list":["a","b"],"sub_interface_list":["d","c"],"version":3}}
```

- fail

```json
{"code":1,"msg":"zk not connected.","service_list":{"version":0}}
```

## 说明

response 返回的结果 0 表示成功， 1 表示失败。如果返回 0 都会有已经注册的服务列表，用于 SDK 逻辑判断

**version 每次修改都会自增**
