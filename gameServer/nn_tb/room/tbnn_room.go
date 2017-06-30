package room

import (
	. "mj/common/cost"
	"mj/common/msg"
	"mj/gameServer/RoomMgr"
	"mj/gameServer/common"
	"mj/gameServer/common/pk_base"
	"mj/gameServer/common/pk_base/NNBaseLogic"
	"mj/gameServer/db/model"
	"mj/gameServer/user"

	"mj/gameServer/common/room_base"
	"mj/gameServer/db/model/base"
)

func CreaterRoom(args []interface{}) RoomMgr.IRoom {
	info := args[0].(*model.CreateRoomInfo)

	u := args[1].(*user.User)
	retCode := 0
	defer func() {
		if retCode != 0 {
			u.WriteMsg(&msg.L2C_CreateTableFailure{ErrorCode: retCode, DescribeString: "创建房间失败"})
		}
	}()

	if info.KindId != common.KIND_TYPE_TBNN {
		retCode = ErrParamError
		return nil
	}

	temp, ok := base.GameServiceOptionCache.Get(info.KindId, info.ServiceId)
	if !ok {
		retCode = NoFoudTemplate
		return nil
	}
	r := pk_base.NewPKBase(info)
	cfg := &pk_base.NewPKCtlConfig{
		BaseMgr:  room_base.NewRoomBase(),
		DataMgr:  pk_base.NewDataMgr(info.RoomId, u.Id, pk_base.IDX_NOMAL, temp.GameName, temp, r),
		UserMgr:  room_base.NewRoomUserMgr(info.RoomId, info.MaxPlayerCnt, temp),
		LogicMgr: NNBaseLogic.NewNNBaseLogic(),
		TimerMgr: room_base.NewRoomTimerMgr(),
	}
	r.Init(cfg)
	if r == nil {
		retCode = Errunlawful
		return nil
	}

	u.KindID = info.KindId
	u.RoomId = r.DataMgr.GetRoomId()
	RegisterHandler(r)
	return r
}