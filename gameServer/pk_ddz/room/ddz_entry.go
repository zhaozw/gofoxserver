package room

import (
	"mj/common/msg/pk_ddz_msg"
	"mj/gameServer/common/pk/pk_base"
	"mj/gameServer/db/model"
	"mj/gameServer/user"
)

func NewDDZEntry(info *model.CreateRoomInfo) *DDZ_Entry {
	e := new(DDZ_Entry)
	return e
	e.Entry_base = pk_base.NewPKBase(info)
	return e
}

///主消息入口
type DDZ_Entry struct {
	*pk_base.Entry_base
}

// 叫分(倍数)
func (room *DDZ_Entry) CallScore(args []interface{}) {
	recvMsg := args[0].(*pk_ddz_msg.C2G_DDZ_CallScore)
	u := args[1].(*user.User)

	room.DataMgr.CallScore(u, recvMsg.CallScore)
	return
}

// 用户出牌
func (room *DDZ_Entry) OutCard(args []interface{}) {

}

// 托管
func (room *DDZ_Entry) CTrustee(args []interface{}) {

}

// 空闲状态
func (room *DDZ_Entry) OnEventGameSceneStatusFree(args []interface{}) {

}

// 叫分状态
func (room *DDZ_Entry) OnEventGameSceneStatusCall(args []interface{}) {

}

// 游戏状态
func (room *DDZ_Entry) OnEventGameSceneStatusPlaying(args []interface{}) {

}

// 发送扑克
func (room *DDZ_Entry) GameStartSendCards(args []interface{}) {

}

// 机器人扑克
func (room *DDZ_Entry) AndroidCard(args []interface{}) {

}

// 作弊扑克
func (room *DDZ_Entry) CheatCard(args []interface{}) {

}

// 庄家信息
func (room *DDZ_Entry) BankerInfo(args []interface{}) {

}

// 放弃出牌
func (room *DDZ_Entry) PassCard(args []interface{}) {

}

// 托管
func (room *DDZ_Entry) GTrustee(args []interface{}) {

}
