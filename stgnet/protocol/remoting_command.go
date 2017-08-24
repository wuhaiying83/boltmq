package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"github.com/pquerna/ffjson/ffjson"
)

// RemotingCommand 服务器与客户端通过传递RemotingCommand来交互
// Author gaoyanlei
// Since 2017/8/15
const (
	RemotingVersionKey = "rocketmq.remoting.version"
	rpcType            = 0
	rpcOneway          = 1
)

var (
	configVersion = -1
)

// RemotingCommand remoting command
// Author: jerrylou, <gunsluo@gmail.com>
// Since: 2017-08-22
type RemotingCommand struct {
	//header
	Code      int               `json:"code"`
	Language  string            `json:"language"`
	Version   int               `json:"version"`
	Opaque    int32             `json:"opaque"`
	Flag      int               `json:"flag"`
	Remark    string            `json:"remark"`
	ExtFields map[string]string `json:"extFields"`
	// 修改字段类型 2017/8/16 Add by yintongqiang
	// 字段不序列化 Modify: jerrylou, <gunsluo@gmail.com> Since: 2017-08-24
	CustomHeader CommandCustomHeader `json:"-"`
	//body
	Body []byte `json:"-"`
}

// CreateResponseCommand
func CreateResponseCommand(code int, remark string) *RemotingCommand {
	return &RemotingCommand{
		Code:   code,
		Remark: remark,
	}
}

// CreateRequestCommand 创建客户端请求信息 2017/8/16 Add by yintongqiang
func CreateRequestCommand(code int, customHeader CommandCustomHeader) *RemotingCommand {
	remotingClient := &RemotingCommand{
		Code:         code,
		CustomHeader: customHeader,
		ExtFields:    make(map[string]string),
	}
	remotingClient.setCMDVersion()

	return remotingClient
}

// Author: jerrylou, <gunsluo@gmail.com>
// Since: 2017-08-22
func (rc *RemotingCommand) setCMDVersion() {
	if configVersion >= 0 {
		rc.Version = configVersion
		return
	}

	version := os.Getenv(RemotingVersionKey)
	if version == "" {
		return
	}

	v, e := strconv.Atoi(version)
	if e == nil {
		rc.Version = v
	}
}

// IsOnewayRPC is oneway rpc, return bool
func (rc *RemotingCommand) IsOnewayRPC() bool {
	bits := 1 << rpcOneway
	return (rc.Flag & bits) == bits
}

// MarkResponseType mark response type
func (rc *RemotingCommand) MarkResponseType() {
	bits := 1 << rpcType
	rc.Flag |= bits
}

// IsResponseType is response type, return bool
func (rc *RemotingCommand) IsResponseType() bool {
	bits := 1 << rpcType
	return (rc.Flag & bits) == bits
}

// EncodeHeader 编码头部
func (rc *RemotingCommand) EncodeHeader() []byte {
	length := 4
	headerData := rc.buildHeader()
	length += len(headerData)

	if rc.Body != nil {
		length += len(rc.Body)
	}

	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, length)
	binary.Write(buf, binary.BigEndian, len(headerData))
	buf.Write(headerData)

	return buf.Bytes()
}

func (rc *RemotingCommand) buildHeader() []byte {
	buf, err := ffjson.Marshal(rc)
	if err != nil {
		return nil
	}
	return buf
}

// Type return remoting command type
func (rc *RemotingCommand) Type() RemotingCommandType {
	if rc.IsResponseType() {
		return RESPONSE_COMMAND
	}

	return REQUEST_COMMAND
}

// DecodeRemotingCommand 解析返回RemotingCommand
func DecodeRemotingCommand(buf *bytes.Buffer) (*RemotingCommand, error) {
	var (
		length       int32
		headerLength int32
		bodyLength   int32
	)

	// step 1 读取报文长度
	if buf.Len() < 4 {
		return nil, fmt.Errorf("buffer length %d < 4", buf.Len())
	}

	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("read buffer length failed: %v", err)
	}

	// step 2 读取报文头长度
	if buf.Len() < 4 {
		return nil, fmt.Errorf("buffer header length %d < 4", buf.Len())
	}

	err = binary.Read(buf, binary.BigEndian, &headerLength)
	if err != nil {
		return nil, fmt.Errorf("read buffer header length failed: %v", err)
	}

	// step 3 读取报文头数据
	if buf.Len() == 0 || buf.Len() < int(headerLength) {
		return nil, fmt.Errorf("header data invalid, length: %d", buf.Len())
	}

	header := make([]byte, headerLength)
	_, err = buf.Read(header)
	if err != nil {
		return nil, fmt.Errorf("read header data failed: %v", err)
	}

	// step 4 读取报文Body
	bodyLength = length - 4 - headerLength
	if buf.Len() < int(bodyLength) {
		return nil, fmt.Errorf("body length %d < %d", bodyLength, buf.Len())
	}

	body := make([]byte, bodyLength)
	_, err = buf.Read(body)
	if err != nil {
		return nil, fmt.Errorf("read body data failed: %v", err)
	}

	return decodeRemotingCommand(header, body)
}

func decodeRemotingCommand(header, body []byte) (*RemotingCommand, error) {
	remotingCommand := &RemotingCommand{}
	remotingCommand.ExtFields = make(map[string]string)
	err := ffjson.Unmarshal(header, remotingCommand)
	if err != nil {
		return nil, err
	}
	remotingCommand.Body = body
	return remotingCommand, nil
}
