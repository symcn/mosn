/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dubbodm

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"

	hessian "github.com/apache/dubbo-go-hessian2"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/mosn/pkg/types"
	"mosn.io/pkg/buffer"
)

// Decoder is heavy and caches to improve performance.
// Avoid allocating 4k memory every time you create an object
var (
	decodePoolCheap = &sync.Pool{
		New: func() interface{} {
			return hessian.NewCheapDecoderWithSkip([]byte{})
		},
	}
	decodePool = &sync.Pool{
		New: func() interface{} {
			return hessian.NewDecoderWithSkip([]byte{})
		},
	}
)

func decodeFrame(ctx context.Context, data types.IoBuffer) (cmd interface{}, err error) {
	// convert data to dubbo frame
	dataBytes := data.Bytes()
	frame := &Frame{
		Header: Header{
			CommonHeader: protocol.CommonHeader{},
		},
	}
	// decode magic
	frame.Magic = dataBytes[MagicIdx:FlagIdx]
	// decode flag
	frame.Flag = dataBytes[FlagIdx]
	// decode status
	frame.Status = dataBytes[StatusIdx]
	// decode request id
	reqIDRaw := dataBytes[IdIdx:(IdIdx + IdLen)]
	frame.Id = binary.BigEndian.Uint64(reqIDRaw)
	// decode data length
	frame.DataLen = binary.BigEndian.Uint32(dataBytes[DataLenIdx:(DataLenIdx + DataLenSize)])

	// decode event
	frame.IsEvent = (frame.Flag & (1 << 5)) != 0

	// decode twoway
	frame.IsTwoWay = (frame.Flag & (1 << 6)) != 0

	// decode direction
	directionBool := frame.Flag & (1 << 7)
	if directionBool != 0 {
		frame.Direction = EventRequest
	} else {
		frame.Direction = EventResponse
	}
	// decode serializationId
	frame.SerializationId = int(frame.Flag & 0x1f)

	frameLen := HeaderLen + frame.DataLen
	// decode payload
	body := make([]byte, frameLen)
	copy(body, dataBytes[:frameLen])
	frame.payload = body[HeaderLen:]
	frame.content = buffer.NewIoBufferBytes(frame.payload)

	// not heartbeat & is request
	if !frame.IsEvent && frame.Direction == EventRequest {
		// service aware
		err := getServiceAwareMeta(ctx, frame)
		if err != nil {
			return nil, err
		}
	}

	frame.rawData = body
	frame.data = buffer.NewIoBufferBytes(frame.rawData)
	data.Drain(int(frameLen))
	return frame, nil
}

func getServiceAwareMeta(ctx context.Context, frame *Frame) error {
	if frame.SerializationId != 2 {
		// not hessian , do not support
		return fmt.Errorf("[xprotocol][dubbo] not hessian,do not support")
	}

	decoder := decodePoolCheap.Get().(*hessian.Decoder)
	decoder.Reset(frame.payload[:])

	// Recycle decode
	defer decodePoolCheap.Put(decoder)

	var (
		field               interface{}
		err                 error
		ok                  bool
		attachmentOffsetKey string
		offset              int
	)

	// framework version + path + version + method
	// get service name
	field, err = decoder.Decode()
	if err != nil {
		return fmt.Errorf("[xprotocol][dubbo] decode framework version fail:%s", err)
	}
	// attachments
	attachmentOffsetKey, ok = field.(string)
	if !ok {
		return fmt.Errorf("[xprotocol][dubbo] decode framework version {%v} type error", attachmentOffsetKey)
	}
	offset, err = strconv.Atoi(attachmentOffsetKey)
	if err != nil || offset < 1 {
		return fmt.Errorf("framework version {%s} is not number string or negetive number, err:%v", attachmentOffsetKey, err)
	}
	if offset >= len(frame.payload) {
		return fmt.Errorf("illegal offset number {%d} must <= body length {%d}", offset, len(frame.payload))
	}

	// reset attachement to decode
	decoder.Reset(frame.payload[offset:])

	field, err = decoder.Decode()
	if err != nil {
		return fmt.Errorf("[xprotocol][dubbo] decode dubbo attachments error, %v", err)
	}
	if field != nil {
		if origin, ok := field.(map[interface{}]interface{}); ok {
			for k, v := range origin {
				if key, ok := k.(string); ok {
					if val, ok := v.(string); ok {
						switch key {
						case InterfaceNameHeader:
							// interface => service
							key = ServiceNameHeader
						case PathNameHeadere:
							// path => method
							key = MethodNameHeader
						default:
						}

						// meta[key] = val
						frame.Set(key, val)
					}
				}
			}
		}
	}
	return nil
}
