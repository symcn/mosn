package registry

import (
	"encoding/binary"
	"net"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/samuel/go-zookeeper/zk"
	"k8s.io/klog"
	"mosn.io/api"
	"mosn.io/pkg/buffer"
)

type zkRegistry struct {
	upStream net.Conn
	write    chan []byte
	upWrite  chan []byte
	cb       api.ReadFilterCallbacks
}

func NewZkRegistry() *zkRegistry {
	z := &zkRegistry{
		write:   make(chan []byte, 1),
		upWrite: make(chan []byte, 1),
	}

	var err error
	z.upStream, err = net.Dial("tcp", "127.0.0.1:2181")
	if err != nil {
		klog.Fatal(err)
	}

	go z.upStreamRead()
	go z.upStreamWrite()
	go z.writeData()
	return z
}

func (z *zkRegistry) OnData(buf buffer.IoBuffer) api.FilterStatus {
	d := buf.Bytes()

	var err error
	n := int(binary.BigEndian.Uint32(d[:4]))
	n += 4
	data := make([]byte, n)
	copy(data, d[:n])

	spew.Dump("downsteam request", data, n)

	if n > 12 {
		r := &requestHeader{}
		n, err = decodePacket(data[4:12], r)
		if err == nil && r.Opcode == opCreate {
			rr := &CreateRequest{}
			l := binary.BigEndian.Uint32(data[:4])
			decodePacket(data[12:l], rr)

			rr.Path = rr.Path + time.Now().Format("2006-01-02 15:04:05")

			dd := make([]byte, 4*1024)
			nh, err := encodePacket(dd[4:], r)
			if err != nil {
				klog.Error(err)
			}
			nd, err := encodePacket(dd[4+nh:], rr)
			if err != nil {
				klog.Error(err)
			}
			binary.BigEndian.PutUint32(dd[:4], uint32(nh+nd))

			data = make([]byte, nh+nd+4)
			copy(data, dd[:4+nh+nd])
		}
	}

	z.upWrite <- data
	buf.Drain(buf.Len())

	return api.Continue
}

func (z *zkRegistry) OnNewConnection() api.FilterStatus {
	if z.upStream != nil {
		z.upStream.Close()
	}

	return api.Continue
}

func (z *zkRegistry) InitializeReadFilterCallbacks(cb api.ReadFilterCallbacks) {
	klog.Info("InitializeReadFilterCallbacks..........")
	z.cb = cb
	return
}

func (z *zkRegistry) upStreamWrite() {
	for {
		data, ok := <-z.upWrite
		if !ok {
			return
		}
		spew.Dump("upstream request", data)
		z.upStream.Write(data)
	}
}

func (z *zkRegistry) upStreamRead() {
	for {
		buf := make([]byte, 1024)
		n, err := z.upStream.Read(buf)
		if err != nil {
			klog.Fatal(err)
		}
		if n == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		spew.Dump("upStream response", data)
		z.write <- data
	}
}

func (z *zkRegistry) writeData() {
	for {
		data, ok := <-z.write
		if !ok {
			// klog.Infof("%s close conn", z.conn.RemoteAddr().String())
			return
		}
		spew.Dump("write downstream", data)

		z.cb.Connection().Write(buffer.NewIoBufferBytes(data))
		// z.conn.Write(data)
	}
}

func decodePacket(buf []byte, st interface{}) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(runtime.Error); ok && strings.HasPrefix(e.Error(), "runtime error: slice bounds out of range") {
				err = zk.ErrShortBuffer
			} else {
				panic(r)
			}
		}
	}()

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, zk.ErrPtrExpected
	}
	return decodePacketValue(buf, v)
}

