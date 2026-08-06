package dubbo

import (
	"bytes"
	"strings"
	"sync"

	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/dubbogo/gost/log/logger"
	perrors "github.com/pkg/errors"

	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	invct "dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"dubbo.apache.org/dubbo-go/v3/remoting"
)

// jniTypeCache JNI type conversion cache
var jniTypeCache sync.Map

// MockDubboCodec combines dubbo.DubboCodec, only overriding Decode to support ParameterTypeNames
type MockDubboCodec struct {
	dubbo.DubboCodec
}

// init registers MockDubboCodec to replace native DubboCodec
func init() {
	remoting.RegistryCodec("dubbo", &MockDubboCodec{})
}

// Decode copy from dubbo.DubboCodec
func (c *MockDubboCodec) Decode(data []byte) (*remoting.DecodeResult, int, error) {
	dataLen := len(data)
	if dataLen < impl.HEADER_LENGTH { // check whether header bytes is enough or not
		return nil, 0, nil
	}
	if c.isRequest(data) {
		req, length, err := c.decodeRequest(data)
		if err != nil {
			return nil, length, perrors.WithStack(err)
		}
		if req == ((*remoting.Request)(nil)) {
			return nil, length, err
		}
		return &remoting.DecodeResult{IsRequest: true, Result: req}, length, perrors.WithStack(err)
	}

	rsp, length, err := c.decodeResponse(data)
	if err != nil {
		return nil, length, perrors.WithStack(err)
	}
	if rsp == ((*remoting.Response)(nil)) {
		return nil, length, err
	}
	return &remoting.DecodeResult{IsRequest: false, Result: rsp}, length, perrors.WithStack(err)
}

// decodeRequest copy from dubbo.DubboCodec
// Key change: inject ParameterTypeNames
func (c *MockDubboCodec) decodeRequest(data []byte) (*remoting.Request, int, error) {
	var request *remoting.Request
	buf := bytes.NewBuffer(data)
	pkg := impl.NewDubboPackage(buf)
	pkg.SetBody(make([]interface{}, 7))
	err := pkg.Unmarshal()
	if err != nil {
		originErr := perrors.Cause(err)
		if originErr == hessian.ErrHeaderNotEnough {
			return nil, 0, nil
		}
		if originErr == hessian.ErrBodyNotEnough {
			return nil, hessian.HEADER_LENGTH + pkg.GetBodyLen(), nil
		}
		logger.Errorf("pkg.Unmarshal(len(@data):%d) = error:%+v", buf.Len(), err)
		return request, 0, perrors.WithStack(err)
	}
	request = &remoting.Request{
		ID:       pkg.Header.ID,
		SerialID: pkg.Header.SerialID,
		TwoWay:   pkg.Header.Type&impl.PackageRequest_TwoWay != 0x00,
		Event:    pkg.Header.Type&impl.PackageHeartbeat != 0x00,
	}
	if (pkg.Header.Type & impl.PackageHeartbeat) == 0x00 {
		req := pkg.Body.(map[string]interface{})

		var methodName string
		var args []interface{}
		attachments := make(map[string]interface{})
		if req[impl.DubboVersionKey] != nil {
			request.Version = req[impl.DubboVersionKey].(string)
		}
		attachments[constant.PathKey] = pkg.Service.Path
		attachments[constant.VersionKey] = pkg.Service.Version
		methodName = pkg.Service.Method
		args = req[impl.ArgsKey].([]interface{})
		attachments = req[impl.AttachmentsKey].(map[string]interface{})

		// 提取 argsTypes 并设置到 RPCInvocation
		var typeNames []string
		if argsType, ok := req[impl.ArgsTypesKey].(string); ok {
			argsTypes := strings.Split(strings.TrimRight(argsType, ";"), ";")
			typeNames = make([]string, len(argsTypes))
			for i, t := range argsTypes {
				typeNames[i] = jniToJavaType(t)
			}
		}

		invoc := invct.NewRPCInvocationWithOptions(invct.WithAttachments(attachments),
			invct.WithArguments(args), invct.WithMethodName(methodName), invct.WithParameterTypeNames(typeNames))
		request.Data = invoc
	}
	return request, hessian.HEADER_LENGTH + pkg.Header.BodyLen, nil
}

