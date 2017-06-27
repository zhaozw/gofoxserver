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
	// c 2 s
	//handlerC2S(&mj_hz_msg.C2G_HZMJ_HZOutCard{}, HZOutCard)
	//handlerC2S(&mj_hz_msg.C2G_HZMJ_OperateCard{}, OperateCard)

}

//func HZOutCard(args []interface{}) {
//	agent := args[1].(gate.Agent)
//	user := agent.UserData().(*user.User)
//
//	r := getRoom(user.RoomId)
//	if r != nil {
//		r.GetChanRPC().Go("OutCard", args[0], user)
//	}
//}
//
//func OperateCard(args []interface{}) {
//	agent := args[1].(gate.Agent)
//	user := agent.UserData().(*user.User)
//
//	r := getRoom(user.RoomId)
//	if r != nil {
//		r.GetChanRPC().Go("OperateCard", args[0], user)
//	}
//}