func decodePacketValue(buf []byte, v reflect.Value) (int, error) {
	rv := v
	kind := v.Kind()
	if kind == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
		kind = v.Kind()
	}

	n := 0
	switch kind {
	default:
		return n, zk.ErrUnhandledFieldType
	case reflect.Struct:
		if de, ok := rv.Interface().(decoder); ok {
			return de.Decode(buf)
		} else if de, ok := v.Interface().(decoder); ok {
			return de.Decode(buf)
		} else {
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				n2, err := decodePacketValue(buf[n:], field)
				n += n2
				if err != nil {
					return n, err
				}
			}
		}
	case reflect.Bool:
		v.SetBool(buf[n] != 0)
		n++
	case reflect.Int32:
		v.SetInt(int64(binary.BigEndian.Uint32(buf[n : n+4])))
		n += 4
	case reflect.Int64:
		v.SetInt(int64(binary.BigEndian.Uint64(buf[n : n+8])))
		n += 8
	case reflect.String:
		ln := int(binary.BigEndian.Uint32(buf[n : n+4]))
		v.SetString(string(buf[n+4 : n+4+ln]))
		n += 4 + ln
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		default:
			count := int(binary.BigEndian.Uint32(buf[n : n+4]))
			n += 4
			values := reflect.MakeSlice(v.Type(), count, count)
			v.Set(values)
			for i := 0; i < count; i++ {
				n2, err := decodePacketValue(buf[n:], values.Index(i))
				n += n2
				if err != nil {
					return n, err
				}
			}
		case reflect.Uint8:
			ln := int(int32(binary.BigEndian.Uint32(buf[n : n+4])))
			if ln < 0 {
				n += 4
				v.SetBytes(nil)
			} else {
				bytes := make([]byte, ln)
				copy(bytes, buf[n+4:n+4+ln])
				v.SetBytes(bytes)
				n += 4 + ln
			}
		}
	}
	return n, nil
}

type decoder interface {
	Decode(buf []byte) (int, error)
}

type encoder interface {
	Encode(buf []byte) (int, error)
}

type requestHeader struct {
	Xid    int32
	Opcode int32
}

type ACL struct {
	Perms  int32
	Scheme string
	ID     string
}

type CreateRequest struct {
	Path  string
	Data  []byte
	Acl   []ACL
	Flags int32
}

const (
	opNotify       = 0
	opCreate       = 1
	opDelete       = 2
	opExists       = 3
	opGetData      = 4
	opSetData      = 5
	opGetAcl       = 6
	opSetAcl       = 7
	opGetChildren  = 8
	opSync         = 9
	opPing         = 11
	opGetChildren2 = 12
	opCheck        = 13
	opMulti        = 14
	opReconfig     = 16
	opClose        = -11
	opSetAuth      = 100
	opSetWatches   = 101
	opError        = -1
	// Not in protocol, used internally
	opWatcherEvent = -2
)

func encodePacket(buf []byte, st interface{}) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(runtime.Error); ok && strings.HasPrefix(e.Error(), "runtime error: slice bounds out of range") {
				err = zk.ErrShortBuffer
			} else {
				panic(r)
			}
		}
	}()

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, zk.ErrPtrExpected
	}
	return encodePacketValue(buf, v)
}

func encodePacketValue(buf []byte, v reflect.Value) (int, error) {
	rv := v
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	n := 0
	switch v.Kind() {
	default:
		return n, zk.ErrUnhandledFieldType
	case reflect.Struct:
		if en, ok := rv.Interface().(encoder); ok {
			return en.Encode(buf)
		} else if en, ok := v.Interface().(encoder); ok {
			return en.Encode(buf)
		} else {
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				n2, err := encodePacketValue(buf[n:], field)
				n += n2
				if err != nil {
					return n, err
				}
			}
		}
	case reflect.Bool:
		if v.Bool() {
			buf[n] = 1
		} else {
			buf[n] = 0
		}
		n++
	case reflect.Int32:
		binary.BigEndian.PutUint32(buf[n:n+4], uint32(v.Int()))
		n += 4
	case reflect.Int64:
		binary.BigEndian.PutUint64(buf[n:n+8], uint64(v.Int()))
		n += 8
	case reflect.String:
		str := v.String()
		binary.BigEndian.PutUint32(buf[n:n+4], uint32(len(str)))
		copy(buf[n+4:n+4+len(str)], []byte(str))
		n += 4 + len(str)
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		default:
			count := v.Len()
			startN := n
			n += 4
			for i := 0; i < count; i++ {
				n2, err := encodePacketValue(buf[n:], v.Index(i))
				n += n2
				if err != nil {
					return n, err
				}
			}
			binary.BigEndian.PutUint32(buf[startN:startN+4], uint32(count))
		case reflect.Uint8:
			if v.IsNil() {
				binary.BigEndian.PutUint32(buf[n:n+4], uint32(0xffffffff))
				n += 4
			} else {
				bytes := v.Bytes()
				binary.BigEndian.PutUint32(buf[n:n+4], uint32(len(bytes)))
				copy(buf[n+4:n+4+len(bytes)], bytes)
				n += 4 + len(bytes)
			}
		}
	}
	return n, nil
}