// decodeResponse  copy from dubbo.DubboCodec
func (c *MockDubboCodec) decodeResponse(data []byte) (*remoting.Response, int, error) {
	buf := bytes.NewBuffer(data)
	pkg := impl.NewDubboPackage(buf)
	err := pkg.Unmarshal()
	if err != nil {
		originErr := perrors.Cause(err)
		// if the data is very big, so the receive need much times.
		if originErr == hessian.ErrHeaderNotEnough { // this is impossible, as dubbo_codec.go:DubboCodec::Decode() line 167
			return nil, 0, nil
		}
		if originErr == hessian.ErrBodyNotEnough {
			return nil, hessian.HEADER_LENGTH + pkg.GetBodyLen(), nil
		}

		logger.Warnf("pkg.Unmarshal(len(@data):%d) = error:%+v", buf.Len(), err)
		return nil, 0, perrors.WithStack(err)
	}
	response := &remoting.Response{
		ID: pkg.Header.ID,
		// Version:  pkg.Header.,
		SerialID: pkg.Header.SerialID,
		Status:   pkg.Header.ResponseStatus,
		Event:    (pkg.Header.Type & impl.PackageHeartbeat) != 0,
	}
	var pkgerr error
	if pkg.Header.Type&impl.PackageHeartbeat != 0x00 {
		if pkg.Header.Type&impl.PackageResponse != 0x00 {
			logger.Debugf("get rpc heartbeat response{header: %#v, body: %#v}", pkg.Header, pkg.Body)
			if pkg.Err != nil {
				logger.Errorf("rpc heartbeat response{error: %#v}", pkg.Err)
				pkgerr = pkg.Err
			}
		} else {
			logger.Debugf("get rpc heartbeat request{header: %#v, service: %#v, body: %#v}", pkg.Header, pkg.Service, pkg.Body)
			response.Status = hessian.Response_OK
			// reply(session, p, hessian.PackageHeartbeat)
		}
		return response, hessian.HEADER_LENGTH + pkg.Header.BodyLen, pkgerr
	}
	logger.Debugf("get rpc response{header: %#v, body: %#v}", pkg.Header, pkg.Body)
	rpcResult := &protocol.RPCResult{}
	response.Result = rpcResult
	if pkg.Header.Type&impl.PackageRequest == 0x00 {
		if pkg.Err != nil {
			rpcResult.Err = pkg.Err
		} else if pkg.Body.(*impl.ResponsePayload).Exception != nil {
			rpcResult.Err = pkg.Body.(*impl.ResponsePayload).Exception
			response.Error = rpcResult.Err
		}
		rpcResult.Attrs = pkg.Body.(*impl.ResponsePayload).Attachments
		rpcResult.Rest = pkg.Body.(*impl.ResponsePayload).RspObj
	}

	return response, hessian.HEADER_LENGTH + pkg.Header.BodyLen, nil
}

// isRequest copy from dubbo.DubboCodec
func (c *MockDubboCodec) isRequest(data []byte) bool {
	return data[2]&byte(0x80) != 0x00
}

// jniToJavaType converts JNI type descriptor to Java type name (with cache)
func jniToJavaType(jniType string) string {
	if cached, ok := jniTypeCache.Load(jniType); ok {
		return cached.(string)
	}

	var result string
	switch jniType[0] {
	case 'B':
		result = "byte"
	case 'C':
		result = "char"
	case 'D':
		result = "double"
	case 'F':
		result = "float"
	case 'I':
		result = "int"
	case 'J':
		result = "long"
	case 'S':
		result = "short"
	case 'Z':
		result = "boolean"
	case 'V':
		result = "void"
	case 'L':
		className := strings.TrimPrefix(jniType, "L")
		result = strings.ReplaceAll(className, "/", ".")
	case '[':
		elementType := jniToJavaType(jniType[1:])
		result = elementType + "[]"
	default:
		result = jniType
	}

	jniTypeCache.Store(jniType, result)
	return result
}
