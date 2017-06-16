package internal

import (
	"mj/common/msg"
	"reflect"

)

////注册rpc 消息
func handleRpc(id interface{}, f interface{}) {
	ChanRPC.Register(id, f)
}

//注册 客户端消息调用
func handlerC2S(m interface{}, h interface{}) {
	msg.Processor.SetRouter(m, ChanRPC)
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func init() {

}

//发送给房间所有人
func SendChatMsgToAll(args []interface{}) {

}

//发送给房间某人
func sendCharMsgToUser(args []interface{}) {

}