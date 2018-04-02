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
// | id | protobuf message |
// -------------------------
type Processor struct {
	littleEndian bool
	msgInfo      map[uint16]*MsgInfo
}

type MsgInfo struct {
	msgType    reflect.Type
	msgRouter  *chanrpc.Server
	msgHandler MsgHandler
}

type MsgHandler func([]interface{})

type MsgRaw struct {
	Id  uint16
	Msg interface{}
}

func NewProcessor() *Processor {
	p := new(Processor)
	p.littleEndian = false
	p.msgInfo = make(map[uint16]*MsgInfo)
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) RegisterWithRouter(id uint16, msg proto.Message, msgRouter *chanrpc.Server) {
	if id >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}
	if _, ok := p.msgInfo[id]; ok {
		log.Fatal("message id:%d  is already registered", id)
	}

	i := new(MsgInfo)
	i.msgRouter = msgRouter

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
	// protobuf
	if raw, ok := msg.(MsgRaw); ok {
		i := p.msgInfo[raw.Id]
		recv := false
		if i.msgHandler != nil {
			i.msgHandler([]interface{}{raw.Msg, userData})
			recv = true
		}
		if i.msgRouter != nil {
			i.msgRouter.Go(raw.Id, raw.Id, raw.Msg, userData)
			recv = true
		}

		if recv == false {
			return fmt.Errorf("cannt find msgid:%d, %v", raw.Id, raw.Msg)
		}
	} else {
		return fmt.Errorf("msgRaw err:%v", msg)
	}

	return nil
}

// goroutine safe
func (p *Processor) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	// id
	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	// msg
	if i, ok := p.msgInfo[id]; ok {
		if i.msgType == nil {
			return MsgRaw{id, nil}, nil
		} else {
			msg := reflect.New(i.msgType.Elem()).Interface()
			return MsgRaw{id, msg}, proto.UnmarshalMerge(data[2:], msg.(proto.Message))
		}
	}
	return nil, fmt.Errorf("message id %v not registered", id)
}

// goroutine safe
func (p *Processor) Marshal(msg interface{}) ([][]byte, error) {
	if raw, ok := msg.(MsgRaw); ok {
		id := make([]byte, 2)
		if p.littleEndian {
			binary.LittleEndian.PutUint16(id, raw.Id)
		} else {
			binary.BigEndian.PutUint16(id, raw.Id)
		}

		// data
		if raw.Msg == nil {
			return [][]byte{id, nil}, nil
		}

		data, err := proto.Marshal(raw.Msg.(proto.Message))
		return [][]byte{id, data}, err
	}

	return nil, fmt.Errorf("message %v not registered\n", msg)

}

// goroutine safe
func (p *Processor) Range(f func(id uint16, t reflect.Type)) {
	for id, i := range p.msgInfo {
		f(uint16(id), i.msgType)
	}
}
