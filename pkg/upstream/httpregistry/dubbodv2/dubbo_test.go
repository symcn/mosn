package dubbodv2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	registry "github.com/symcn/registry/dubbo"
	dubbocommon "github.com/symcn/registry/dubbo/common"
	zkreg "github.com/symcn/registry/dubbo/zookeeper"
	"mosn.io/mosn/pkg/upstream/cluster"
	_ "mosn.io/mosn/pkg/upstream/cluster"
)

func init() {
	Init()

	monkey.Patch(zkreg.NewZkRegistry, func(url *dubbocommon.URL) (registry.Registry, error) {
		return &registry.BaseRegistry{}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(&registry.BaseRegistry{}), "Register", func(r *registry.BaseRegistry, conf *dubbocommon.URL) error {
		return nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(&registry.BaseRegistry{}), "Subscribe", func(r *registry.BaseRegistry, url *dubbocommon.URL, notifyListener registry.NotifyListener) error {
		// do nothing
		return nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(&registry.BaseRegistry{}), "ConnectState", func(r *registry.BaseRegistry) bool {
		return true
	})

	cluster.NewClusterManagerSingleton(nil, nil, nil)

}

func TestPubBindFail(t *testing.T) {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/registry/info/sync", registryInfoSync)

	req, _ := http.NewRequest("POST", "/registry/info/sync", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)

	var res ResponseInfo
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.Nil(t, err)
	assert.Equal(t, res.Errno, fail)
}

func TestPub(t *testing.T) {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/registry/info/sync", registryInfoSync)

	req, _ := http.NewRequest("POST", "/registry/info/sync", strings.NewReader(testRequestStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)

	var res ResponseInfo
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.Nil(t, err)
	assert.Equal(t, res.Errno, succ)
}

func TestSubBindFail(t *testing.T) {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/registry/info/sync", registryInfoSync)

	req, _ := http.NewRequest("POST", "/registry/info/sync", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)

	var res ResponseInfo
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.Nil(t, err)
	assert.Equal(t, res.Errno, fail)
}

func TestSub(t *testing.T) {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/registry/info/sync", registryInfoSync)

	req, _ := http.NewRequest("POST", "/registry/info/sync", strings.NewReader(testRequestStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)

	var res ResponseInfo
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.Nil(t, err)
	assert.Equal(t, res.Errno, succ)
}

func TestUnsub(t *testing.T) {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/registry/info/sync", registryInfoSync)

	req, _ := http.NewRequest("POST", "/registry/info/sync", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)
}

func TestUnpub(t *testing.T) {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/registry/info/sync", registryInfoSync)

	req, _ := http.NewRequest("POST", "/registry/info/sync", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)
}

var testRequestStr = `
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
`
