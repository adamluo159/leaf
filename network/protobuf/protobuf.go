package protobuf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/adamluo159/leaf/chanrpc"
	"github.com/adamluo159/leaf/log"
	"github.com/golang/protobuf/proto"
)

// -------------------------
// | id | req | protobuf message |
// -------------------------
type Processor struct {
	littleEndian bool
	msgRouter    *chanrpc.Server
	msgInfo      map[uint16]*MsgInfo
}

type MsgInfo struct {
	msgType    reflect.Type
	msgHandler MsgHandler
}

type MsgRaw struct {
	Id  uint16
	Req uint16
	Msg interface{}
}

type MsgHandler func([]interface{})

func NewProcessor(router *chanrpc.Server) *Processor {
	p := new(Processor)
	p.littleEndian = false
	p.msgInfo = make(map[uint16]*MsgInfo)
	p.msgRouter = router
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) Register(id uint16, msg proto.Message) {
	if id >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}
	if _, ok := p.msgInfo[id]; ok {
		log.Fatal("message id:%d  is already registered", id)
	}

	i := new(MsgInfo)

	p.msgInfo[id] = i
	if msg != nil {
		msgType := reflect.TypeOf(msg)
		i.msgType = msgType
	}
}

func (p *Processor) RegisterWithHandle(id uint16, msg proto.Message, msgHandler MsgHandler) {
	if id >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}
	if _, ok := p.msgInfo[id]; ok {
		log.Fatal("message id:%d  is already registered", id)
	}

	i := new(MsgInfo)
	p.msgInfo[id] = i
	i.msgHandler = msgHandler

	if msg != nil {
		msgType := reflect.TypeOf(msg)
		i.msgType = msgType
	}
}

// goroutine safe
func (p *Processor) Route(msg interface{}, userData interface{}) error {
	raw := msg.(*MsgRaw)
	// protobuf
	i := p.msgInfo[raw.Id]
	if p.msgRouter != nil {
		p.msgRouter.Go(raw.Id, raw, userData)
	} else if i.msgHandler != nil {
		i.msgHandler([]interface{}{raw, userData})
	} else {
		return fmt.Errorf("cannt find msgid:%d, %v", raw.Id, raw.Msg)
	}

	return nil
}

// goroutine safe
func (p *Processor) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 4 {
		return nil, errors.New("protobuf data too short")
	}

	// id
	var id, req uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
		req = binary.LittleEndian.Uint16(data[2:])
	} else {
		id = binary.BigEndian.Uint16(data)
		req = binary.BigEndian.Uint16(data[2:])
	}

	// msg
	if i, ok := p.msgInfo[id]; ok {
		if i.msgType == nil {
			return &MsgRaw{id, req, nil}, nil
		} else {
			if len(data) == 4 {
				return &MsgRaw{id, req, nil}, errors.New(fmt.Sprintf("msg should have body, id:%v, req:%v, %v", id, req, i.msgType))
			}
			msg := reflect.New(i.msgType.Elem()).Interface()
			return &MsgRaw{id, req, msg}, proto.UnmarshalMerge(data[4:], msg.(proto.Message))
		}
	}
	return nil, fmt.Errorf("message id %v not registered", id)
}

// goroutine safe
func (p *Processor) Marshal(msg interface{}) ([][]byte, error) {
	raw := msg.(*MsgRaw)
	pre := make([]byte, 4)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(pre, raw.Id)
		binary.LittleEndian.PutUint16(pre[2:], raw.Req)
	} else {
		binary.BigEndian.PutUint16(pre, raw.Id)
		binary.BigEndian.PutUint16(pre[2:], raw.Req)
	}

	// data
	if raw.Msg == nil {
		return [][]byte{pre, nil}, nil
	}

	data, err := proto.Marshal(raw.Msg.(proto.Message))
	return [][]byte{pre, data}, err
}

// goroutine safe
func (p *Processor) Range(f func(id uint16, t reflect.Type)) {
	for id, i := range p.msgInfo {
		f(uint16(id), i.msgType)
	}
}
