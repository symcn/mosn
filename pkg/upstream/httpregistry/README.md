# HttpRegistry

_所有的成功请求都会返回当前已经注册的服务的名称 `interface_list`，具体的`$host` `$port`按照实际情况填写_

## provider 注册

### route

`$host:$port/pub`

### method

`POST`

### request struct

```golang
Service struct {
		Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
		Methods   []string               `json:"methods" binding:"required"`   // eg. GetUser,GetProfile,UpdateName
		Params    map[string]interface{} `json:"params"`
	} `json:"service"`
```

### example

- request

```json
{
  "service": {
    "interface": "com.test.cc",
    "methods": [""]
  }
}
```

- succ

```json
{
  "code": 0,
  "msg": "publish success",
  "interface_list": ["com.test.cc", "com.test.cc1"]
}
```

- fail

```json
{
  "code": 1,
  "msg": "publish fail, err: Path{dubbo://:@10.12.214.61:20882/?interface=com.test.cc1&group=blue&version=} has been registered"
}
```

## provider 取消注册

### route

`$host:$port/unpub`

### method

`POST`

### request struct

```golang
type pubReq struct {
	Service struct {
		Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
		Methods   []string               `json:"methods" binding:"required"`   // eg. GetUser,GetProfile,UpdateName
		Params    map[string]interface{} `json:"params"`
	} `json:"service"`
}
```

### example

- request

```json
{
  "service": {
    "interface": "com.test.cc1",
    "methods": [""]
  }
}
```

- succ

```json
{
  "code": 0,
  "msg": "unpub success",
  "interface_list": ["com.test.cc2", "com.test.cc"]
}
```

- fail

```json
{
  "code": 1,
  "msg": "unpub fail, err: Path{dubbo://:@10.12.214.61:20882/?interface=com.test.cc&group=blue&version=} has not registered"
}
```

## consumer 注册

### route

`$host:$port/sub`

### method

`POST`

### request struct

```golang
Service struct {
		Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
		Methods   []string               `json:"methods" binding:"required"`   // eg. GetUser,GetProfile,UpdateName
		Params    map[string]interface{} `json:"params"`
	} `json:"service"`
```

### example

- request

```json
{
  "service": {
    "interface": "com.test.cc",
    "methods": [""]
  }
}
```

- succ

```json
{
  "code": 0,
  "msg": "subscribe success"
}
```

- fail

```json
{
  "code": 1,
  "msg": "publish fail, err: Path{dubbo://:@10.12.214.61:20882/?interface=com.test.cc1&group=blue&version=} has been registered"
}
```

## consumer 取消注册

### route

`$host:$port/unsub`

### method

`POST`

### request struct

```golang
type pubReq struct {
	Service struct {
		Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
		Methods   []string               `json:"methods" binding:"required"`   // eg. GetUser,GetProfile,UpdateName
		Params    map[string]interface{} `json:"params"`
	} `json:"service"`
}
```

### example

- request

```json
{
  "service": {
    "interface": "com.test.cc1",
    "methods": [""]
  }
}
```

- succ

```json
{
  "code": 0,
  "msg": "unsubscribe success"
}
```

- fail

```json
{
  "code": 1,
  "msg": "unpub fail, err: Path{dubbo://:@10.12.214.61:20882/?interface=com.test.cc&group=blue&version=} has not registered"
}
```

## 心跳

### route

`$host:$port/heartbeat`

### method

`GET`

### example

- succ

```json
{
  "code": 0,
  "msg": "ack success",
  "interface_list": ["com.test.cc2", "com.test.cc"]
}
```

- fail

```json
{
  "code": 1,
  "msg": "ack fail timeout"
}
```

## 说明

response 返回的结果 0 表示成功， 1 表示失败。如果返回 0 都会有已经注册的服务列表，用于 SDK 逻辑判断
