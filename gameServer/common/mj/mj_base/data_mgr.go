package mj_base

import (
	"math"
	. "mj/common/cost"
	"mj/common/msg"
	"mj/common/msg/mj_hz_msg"
	"mj/common/msg/mj_zp_msg"
	"mj/common/utils"
	"mj/gameServer/conf"
	"mj/gameServer/db/model/base"
	"mj/gameServer/user"
	"strconv"
	"time"

	"github.com/lovelly/leaf/log"
	"github.com/lovelly/leaf/util"
)

func NewDataMgr(id, uid, configIdx int, name string, temp *base.GameServiceOption, base *Mj_base) *RoomData {
	r := new(RoomData)
	r.id = id
	if name == "" {
		r.Name = temp.RoomName
	} else {
		r.Name = name
	}
	r.CreateUser = uid
	r.MjBase = base
	r.ConfigIdx = configIdx
	return r
}

//当一张桌子理解
type RoomData struct {
	id         int
	Name       string //房间名字
	CreateUser int    //创建房间的人
	MjBase     *Mj_base
	ConfigIdx  int //配置索引

	IsResponse        []bool //标记是否对吃碰杠胡做出过动作
	PerformAction     []int  //记住玩家出的动作， 用来等待优先级更高的玩家
	Source            int    //底分
	IniSource         int    //初始分数
	IsGoldOrGameScore int    //金币场还是积分场 0 标识 金币场 1 标识 积分场
	Password          string // 密码
	ProvideCard       int    //供应扑克
	ResumeUser        int    //还原用户
	ProvideUser       int    //供应用户
	LeftCardCount     int    //剩下拍的数量
	EndLeftCount      int    //荒庄牌数
	LastCatchCardUser int    //最后一个摸牌的用户
	MinusHeadCount    int    //头部空缺
	MinusLastCount    int    //尾部空缺

	SiceCount       int                //色子大小
	UserActionDone  bool               //操作完成
	SendStatus      int                //发牌状态
	GangStatus      int                //杠牌状态
	GangOutCard     bool               //杠后出牌
	ProvideGangUser int                //供杠用户
	GangCard        []bool             //杠牌状态
	GangCount       []int              //杠牌次数
	RepertoryCard   []int              //库存扑克
	UserGangScore   []int              //游戏中杠的输赢
	ChiHuKind       []int              //吃胡结果
	ChiHuRight      []int              //胡牌类型
	UserAction      []int              //用户动作
	OperateCard     [][]int            //操作扑克
	ChiPengCount    []int              //吃碰杠次数
	CardIndex       [][]int            //用户扑克[GAME_PLAYER][MAX_INDEX]
	WeaveItemArray  [][]*msg.WeaveItem //组合扑克
	DiscardCard     [][]int            //丢弃记录
	OutCardData     int                //出牌扑克
	OutCardUser     int                //当前出牌用户
	HeapHead        int                //堆立头部
	HeapTail        int                //堆立尾部
	HeapCardInfo    [][]int            //堆牌信息
	SendCardData    int                //发牌扑克
	HistoryScores   []*HistoryScore    //历史积分
	CurrentUser     int                //当前操作用户
	Ting            []bool             //是否听牌
	BankerUser      int                //庄家用户
	FlowerCnt       []int              //补花数
}

func (room *RoomData) GetCfg() *MJ_CFG {
	return GetCfg(room.ConfigIdx)
}

func (room *RoomData) CanOperatorRoom(uid int) bool {
	if uid == room.CreateUser {
		return true
	}
	return false
}

func (room *RoomData) GetCurrentUser() int {
	return room.CurrentUser
}

func (room *RoomData) GetGangStatus() int {
	return room.GangStatus
}

func (room *RoomData) GetProvideUser() int {
	return room.ProvideUser
}

func (room *RoomData) GetResumeUser() int {
	return room.ResumeUser
}

func (room *RoomData) GetRoomId() int {
	return room.id
}
func (room *RoomData) SendPersonalTableTip(u *user.User) {
	u.WriteMsg(&msg.G2C_PersonalTableTip{
		TableOwnerUserID:  room.CreateUser,                                               //桌主 I D
		DrawCountLimit:    room.MjBase.TimerMgr.GetMaxPayCnt(),                           //局数限制
		DrawTimeLimit:     room.MjBase.TimerMgr.GetTimeLimit(),                           //时间限制
		PlayCount:         room.MjBase.TimerMgr.GetPlayCount(),                           //已玩局数
		PlayTime:          int(room.MjBase.TimerMgr.GetCreatrTime() - time.Now().Unix()), //已玩时间
		CellScore:         room.Source,                                                   //游戏底分
		IniScore:          room.IniSource,                                                //初始分数
		ServerID:          strconv.Itoa(room.id),                                         //房间编号
		IsJoinGame:        0,                                                             //是否参与游戏 todo  tagPersonalTableParameter
		IsGoldOrGameScore: room.IsGoldOrGameScore,                                        //金币场还是积分场 0 标识 金币场 1 标识 积分场
	})
}

func (room *RoomData) SendStatusReady(u *user.User) {
	StatusFree := &msg.G2C_StatusFree{}
	StatusFree.CellScore = room.Source                                     //基础积分
	StatusFree.TimeOutCard = room.MjBase.TimerMgr.GetTimeOutCard()         //出牌时间
	StatusFree.TimeOperateCard = room.MjBase.TimerMgr.GetTimeOperateCard() //操作时间
	StatusFree.CreateTime = room.MjBase.TimerMgr.GetCreatrTime()           //开始时间
	for _, v := range room.HistoryScores {
		StatusFree.TurnScore = append(StatusFree.TurnScore, v.TurnScore)
		StatusFree.CollectScore = append(StatusFree.TurnScore, v.CollectScore)
	}
	StatusFree.PlayerCount = room.MjBase.TimerMgr.GetPlayCount() //玩家人数
	StatusFree.MaCount = 0                                       //码数
	StatusFree.CountLimit = room.MjBase.TimerMgr.GetMaxPayCnt()  //局数限制
	u.WriteMsg(StatusFree)
}

//注意这个函数仅供调试用
func (room *RoomData) SetUserCard(charirID int, cards []int) {
	log.Debug("begin SetUserCard", room.CardIndex[charirID])
	gameLogic := room.MjBase.LogicMgr
	//repalce := func(old, new int) {
	//	for idx, v := range room.RepertoryCard {
	//		if v == old {
	//			room.RepertoryCard[idx] = new
	//		}
	//	}
	//}

	inc := 0
	userCard := room.CardIndex[charirID]
	for idx, cnt := range userCard {
		for i := 0; i < cnt; i++ {
			if inc >= len(cards) {
				break
			}
			//repalce(cards[inc], gameLogic.SwitchToCardData(idx))
			userCard[idx]--
			userCard[gameLogic.SwitchToCardIndex(cards[inc])]++
			inc++
		}
	}
	log.Debug("end SetUserCard", room.CardIndex[charirID])
}

func (room *RoomData) SendStatusPlay(u *user.User) {
	StatusPlay := &msg.G2C_StatusPlay{}
	//自定规则
	StatusPlay.TimeOutCard = room.MjBase.TimerMgr.GetTimeOutCard()
	StatusPlay.TimeOperateCard = room.MjBase.TimerMgr.GetTimeOperateCard()
	StatusPlay.CreateTime = room.MjBase.TimerMgr.GetCreatrTime()

	//规则
	StatusPlay.PlayerCount = room.MjBase.TimerMgr.GetPlayCount()
	UserCnt := room.MjBase.UserMgr.GetMaxPlayerCnt()
	//游戏变量
	StatusPlay.BankerUser = room.BankerUser
	StatusPlay.CurrentUser = room.OutCardUser
	StatusPlay.CellScore = room.Source
	StatusPlay.MagicIndex = room.MjBase.LogicMgr.GetMagicIndex()
	StatusPlay.Trustee = room.MjBase.UserMgr.GetTrustees()
	StatusPlay.HuCardCount = make([]int, room.GetCfg().MaxCount)
	StatusPlay.HuCardData = make([][]int, room.GetCfg().MaxCount)
	StatusPlay.OutCardDataEx = make([]int, room.GetCfg().MaxCount)
	StatusPlay.CardCount = make([]int, UserCnt)
	StatusPlay.TurnScore = make([]int, UserCnt)
	StatusPlay.CollectScore = make([]int, UserCnt)

	//状态变量
	StatusPlay.ActionCard = room.ProvideCard
	StatusPlay.LeftCardCount = room.LeftCardCount
	StatusPlay.ActionMask = room.UserAction[u.ChairId]

	StatusPlay.Ting = room.Ting
	//当前能胡的牌
	StatusPlay.OutCardCount = room.MjBase.LogicMgr.AnalyseTingCard(room.CardIndex[u.ChairId], room.WeaveItemArray[u.ChairId],
		StatusPlay.OutCardDataEx, StatusPlay.HuCardCount, StatusPlay.HuCardData, room.GetCfg().MaxCount)

	//历史记录
	StatusPlay.OutCardUser = room.OutCardUser
	StatusPlay.OutCardData = room.OutCardData
	StatusPlay.DiscardCard = room.DiscardCard
	for _, v := range room.DiscardCard {
		StatusPlay.DiscardCount = append(StatusPlay.DiscardCount, len(v))
	}

	StatusPlay.WeaveItemArray = room.WeaveItemArray
	for _, v := range room.WeaveItemArray {
		StatusPlay.WeaveItemCount = append(StatusPlay.WeaveItemCount, len(v))
	}

	//堆立信息
	StatusPlay.HeapHead = room.HeapHead
	StatusPlay.HeapTail = room.HeapTail
	StatusPlay.HeapCardInfo = room.HeapCardInfo

	//扑克数据
	for j := 0; j < UserCnt; j++ {
		StatusPlay.CardCount[j] = room.MjBase.LogicMgr.GetCardCount(room.CardIndex[j])
	}

	StatusPlay.CardData = room.MjBase.LogicMgr.GetUserCards(room.CardIndex[u.ChairId])
	if room.CurrentUser == u.ChairId {
		StatusPlay.SendCardData = room.SendCardData
	} else {
		StatusPlay.SendCardData = 0x00
	}

	//历史积分
	for j := 0; j < UserCnt; j++ {
		//设置变量
		if room.HistoryScores[j] != nil {
			StatusPlay.TurnScore[j] = room.HistoryScores[j].TurnScore
			StatusPlay.CollectScore[j] = room.HistoryScores[j].CollectScore
		}
	}

	u.WriteMsg(StatusPlay)
}

func (room *RoomData) NotifySendCard(u *user.User, cbCardData int, bSysOut bool) {
	//设置变量
	room.SendStatus = OutCard_Send
	room.SendCardData = 0
	room.UserAction[u.ChairId] = WIK_NULL

	//出牌记录
	room.OutCardUser = u.ChairId
	room.OutCardData = cbCardData

	//构造数据
	OutCard := &mj_hz_msg.G2C_HZMJ_OutCard{}
	OutCard.OutCardUser = u.ChairId
	OutCard.OutCardData = cbCardData
	OutCard.SysOut = bSysOut

	//发送消息
	room.MjBase.UserMgr.SendMsgAll(OutCard)
	room.ProvideUser = u.ChairId
	room.ProvideCard = cbCardData

	//用户切换
	room.CurrentUser = (u.ChairId + 1) % room.MjBase.UserMgr.GetMaxPlayerCnt()
}

func (room *RoomData) GetUserCardIndex(ChairId int) []int {
	return room.CardIndex[ChairId]
}

//检测是否可以做某个操作
func (room *RoomData) HasOperator(ChairId, OperateCode int) bool {
	if OperateCode == WIK_NULL {
		return false
	}
	if room.UserAction[ChairId] == WIK_NULL {
		return false
	}

	if (room.UserAction[ChairId] & OperateCode) == 0 {
		return false
	}

	return true
}

//手中是否存在某张牌
func (room *RoomData) HasCard(ChairId, cardIdx int) bool {
	if cardIdx > room.GetCfg().MaxIdx {
		return false
	}
	return room.CardIndex[ChairId][cardIdx] > 0
}

//设置用户相应牌的操作 ,返回是否可以操作
func (room *RoomData) CheckUserOperator(u *user.User, userCnt, OperateCode int, OperateCard []int) (int, int) {
	if room.IsResponse[u.ChairId] {
		return -1, u.ChairId
	}
	room.IsResponse[u.ChairId] = true
	room.PerformAction[u.ChairId] = OperateCode
	room.OperateCard[u.ChairId] = OperateCard

	u.UserLimit = 0
	//放弃操作
	if OperateCode == WIK_NULL {
		////禁止这轮吃胡
		if room.HasOperator(u.ChairId, WIK_CHI_HU) {
			u.UserLimit |= LimitChiHu
		}
	}

	cbTargetAction := OperateCode
	wTargetUser := u.ChairId
	//执行判断
	for i := 0; i < userCnt; i++ {
		//获取动作
		cbUserAction := room.UserAction[i]
		if room.IsResponse[wTargetUser] {
			cbUserAction = room.PerformAction[i]
		}

		//优先级别
		cbUserActionRank := room.MjBase.LogicMgr.GetUserActionRank(cbUserAction)
		cbTargetActionRank := room.MjBase.LogicMgr.GetUserActionRank(cbTargetAction)

		//动作判断
		if cbUserActionRank > cbTargetActionRank {
			wTargetUser = i
			cbTargetAction = cbUserAction
		}
	}

	if room.IsResponse[wTargetUser] == false { //最高权限的人没响应
		return -1, u.ChairId
	}

	if cbTargetAction == WIK_NULL {
		room.UserAction = make([]int, userCnt)
		room.OperateCard = make([][]int, userCnt)
		room.PerformAction = make([]int, userCnt)
		return cbTargetAction, wTargetUser
	}

	//走到这里一定是所有人都响应完了
	return cbTargetAction, wTargetUser
}

func (room *RoomData) UserChiHu(wTargetUser, userCnt int) {
	//结束信息
	wChiHuUser := room.BankerUser
	for i := 0; i < userCnt; i++ {
		wChiHuUser = (room.BankerUser + i) % userCnt
		//过虑判断
		if (room.PerformAction[wChiHuUser] & WIK_CHI_HU) == 0 { //一跑多响
			continue
		}

		//胡牌判断
		pWeaveItem := room.WeaveItemArray[wChiHuUser]
		chihuKind, _ := room.MjBase.LogicMgr.AnalyseChiHuCard(room.CardIndex[wChiHuUser], pWeaveItem, room.OperateCard[wTargetUser][0], room.ChiHuRight[wChiHuUser], room.GetCfg().MaxCount, false)
		room.ChiHuKind[wChiHuUser] = chihuKind
		//插入扑克
		if room.ChiHuKind[wChiHuUser] != WIK_NULL {
			wTargetUser = wChiHuUser
		}
	}
}

//组合要操作的牌
func (room *RoomData) WeaveCard(cbTargetAction, wTargetUser int) {
	//变量定义
	cbTargetCard := room.OperateCard[wTargetUser][0]

	//出牌变量
	room.SendStatus = Gang_Send
	room.SendCardData = 0
	room.OutCardUser = INVALID_CHAIR
	room.OutCardData = 0

	//组合扑克
	Wrave := &msg.WeaveItem{}
	Wrave.Param = WIK_GANERAL
	Wrave.CenterCard = cbTargetCard
	Wrave.WeaveKind = cbTargetAction
	if room.ProvideUser == INVALID_CHAIR {
		Wrave.ProvideUser = wTargetUser
	} else {
		Wrave.ProvideUser = room.ProvideUser
	}

	Wrave.CardData[0] = cbTargetCard
	if cbTargetAction&(WIK_LEFT|WIK_CENTER|WIK_RIGHT) != 0 {
		Wrave.CardData[1] = room.OperateCard[wTargetUser][1]
		Wrave.CardData[2] = room.OperateCard[wTargetUser][2]
	} else {
		Wrave.CardData[1] = cbTargetCard
		Wrave.CardData[2] = cbTargetCard
		if cbTargetAction&WIK_GANG != 0 {
			Wrave.Param = WIK_FANG_GANG
			Wrave.CardData[3] = cbTargetCard
		}
	}
}

func (room *RoomData) RemoveCardByOP(wTargetUser, ChoOp int) bool {
	opCard := room.OperateCard[wTargetUser]
	var card []int
	switch ChoOp {
	case WIK_LEFT, WIK_RIGHT, WIK_CENTER:
		card = opCard[1:]
	case WIK_PENG:
		card = []int{opCard[0], opCard[0]}
	case WIK_GANG: //杠牌操作
		card = []int{opCard[0], opCard[0], opCard[0]}
	default:
		return false
	}
	//删除扑克
	if !room.MjBase.LogicMgr.RemoveCardByArr(room.CardIndex[wTargetUser], card) {
		log.Error("not foud card at RemoveCardByOP")
		return false
	}
	room.ChiPengCount[wTargetUser]++
	return true
}

func (room *RoomData) AnGang(u *user.User, cbOperateCode int, cbOperateCard []int) int {
	room.SendStatus = Gang_Send
	//变量定义
	var cbWeave *msg.WeaveItem
	cbCardIndex := room.MjBase.LogicMgr.SwitchToCardIndex(cbOperateCard[0])
	wProvideUser := u.ChairId
	cbGangKind := WIK_MING_GANG
	//杠牌处理
	if room.CardIndex[u.ChairId][cbCardIndex] == 1 {
		//寻找组合
		for _, v := range room.WeaveItemArray[u.ChairId] {
			if (v.CenterCard == cbOperateCard[0]) && (v.WeaveKind == WIK_PENG) {
				cbWeave = v
				break
			}
		}

		//没找到明杠
		if cbWeave == nil {
			return 0
		}
		cbGangKind = WIK_MING_GANG

		//组合扑克
		cbWeave.Param = WIK_MING_GANG
		cbWeave.WeaveKind = cbOperateCode
		cbWeave.CenterCard = cbOperateCard[0]
		cbWeave.CardData[3] = cbOperateCard[0]

		//杠牌得分
		wProvideUser = cbWeave.ProvideUser
	} else {
		//扑克效验

		if room.CardIndex[u.ChairId][cbCardIndex] != 4 {
			return 0
		}

		Wrave := &msg.WeaveItem{}
		Wrave.Param = WIK_AN_GANG
		Wrave.ProvideUser = u.ChairId
		Wrave.WeaveKind = cbOperateCode
		Wrave.CenterCard = cbOperateCard[0]
		for j := 0; j < 4; j++ {
			Wrave.CardData[j] = cbOperateCard[0]
		}
		room.WeaveItemArray[u.ChairId] = append(room.WeaveItemArray[u.ChairId], Wrave)
	}

	//删除扑克
	room.CardIndex[u.ChairId][cbCardIndex] = 0
	room.GangStatus = cbGangKind
	room.ProvideGangUser = wProvideUser
	room.GangCard[u.ChairId] = true
	room.GangCount[u.ChairId]++

	//构造结果
	OperateResult := &mj_hz_msg.G2C_HZMJ_OperateResult{}
	OperateResult.OperateUser = u.ChairId
	OperateResult.ProvideUser = wProvideUser
	OperateResult.OperateCode = cbOperateCode
	OperateResult.OperateCard[0] = cbOperateCard[0]

	//发送消息
	room.MjBase.UserMgr.SendMsgAll(OperateResult)

	return cbGangKind
}

func (room *RoomData) ZiMo(u *user.User) {
	//普通胡牌
	pWeaveItem := room.WeaveItemArray[u.ChairId]
	if !room.MjBase.LogicMgr.RemoveCard(room.CardIndex[u.ChairId], room.SendCardData) {
		log.Error("not foud card at Operater")
		return
	}
	kind, _ := room.MjBase.LogicMgr.AnalyseChiHuCard(room.CardIndex[u.ChairId], pWeaveItem, room.SendCardData, room.ChiHuRight[u.ChairId], room.GetCfg().MaxCount, false)
	room.ChiHuKind[u.ChairId] = int(kind)
	//结束信息

	room.ProvideCard = room.SendCardData
	return
}

func (room *RoomData) CallOperateResult(wTargetUser, cbTargetAction int) {
	//构造结果
	OperateResult := &mj_hz_msg.G2C_HZMJ_OperateResult{}
	OperateResult.OperateUser = wTargetUser
	OperateResult.OperateCode = cbTargetAction
	if room.ProvideUser == INVALID_CHAIR {
		OperateResult.ProvideUser = wTargetUser
	} else {
		OperateResult.ProvideUser = room.ProvideUser
	}

	cbTargetCard := room.OperateCard[wTargetUser][0]
	OperateResult.OperateCard[0] = cbTargetCard
	if cbTargetAction&(WIK_LEFT|WIK_CENTER|WIK_RIGHT) != 0 {
		OperateResult.OperateCard[1] = room.OperateCard[wTargetUser][1]
	} else if cbTargetAction&WIK_PENG != 0 {
		OperateResult.OperateCard[1] = cbTargetCard
		OperateResult.OperateCard[2] = cbTargetCard
	}

	//用户状态
	UserCnt := room.MjBase.UserMgr.GetMaxPlayerCnt()
	room.IsResponse = make([]bool, UserCnt)
	room.UserAction = make([]int, UserCnt)
	room.OperateCard = make([][]int, UserCnt)

	//如果非杠牌
	if cbTargetAction != WIK_GANG {
		room.ProvideUser = INVALID_CHAIR
		room.ProvideCard = 0

		gcr := &TagGangCardResult{}
		room.UserAction[wTargetUser] |= room.MjBase.LogicMgr.AnalyseGangCard(room.CardIndex[wTargetUser], room.WeaveItemArray[wTargetUser], 0, gcr)

		if room.Ting[wTargetUser] == false {
			HuData := &msg.G2C_Hu_Data{OutCardData: make([]int, room.GetCfg().MaxCount), HuCardCount: make([]int, room.GetCfg().MaxCount), HuCardData: make([][]int, room.GetCfg().MaxCount), HuCardRemainingCount: make([][]int, room.GetCfg().MaxCount)}
			cbCount := room.MjBase.LogicMgr.AnalyseTingCard(room.CardIndex[wTargetUser], room.WeaveItemArray[wTargetUser], HuData.OutCardData, HuData.HuCardCount, HuData.HuCardData, room.GetCfg().MaxCount)
			HuData.OutCardCount = cbCount
			if cbCount > 0 {
				room.UserAction[wTargetUser] |= WIK_LISTEN
				for i := 0; i < room.GetCfg().MaxCount; i++ {
					if HuData.HuCardCount[i] > 0 {
						for j := 0; j < HuData.HuCardCount[i]; j++ {
							HuData.HuCardRemainingCount[i][j] = room.GetRemainingCount(wTargetUser, HuData.HuCardData[i][j])
						}
					} else {
						break
					}
				}
				u := room.MjBase.UserMgr.GetUserByChairId(wTargetUser)
				u.WriteMsg(HuData)
			}
		}
		OperateResult.ActionMask |= room.UserAction[wTargetUser]
	}

	//发送消息
	room.MjBase.UserMgr.SendMsgAll(OperateResult)

	//设置用户
	room.CurrentUser = wTargetUser

	//杠牌处理
	if cbTargetAction == WIK_GANG {
		room.GangStatus = WIK_FANG_GANG
		if room.ProvideUser == INVALID_CHAIR {
			room.ProvideGangUser = wTargetUser
		} else {
			room.ProvideGangUser = room.ProvideUser
		}
		room.GangCard[wTargetUser] = true
		room.GangCount[wTargetUser]++

	}
	return
}

//响应判断
func (room *RoomData) EstimateUserRespond(wCenterUser int, cbCenterCard int, EstimatKind int) bool {
	log.Debug("at EstimateUserRespond ================= ")
	//变量定义
	bAroseAction := false

	//用户状态
	UserCnt := room.MjBase.UserMgr.GetMaxPlayerCnt()
	room.UserAction = make([]int, UserCnt)

	//动作判断
	room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
		//用户过滤
		if wCenterUser == u.ChairId {
			log.Debug("at EstimateUserRespond ======== wCenterUser:%v", wCenterUser)
			return
		}

		//托管了不响应
		if room.MjBase.UserMgr.IsTrustee(u.ChairId) {
			log.Debug("at EstimateUserRespond ======== IsTrustee ChairId:%v", u.ChairId)
			return
		}

		//出牌类型
		if EstimatKind == EstimatKind_OutCard {
			//吃碰判断
			log.Debug(".UserLimit&LimitPen %v, %v", u.UserLimit, u.UserLimit&LimitPeng)
			if u.UserLimit&LimitPeng == 0 {
				//碰牌判断
				room.UserAction[u.ChairId] |= room.MjBase.LogicMgr.EstimatePengCard(room.CardIndex[u.ChairId], cbCenterCard)
			}

			//杠牌判断
			log.Debug(".room.LeftCardCount > room.EndLeftCount %v, %v", room.LeftCardCount > room.EndLeftCount, u.UserLimit&LimitGang)
			if room.LeftCardCount > room.EndLeftCount && u.UserLimit&LimitGang == 0 {
				room.UserAction[u.ChairId] |= room.MjBase.LogicMgr.EstimateGangCard(room.CardIndex[u.ChairId], cbCenterCard)
			}
		}

		//检查抢杠胡
		if EstimatKind == EstimatKind_GangCard {
			//只有庄家和闲家之间才能放炮
			MogicCard := room.MjBase.LogicMgr.SwitchToCardData(room.MjBase.LogicMgr.GetMagicIndex())
			if room.MjBase.LogicMgr.GetMagicIndex() == room.GetCfg().MaxIdx || (room.MjBase.LogicMgr.GetMagicIndex() != room.GetCfg().MaxIdx && cbCenterCard != MogicCard) {
				if u.UserLimit|LimitChiHu == 0 {
					//吃胡判断
					chr := 0
					huKind, _ := room.MjBase.LogicMgr.AnalyseChiHuCard(room.CardIndex[u.ChairId], room.WeaveItemArray[u.ChairId], cbCenterCard, chr, room.GetCfg().MaxCount, false)
					room.UserAction[u.ChairId] |= huKind
				}
			}
		}

		//结果判断
		if room.UserAction[u.ChairId] != WIK_NULL {
			bAroseAction = true
		}
	})

	log.Debug("AaaaaaaaaaAAAAAAAAAAAAAAAA : %v", bAroseAction)
	//结果处理
	if bAroseAction {
		//设置变量
		room.ProvideUser = wCenterUser
		room.ProvideCard = cbCenterCard
		room.ResumeUser = room.CurrentUser
		room.CurrentUser = INVALID_CHAIR

		//发送提示
		room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
			if room.UserAction[u.ChairId] != WIK_NULL {
				u.WriteMsg(&mj_hz_msg.G2C_HZMJ_OperateNotify{
					ActionMask: room.UserAction[u.ChairId],
					ActionCard: room.ProvideCard,
				})
			}
		})
		return true
	}

	if room.GangStatus != WIK_GANERAL {
		room.GangOutCard = true
		room.GangStatus = WIK_GANERAL
		room.ProvideGangUser = INVALID_CHAIR
	} else {
		room.GangOutCard = false
	}

	return false
}

//派发扑克
func (room *RoomData) DispatchCardData(wCurrentUser int, bTail bool) int {
	//状态效验
	if room.SendStatus == Not_Send {
		log.Error("at DispatchCardData f room.SendStatus == Not_Send")
		return -1
	}

	//丢弃扑克
	if (room.OutCardUser != INVALID_CHAIR) && (room.OutCardData != 0) {
		if len(room.DiscardCard[room.OutCardUser]) < 1 {
			room.DiscardCard[room.OutCardUser] = make([]int, 60)
		}

		room.DiscardCard[room.OutCardUser] = append(room.DiscardCard[room.OutCardUser], room.OutCardData)
	}

	//荒庄结束
	if room.LeftCardCount <= room.EndLeftCount {
		room.ProvideUser = INVALID_CHAIR
		return 1
	}

	//发送扑克
	room.ProvideCard = room.GetSendCard(bTail, room.MjBase.UserMgr.GetMaxPlayerCnt())
	room.SendCardData = room.ProvideCard
	room.LastCatchCardUser = wCurrentUser
	//清除禁止胡牌的牌

	u := room.MjBase.UserMgr.GetUserByChairId(wCurrentUser)
	if u == nil {
		log.Error("at DispatchCardData not foud user ")
	}

	//清除禁止胡牌的牌
	u.UserLimit &= ^LimitChiHu
	u.UserLimit &= ^LimitPeng
	u.UserLimit &= ^LimitGang

	//设置变量
	room.OutCardUser = INVALID_CHAIR
	room.OutCardData = 0
	room.CurrentUser = wCurrentUser
	room.ProvideUser = wCurrentUser
	room.GangOutCard = false

	if bTail { //从尾部取牌，说明玩家杠牌了,计算分数
		room.CallGangScore()
	}

	//加牌
	room.CardIndex[wCurrentUser][room.MjBase.LogicMgr.SwitchToCardIndex(room.ProvideCard)]++
	//room.UserCatchCardCount[wCurrentUser]++;

	if !room.MjBase.UserMgr.IsTrustee(wCurrentUser) {
		//胡牌判断
		chr := 0
		room.CardIndex[wCurrentUser][room.MjBase.LogicMgr.SwitchToCardIndex(room.SendCardData)]--
		log.Debug("befer %v ", room.UserAction[wCurrentUser])
		huKind, _ := room.MjBase.LogicMgr.AnalyseChiHuCard(room.CardIndex[wCurrentUser], room.WeaveItemArray[wCurrentUser],
			room.SendCardData, chr, room.GetCfg().MaxCount, false)
		room.UserAction[wCurrentUser] |= huKind
		log.Debug("afert %v ", room.UserAction[wCurrentUser])
		room.CardIndex[wCurrentUser][room.MjBase.LogicMgr.SwitchToCardIndex(room.SendCardData)]++

		//杠牌判断
		if (room.LeftCardCount > room.EndLeftCount) && !room.Ting[wCurrentUser] {
			GangCardResult := &TagGangCardResult{}
			room.UserAction[wCurrentUser] |= room.MjBase.LogicMgr.AnalyseGangCard(room.CardIndex[wCurrentUser], room.WeaveItemArray[wCurrentUser], room.ProvideCard, GangCardResult)
		}
	}

	log.Debug("aaaaaaaaa %v", room.WeaveItemArray[wCurrentUser])
	//听牌判断
	HuData := &msg.G2C_Hu_Data{OutCardData: make([]int, room.GetCfg().MaxCount), HuCardCount: make([]int, room.GetCfg().MaxCount), HuCardData: make([][]int, room.GetCfg().MaxCount), HuCardRemainingCount: make([][]int, room.GetCfg().MaxCount)}
	if room.Ting[wCurrentUser] == false {
		cbCount := room.MjBase.LogicMgr.AnalyseTingCard(room.CardIndex[wCurrentUser], room.WeaveItemArray[wCurrentUser], HuData.OutCardData, HuData.HuCardCount, HuData.HuCardData, room.GetCfg().MaxCount)
		HuData.OutCardCount = int(cbCount)
		if cbCount > 0 {
			room.UserAction[wCurrentUser] |= WIK_LISTEN

			for i := 0; i < room.GetCfg().MaxCount; i++ {
				if HuData.HuCardCount[i] > 0 {
					for j := 0; j < HuData.HuCardCount[i]; j++ {
						HuData.HuCardRemainingCount[i] = append(HuData.HuCardRemainingCount[i], room.GetRemainingCount(wCurrentUser, HuData.HuCardData[i][j]))
					}
				} else {
					break
				}
			}

			u.WriteMsg(HuData)
		}
	}

	log.Debug("User Action === %v , %d", room.UserAction, room.UserAction[wCurrentUser])
	//构造数据
	SendCard := &mj_hz_msg.G2C_HZMJ_SendCard{}
	SendCard.SendCardUser = wCurrentUser
	SendCard.CurrentUser = wCurrentUser
	SendCard.Tail = bTail
	SendCard.ActionMask = room.UserAction[wCurrentUser]
	SendCard.CardData = room.ProvideCard
	//发送数据
	u.WriteMsg(SendCard)
	SendCard.CardData = 0
	room.MjBase.UserMgr.SendMsgAllNoSelf(u.Id, SendCard)

	room.UserActionDone = false
	if room.MjBase.UserMgr.IsTrustee(wCurrentUser) {
		room.UserActionDone = true
	}
	return 0
}

func (room *RoomData) BeforeStartGame(UserCnt int) {
	room.InitRoom(UserCnt)
}

func (room *RoomData) StartGameing() {
	room.StartDispatchCard()
}

func (room *RoomData) AfterStartGame() {
	//检查自摸
	room.CheckZiMo()
	//通知客户端开始了
	room.SendGameStart()
}

func (room *RoomData) InitRoom(UserCnt int) {
	//初始化
	log.Debug("初始化房间")
	room.RepertoryCard = make([]int, room.GetCfg().MaxRepertory)
	room.CardIndex = make([][]int, UserCnt)
	for i := 0; i < UserCnt; i++ {
		room.CardIndex[i] = make([]int, room.GetCfg().MaxIdx)
	}
	room.FlowerCnt = make([]int, UserCnt)
	room.ChiHuKind = make([]int, UserCnt)
	room.ChiPengCount = make([]int, UserCnt)
	room.GangCard = make([]bool, UserCnt) //杠牌状态
	room.GangCount = make([]int, UserCnt)
	room.Ting = make([]bool, UserCnt)
	room.UserAction = make([]int, UserCnt)
	room.DiscardCard = make([][]int, UserCnt)
	room.UserGangScore = make([]int, UserCnt)
	room.WeaveItemArray = make([][]*msg.WeaveItem, UserCnt)
	room.ChiHuRight = make([]int, UserCnt)
	room.HeapCardInfo = make([][]int, UserCnt)
	for i := 0; i < UserCnt; i++ {
		room.HeapCardInfo[i] = make([]int, 2)
	}

	room.LeftCardCount = room.GetCfg().MaxRepertory
	room.UserActionDone = false
	room.SendStatus = Not_Send
	room.GangStatus = WIK_GANERAL
	room.ProvideGangUser = INVALID_CHAIR
	room.HistoryScores = make([]*HistoryScore, UserCnt)
	room.MinusLastCount = 0
	room.MinusHeadCount = 0
}

func (room *RoomData) GetSice() (int, int) {
	Sice1 := util.RandInterval(1, 7)
	Sice2 := util.RandInterval(1, 7)
	minSice := int(math.Min(float64(Sice1), float64(Sice2)))
	return Sice2<<8 | Sice1, minSice
}

func (room *RoomData) StartDispatchCard() {
	log.Debug("begin start game hzmj")
	userMgr := room.MjBase.UserMgr
	gameLogic := room.MjBase.LogicMgr

	userMgr.ForEachUser(func(u *user.User) {
		userMgr.SetUsetStatus(u, US_PLAYING)
	})

	var minSice int
	UserCnt := userMgr.GetMaxPlayerCnt()
	room.SiceCount, minSice = room.GetSice()

	gameLogic.RandCardList(room.RepertoryCard, GetCardByIdx(room.ConfigIdx))

	//红中可以当财神
	gameLogic.SetMagicIndex(gameLogic.SwitchToCardIndex(0x35))

	//分发扑克
	userMgr.ForEachUser(func(u *user.User) {
		for i := 0; i < room.GetCfg().MaxCount-1; i++ {
			room.LeftCardCount--
			room.MinusHeadCount++
			cardIdx := gameLogic.SwitchToCardIndex(room.RepertoryCard[room.LeftCardCount])
			room.CardIndex[u.ChairId][cardIdx]++
		}
	})

	OwnerUser, _ := userMgr.GetUserByUid(room.CreateUser)
	if room.BankerUser == INVALID_CHAIR && (room.MjBase.Temp.ServerType&GAME_GENRE_PERSONAL) != 0 { //房卡模式下先把庄家给房主
		if OwnerUser != nil {
			room.BankerUser = OwnerUser.ChairId
		} else {
			log.Error("get bamkerUser error at StartGame")
		}
	}

	if room.BankerUser == INVALID_CHAIR {
		room.BankerUser = util.RandInterval(0, UserCnt-1)
	}

	if room.BankerUser >= UserCnt {
		log.Error(" room.BankerUser >=UserCnt %d,  %d", room.BankerUser, UserCnt)
	}

	room.MinusHeadCount++
	log.Debug("aaaaaaaaaa %d", room.LeftCardCount)
	room.SendCardData = room.RepertoryCard[room.LeftCardCount]
	room.LeftCardCount--

	//替换测试代码
	if conf.Test {
		for _, v := range base.GameTestpaiCache.All() {
			if v.KindID == room.MjBase.Temp.KindID && v.ServerID == room.MjBase.Temp.ServerID && v.IsAcivate == 1 {
				cards := utils.GetStrIntList(v.Cards, ",")
				room.SetUserCard(v.ChairId, cards)
			}
		}
	}
	room.CardIndex[room.BankerUser][gameLogic.SwitchToCardIndex(room.SendCardData)]++
	room.ProvideCard = room.SendCardData
	room.ProvideUser = room.BankerUser
	room.CurrentUser = room.BankerUser

	//堆立信息
	SiceCount := LOBYTE(room.SiceCount) + HIBYTE(room.SiceCount)
	TakeChairID := (room.BankerUser + SiceCount - 1) % UserCnt
	TakeCount := room.GetCfg().MaxRepertory - room.LeftCardCount
	for i := 0; i < UserCnt; i++ {
		//计算数目
		var ValidCount int
		if i == 0 {
			ValidCount = HEAP_FULL_COUNT - room.HeapCardInfo[TakeChairID][1] - (minSice)*2
		} else {
			ValidCount = HEAP_FULL_COUNT - room.HeapCardInfo[TakeChairID][1]
		}

		RemoveCount := int(math.Min(float64(ValidCount), float64(TakeCount)))

		//提取扑克
		TakeCount -= RemoveCount
		if i == 0 {
			room.HeapCardInfo[TakeChairID][1] += RemoveCount
		} else {
			room.HeapCardInfo[TakeChairID][0] += RemoveCount
		}

		//完成判断
		if TakeCount == 0 {
			room.HeapHead = TakeChairID
			room.HeapTail = (room.BankerUser + SiceCount - 1) % UserCnt
			break
		}
		//切换索引
		TakeChairID = (TakeChairID + UserCnt - 1) % UserCnt
	}

	room.UserAction = make([]int, UserCnt)

	gangCardResult := &TagGangCardResult{}
	room.UserAction[room.BankerUser] |= gameLogic.AnalyseGangCard(room.CardIndex[room.BankerUser], nil, 0, gangCardResult)

	//胡牌判断
	chr := 0
	room.CardIndex[room.BankerUser][gameLogic.SwitchToCardIndex(room.SendCardData)]--
	huKind, _ := gameLogic.AnalyseChiHuCard(room.CardIndex[room.BankerUser], []*msg.WeaveItem{}, room.SendCardData, chr, room.GetCfg().MaxCount, true)
	room.UserAction[room.BankerUser] |= huKind
	room.CardIndex[room.BankerUser][gameLogic.SwitchToCardIndex(room.SendCardData)]++

	return
}

func (room *RoomData) CheckZiMo() {
	//听牌判断
	Count := 0
	OwnerUser, _ := room.MjBase.UserMgr.GetUserByUid(room.CreateUser)
	HuData := &msg.G2C_Hu_Data{OutCardData: make([]int, room.GetCfg().MaxCount), HuCardCount: make([]int, room.GetCfg().MaxCount), HuCardData: make([][]int, room.GetCfg().MaxCount), HuCardRemainingCount: make([][]int, room.GetCfg().MaxCount)}
	if room.Ting[room.BankerUser] == false {
		Count = room.MjBase.LogicMgr.AnalyseTingCard(room.CardIndex[room.BankerUser], []*msg.WeaveItem{}, HuData.OutCardData, HuData.HuCardCount, HuData.HuCardData, room.GetCfg().MaxCount)
		HuData.OutCardCount = Count
		if Count > 0 {
			room.UserAction[room.BankerUser] |= WIK_LISTEN
			for i := 0; i < room.GetCfg().MaxCount; i++ {
				if HuData.HuCardCount[i] > 0 {
					for j := 0; j < HuData.HuCardCount[i]; j++ {
						HuData.HuCardRemainingCount[i] = append(HuData.HuCardRemainingCount[i], room.GetRemainingCount(room.BankerUser, HuData.HuCardData[i][j]))
					}
				} else {
					break
				}
			}
			OwnerUser.WriteMsg(HuData)
		}
	}
}

func (room *RoomData) SendGameStart() {

	//构造变量
	GameStart := &mj_hz_msg.G2C_HZMG_GameStart{}
	GameStart.BankerUser = room.BankerUser
	GameStart.SiceCount = room.SiceCount
	GameStart.HeapHead = room.HeapHead
	GameStart.HeapTail = room.HeapTail
	GameStart.MagicIndex = room.MjBase.LogicMgr.GetMagicIndex()
	GameStart.HeapCardInfo = room.HeapCardInfo
	//发送数据
	room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
		GameStart.UserAction = room.UserAction[u.ChairId]
		GameStart.CardData = room.MjBase.LogicMgr.GetUserCards(room.CardIndex[u.ChairId])
		u.WriteMsg(GameStart)
	})

}

//正常结束房间
func (room *RoomData) NormalEnd() {
	//变量定义
	UserCnt := room.MjBase.UserMgr.GetMaxPlayerCnt()
	GameConclude := &mj_hz_msg.G2C_HZMJ_GameConclude{}
	GameConclude.ChiHuKind = make([]int, UserCnt)
	GameConclude.CardCount = make([]int, UserCnt)
	GameConclude.HandCardData = make([][]int, UserCnt)
	GameConclude.GameScore = make([]int, UserCnt)
	GameConclude.GangScore = make([]int, UserCnt)
	GameConclude.Revenue = make([]int, UserCnt)
	GameConclude.ChiHuRight = make([]int, UserCnt)
	GameConclude.MaCount = make([]int, UserCnt)
	GameConclude.MaData = make([]int, UserCnt)

	for i := range GameConclude.HandCardData {
		GameConclude.HandCardData[i] = make([]int, room.GetCfg().MaxCount)
	}

	GameConclude.SendCardData = room.SendCardData
	GameConclude.LeftUser = INVALID_CHAIR
	room.ChiHuKind = make([]int, UserCnt)
	//结束信息
	for i := 0; i < UserCnt; i++ {
		GameConclude.ChiHuKind[i] = room.ChiHuKind[i]
		//权位过滤
		if room.ChiHuKind[i] == WIK_CHI_HU {
			room.FiltrateRight(i, &room.ChiHuRight[i])
			GameConclude.ChiHuRight[i] = room.ChiHuRight[i]
		}
		GameConclude.HandCardData[i] = room.MjBase.LogicMgr.GetUserCards(room.CardIndex[i])
		GameConclude.CardCount[i] = len(GameConclude.HandCardData[i])
	}

	//计算胡牌输赢分
	UserGameScore := make([]int, UserCnt)
	room.CalHuPaiScore(UserGameScore)

	//拷贝码数据
	GameConclude.MaCount = make([]int, 0)

	nCount := 0
	if nCount > 1 {
		nCount++
	}

	for i := 0; i < nCount; i++ {
		GameConclude.MaData[i] = room.RepertoryCard[room.MinusLastCount+i]
	}

	//积分变量
	ScoreInfoArray := make([]*msg.TagScoreInfo, UserCnt)

	GameConclude.ProvideUser = room.ProvideUser
	GameConclude.ProvideCard = room.ProvideCard

	//统计积分
	room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
		if u.Status != US_PLAYING {
			return
		}
		GameConclude.GameScore[u.ChairId] = UserGameScore[u.ChairId]
		//胡牌分算完后再加上杠的输赢分就是玩家本轮最终输赢分
		GameConclude.GameScore[u.ChairId] += room.UserGangScore[u.ChairId]
		GameConclude.GangScore[u.ChairId] = room.UserGangScore[u.ChairId]

		//收税
		if GameConclude.GameScore[u.ChairId] > 0 && (room.MjBase.Temp.ServerType&GAME_GENRE_GOLD) != 0 {
			GameConclude.Revenue[u.ChairId] = room.CalculateRevenue(u.ChairId, GameConclude.GameScore[u.ChairId])
			GameConclude.GameScore[u.ChairId] -= GameConclude.Revenue[u.ChairId]
		}

		ScoreInfoArray[u.ChairId] = &msg.TagScoreInfo{}
		ScoreInfoArray[u.ChairId].Revenue = GameConclude.Revenue[u.ChairId]
		ScoreInfoArray[u.ChairId].Score = GameConclude.GameScore[u.ChairId]
		if ScoreInfoArray[u.ChairId].Score > 0 {
			ScoreInfoArray[u.ChairId].Type = SCORE_TYPE_WIN
		} else {
			ScoreInfoArray[u.ChairId].Type = SCORE_TYPE_LOSE
		}

		//历史积分
		if room.HistoryScores[u.ChairId] == nil {
			room.HistoryScores[u.ChairId] = &HistoryScore{}
		}
		room.HistoryScores[u.ChairId].TurnScore = GameConclude.GameScore[u.ChairId]
		room.HistoryScores[u.ChairId].CollectScore += GameConclude.GameScore[u.ChairId]

	})

	//发送数据
	room.MjBase.UserMgr.SendMsgAll(GameConclude)

	//写入积分 todo
	room.MjBase.UserMgr.WriteTableScore(ScoreInfoArray, room.MjBase.UserMgr.GetMaxPlayerCnt(), HZMJ_CHANGE_SOURCE)
}

//解散接触
func (room *RoomData) DismissEnd() {
	//变量定义
	UserCnt := room.MjBase.UserMgr.GetMaxPlayerCnt()
	GameConclude := &mj_hz_msg.G2C_HZMJ_GameConclude{}
	GameConclude.ChiHuKind = make([]int, UserCnt)
	GameConclude.CardCount = make([]int, UserCnt)
	GameConclude.HandCardData = make([][]int, UserCnt)
	GameConclude.GameScore = make([]int, UserCnt)
	GameConclude.GangScore = make([]int, UserCnt)
	GameConclude.Revenue = make([]int, UserCnt)
	GameConclude.ChiHuRight = make([]int, UserCnt)
	GameConclude.MaCount = make([]int, UserCnt)
	GameConclude.MaData = make([]int, UserCnt)
	for i, _ := range GameConclude.HandCardData {
		GameConclude.HandCardData[i] = make([]int, room.GetCfg().MaxCount)
	}

	room.BankerUser = INVALID_CHAIR

	GameConclude.SendCardData = room.SendCardData

	//用户扑克
	if len(room.CardIndex) > 0 { //没开始就结束情况下小于0
		for i := 0; i < UserCnt; i++ {
			if len(room.CardIndex[i]) > 0 {
				GameConclude.HandCardData[i] = room.MjBase.LogicMgr.GetUserCards(room.CardIndex[i])
				GameConclude.CardCount[i] = len(GameConclude.HandCardData[i])
			}
		}
	}

	//发送信息
	room.MjBase.UserMgr.SendMsgAll(GameConclude)
}

func (room *RoomData) GetRemainingCount(ChairId int, cbCardData int) int {
	cbIndex := room.MjBase.LogicMgr.SwitchToCardIndex(cbCardData)
	Count := 0
	for i := room.MinusLastCount; i < room.GetCfg().MaxRepertory-room.MinusHeadCount; i++ {
		if room.RepertoryCard[i] == cbCardData {
			Count++
		}
	}

	for id, cards := range room.CardIndex {
		if id == ChairId {
			continue
		}
		Count += cards[cbIndex]
	}

	return Count
}

//权位过滤
func (room *RoomData) FiltrateRight(wWinner int, chr *int) {
	//自摸
	if wWinner == room.ProvideUser {
		*chr |= CHR_ZI_MO
	} else if room.GangStatus == WIK_MING_GANG {
		*chr |= CHR_QIANG_GANG_HU
	} else {
		log.Error("AT FiltrateRight")
	}
	return
}

//算分
func (room *RoomData) CalHuPaiScore(EndScore []int) {
	CellScore := room.Source
	UserCnt := room.MjBase.UserMgr.GetMaxPlayerCnt()
	UserScore := make([]int, UserCnt) //玩家手上分
	room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
		if u.Status != US_PLAYING {
			return
		}
		UserScore[u.ChairId] = int(u.Score)
	})

	WinUser := make([]int, UserCnt)
	WinCount := 0

	for i := 0; i < UserCnt; i++ {
		if WIK_CHI_HU == room.ChiHuKind[(room.BankerUser+i)%UserCnt] {
			WinUser[WinCount] = (room.BankerUser + i) % UserCnt
			WinCount++
		}
	}

	if WinCount > 0 {
		//有人胡牌
		bZiMo := room.ProvideUser == WinUser[0]
		if bZiMo {
			for i := 0; i < UserCnt; i++ {

				if i != WinUser[0] {
					EndScore[i] -= CellScore
					EndScore[WinUser[0]] += CellScore
				}
			}
		} else {
			//抢杠
			for i := 0; i < WinCount; i++ {
				for j := 0; j < UserCnt; j++ {
					if j != WinUser[i] {
						EndScore[WinUser[i]] += CellScore
					}
				}
				EndScore[room.ProvideUser] -= EndScore[WinUser[i]]
			}
		}

		//谁胡谁当庄
		room.BankerUser = WinUser[0]
		if WinCount > 1 { //多个玩家胡牌，放炮者当庄
			room.BankerUser = room.ProvideUser
		}
	} else { //荒庄
		room.BankerUser = room.LastCatchCardUser //最后一个摸牌的人当庄
	}
}

//计算税收  暂时没有这个 功能
func (room *RoomData) CalculateRevenue(ChairId, lScore int) int {
	return 0
}

//取得扑克
func (room *RoomData) GetSendCard(bTail bool, UserCnt int) int {
	//发送扑克
	room.LeftCardCount--

	var cbSendCardData int
	var cbIndexCard int
	if bTail {
		cbSendCardData = room.RepertoryCard[room.MinusLastCount]
		room.MinusLastCount++
	} else {
		room.MinusHeadCount++
		cbIndexCard = room.GetCfg().MaxRepertory - room.MinusHeadCount
		cbSendCardData = room.RepertoryCard[cbIndexCard]
	}

	//堆立信息

	if !bTail {
		//切换索引
		cbHeapCount := room.HeapCardInfo[room.HeapHead][0] + room.HeapCardInfo[room.HeapHead][1]
		if cbHeapCount == HEAP_FULL_COUNT {
			room.HeapHead = (room.HeapHead + UserCnt - 1) % len(room.HeapCardInfo)
		}
		room.HeapCardInfo[room.HeapHead][0]++
	} else {
		//切换索引
		cbHeapCount := room.HeapCardInfo[room.HeapTail][0] + room.HeapCardInfo[room.HeapTail][1]
		if cbHeapCount == HEAP_FULL_COUNT {
			room.HeapTail = (room.HeapTail + 1) % len(room.HeapCardInfo)
		}
		room.HeapCardInfo[room.HeapTail][1]++
	}

	return cbSendCardData
}

func (room *RoomData) CallGangScore() {
	lcell := room.Source
	if room.GangStatus == WIK_FANG_GANG { //放杠一家扣分
		room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
			if u.Status != US_PLAYING {
				return
			}
			if u.ChairId != room.CurrentUser {
				room.UserGangScore[room.ProvideGangUser] -= lcell
				room.UserGangScore[room.CurrentUser] += lcell
			}
		})
	} else if room.GangStatus == WIK_MING_GANG { //明杠每家出1倍
		room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
			if u.Status != US_PLAYING {
				return
			}
			if u.ChairId != room.CurrentUser {
				room.UserGangScore[u.ChairId] -= lcell
				room.UserGangScore[room.CurrentUser] += lcell
			}
		})

		//记录明杠次数
	} else if room.GangStatus == WIK_AN_GANG { //暗杠每家出2倍
		room.MjBase.UserMgr.ForEachUser(func(u *user.User) {
			if u.Status != US_PLAYING {
				return
			}
			if u.ChairId != room.CurrentUser {
				room.UserGangScore[u.ChairId] -= 2 * lcell
				room.UserGangScore[room.CurrentUser] += 2 * lcell
			}
		})
	}
}

func (room *RoomData) IsActionDone() bool {
	return room.UserActionDone
}

func (room *RoomData) GetTrusteeOutCard(wChairID int) int {
	cardindex := INVALID_BYTE
	if room.SendCardData != 0 {
		cardindex = room.MjBase.LogicMgr.SwitchToCardIndex(room.SendCardData)
	} else {
		for i := 0; i < room.GetCfg().MaxIdx; i++ {
			if room.CardIndex[wChairID][i] > 0 {
				cardindex = i
				break
			}
		}
	}
	return cardindex
}

//插花
func (room *RoomData) GetChaHua(u *user.User, setCount int) {
}

//补花
func (room *RoomData) OnUserReplaceCard(u *user.User, CardData int) bool {
	gameLogic := room.MjBase.LogicMgr
	if gameLogic.RemoveCard(room.CardIndex[u.ChairId], CardData) == false {
		return false
	}

	//记录补花
	room.FlowerCnt[u.ChairId]++

	//是否花杠
	if room.FlowerCnt[u.ChairId] == 8 {
		room.MjBase.OnEventGameConclude(u.ChairId, u, GER_NORMAL)
	}

	//状态变量
	room.SendStatus = BuHua_Send
	room.GangStatus = WIK_GANERAL
	room.ProvideUser = INVALID_CHAIR

	//派发扑克
	room.DispatchCardData(u.ChairId, true)

	outData := &mj_zp_msg.G2C_MJZP_ReplaceCard{}
	outData.IsInitFlower = false
	outData.ReplaceUser = u.ChairId
	outData.ReplaceCard = CardData
	outData.NewCard = room.SendCardData
	room.MjBase.UserMgr.SendMsgAll(&outData)

	log.Debug("[用户补花] 用户：%d,花牌：%x 新牌：%x", u.ChairId, CardData, room.SendCardData)
	return true
}

//用户听牌
func (room *RoomData) OnUserListenCard(u *user.User, bListenCard bool) bool {
	return true
}

//记录分饼
func (room *RoomData) RecordFollowCard(cbCenterCard int) bool {
	return true
}

//////////////////////////////////////////////////
//胡牌算分类型

//地胡
func (room *RoomData) IsDiHu(pAnalyseItem *TagAnalyseItem) bool {
	if room.BankerUser == room.CurrentUser {
		return false
	}
	//吃碰杠失效
	for k, v := range room.CardIndex {
		if k == room.CurrentUser {
			if len(v) != room.GetCfg().MaxCount {
				return false
			}
		}
		if len(v) != room.GetCfg().MaxCount-1 {
			return false
		}
	}

	var sumFlowerCount int
	for _, v := range room.FlowerCnt {
		sumFlowerCount += v
	}

	sumUserCard := 4*room.GetCfg().MaxCount - 3
	if room.LeftCardCount != room.GetCfg().MaxRepertory-sumUserCard-sumFlowerCount-1 {
		return false
	}

	return true
}

//天胡
func (room *RoomData) IsTianHu(pAnalyseItem *TagAnalyseItem) bool {
	if room.BankerUser != room.CurrentUser {
		return false
	}

	//吃碰杠失效
	for k, v := range room.CardIndex {
		if k == room.CurrentUser {
			if len(v) != room.GetCfg().MaxCount {
				return false
			}
		}
		if len(v) != room.GetCfg().MaxCount-1 {
			return false
		}
	}

	var sumFlowerCount int
	for _, v := range room.FlowerCnt {
		sumFlowerCount += v
	}

	sumUserCard := 4*room.GetCfg().MaxCount - 3
	if room.LeftCardCount != room.GetCfg().MaxRepertory-sumUserCard-sumFlowerCount {
		return false
	}

	return true
}

//杠上开花
func (room *RoomData) IsGangKaiHua(pAnalyseItem *TagAnalyseItem) bool {

	if room.CurrentUser != room.ProvideUser {
		return false
	}

	if pAnalyseItem.WeaveKind[room.GetCfg().MaxWeave] == WIK_GANG && pAnalyseItem.IsAnalyseGet[room.GetCfg().MaxWeave] == false {
		return true
	}

	return true
}

//花上开花
func (room *RoomData) IsHuaKaiHua(pAnalyseItem *TagAnalyseItem) bool {
	if room.CurrentUser != room.ProvideUser {
		return false
	}

	if room.FlowerCnt[room.CurrentUser] == 0 {
		return false
	}

	if room.SendStatus == BuHua_Send {
		return true
	}

	return false
}

//海底捞针
func (room *RoomData) IsHaiDiLaoYue(pAnalyseItem *TagAnalyseItem) bool {
	if room.ProvideUser != room.CurrentUser {
		return false
	}
	if len(pAnalyseItem.WeaveKind) != 0 {
		return false
	}
	return true
}

//字牌刻字
func (room *RoomData) IsKeZi(pAnalyseItem *TagAnalyseItem) bool {

	for _, v := range pAnalyseItem.CenterCard {
		cardColor := v >> 4
		if cardColor == 3 {
			return true
		}
	}
	return false
}

//花杠
func (room *RoomData) IsHuaGang(pAnalyseItem *TagAnalyseItem) bool {
	if room.FlowerCnt[room.CurrentUser] == 8 {
		return true
	}
	return false
}

//暗刻
func (room *RoomData) IsSanAnKe(pAnalyseItem *TagAnalyseItem) int {
	var anKeCount int
	for k, v := range pAnalyseItem.WeaveKind {
		if v == WIK_GANG && pAnalyseItem.IsAnalyseGet[k] == true {
			anKeCount++
		}
	}
	if anKeCount <= 3 {
		return 0
	}
	return anKeCount
}

//无花字
func (room *RoomData) IsWuHuaZi(pAnalyseItem *TagAnalyseItem) bool {
	if room.FlowerCnt[room.CurrentUser] != 0 {
		return false
	}

	for _, v := range pAnalyseItem.CenterCard {
		cardColor := v >> 4
		if cardColor == 3 {
			return false
		}
	}
	return true
}

//对对胡
func (room *RoomData) IsDuiDuiHu(pAnalyseItem *TagAnalyseItem) bool {
	for _, v := range pAnalyseItem.WeaveKind {
		if v&(WIK_PENG|WIK_GANG) == 0 {
			return false
		}
	}
	return true
}

//小四喜
func (room *RoomData) IsXiaoSiXi(pAnalyseItem *TagAnalyseItem) bool {
	var colorCount [4]int
	for _, v := range pAnalyseItem.CenterCard {
		cardColor := v >> 4
		if cardColor == 3 {
			cardV := v & MASK_VALUE
			if cardV > 4 {
				return false
			}
			colorCount[cardV] = 1
		}
	}

	if colorCount[0]+colorCount[1]+colorCount[2]+colorCount[3] == 3 {
		for k, v := range colorCount {
			if v == 0 {
				if k == pAnalyseItem.CardEye&MASK_VALUE {
					return true
				}
			}
		}
	}
	return false
}

//大四喜
func (room *RoomData) IsDaSiXi(pAnalyseItem *TagAnalyseItem) bool {
	var colorCount [4]int
	for _, v := range pAnalyseItem.CenterCard {
		cardColor := v >> 4
		if cardColor == 3 {
			cardV := v & MASK_VALUE
			if cardV > 4 {
				return false
			}
			colorCount[cardV] = 1
		}
	}

	if colorCount[0]+colorCount[1]+colorCount[2]+colorCount[3] == 4 {
		return true
	}
	return false
}

//小三元
func (room *RoomData) IsXiaoSanYuan(pAnalyseItem *TagAnalyseItem) bool {
	var colorCount [3]int
	for _, v := range pAnalyseItem.CenterCard {
		cardColor := v >> 4
		if cardColor == 3 {
			cardV := (v & MASK_VALUE) - 5
			if cardV < 0 {
				return false
			}
			colorCount[cardV] = 1
		}
	}

	if colorCount[0]+colorCount[1]+colorCount[2] == 2 {
		for k, v := range colorCount {
			if v == 0 {
				if k == (pAnalyseItem.CardEye&MASK_VALUE)-5 {
					return true
				}
			}
		}
	}
	return false
}

//大三元
func (room *RoomData) IsDaSanYuan(pAnalyseItem *TagAnalyseItem) bool {
	var colorCount [3]int
	for _, v := range pAnalyseItem.CenterCard {
		cardColor := v >> 4
		if cardColor == 3 {
			cardV := (v & MASK_VALUE) - 5
			if cardV < 0 {
				return false
			}
			colorCount[cardV] = 1
		}
	}

	if colorCount[0]+colorCount[1]+colorCount[2] == 3 {
		return true
	}
	return false
}

//混一色
func (room *RoomData) IsHunYiSe(pAnalyseItem *TagAnalyseItem) bool {
	cardColor := pAnalyseItem.CardEye >> 4
	var colorCount [4]int
	colorCount[cardColor] = 1
	for _, v := range pAnalyseItem.CenterCard {
		cardColor = v >> 4
		if colorCount[cardColor] == 0 {
			colorCount[cardColor] = 1
		}
	}
	if colorCount[0]+colorCount[1]+colorCount[2] == 1 && colorCount[3] == 1 {
		return true
	}

	return false
}

//清一色
func (room *RoomData) IsQingYiSe(pAnalyseItem *TagAnalyseItem, bQuanFan *bool) bool {
	cardColor := pAnalyseItem.CardEye & MASK_COLOR
	for _, v := range pAnalyseItem.CenterCard {
		if v&MASK_COLOR != cardColor {
			return false
		}
	}

	if 0x30 == cardColor {
		*bQuanFan = true
	} else {
		*bQuanFan = false
	}
	return false
}

//花一色
func (room *RoomData) IsHuaYiSe(pAnalyseItem *TagAnalyseItem, bQuanFan *bool) bool {
	cardColor := pAnalyseItem.CardEye & MASK_COLOR
	for _, v := range pAnalyseItem.CenterCard {
		if v&MASK_COLOR != cardColor {
			return false
		}
	}

	if room.FlowerCnt[room.CurrentUser] > 0 {
		return true
	}

	return false
}

//门清
func (room *RoomData) IsMenQing(pAnalyseItem *TagAnalyseItem) bool {

	for k, v := range pAnalyseItem.WeaveKind {
		if (v&(WIK_LEFT|WIK_CENTER|WIK_RIGHT)) != 0 || v == WIK_PENG || v == WIK_GANG {
			if pAnalyseItem.IsAnalyseGet[k] == false {
				return false
			}
		}
	}

	return true
}

//佰六
func (room *RoomData) IsBaiLiu(pAnalyseItem *TagAnalyseItem) bool {

	if room.FlowerCnt[room.CurrentUser] > 0 {
		return false
	}

	for k, v := range pAnalyseItem.WeaveKind {
		if (v & (WIK_LEFT | WIK_RIGHT)) == 0 {
			if pAnalyseItem.IsAnalyseGet[k] == false {
				return false
			}
		}
	}

	return false
}

//门清佰六
func (room *RoomData) IsMenQingBaiLiu(pAnalyseItem *TagAnalyseItem) bool {
	if room.IsMenQing(pAnalyseItem) && room.IsBaiLiu(pAnalyseItem) {
		return true
	}
	return false
}

//胡尾张
func (room *RoomData) IsHuWeiZhang(pAnalyseItem *TagAnalyseItem, cbCurrentCard int) bool {
	logic := room.MjBase.LogicMgr

	if logic.GetCardCount(room.CardIndex[room.CurrentUser]) == 1 {
		return false
	}

	if pAnalyseItem.CardEye == cbCurrentCard {
		return true
	}
	return false
}

//截头
func (room *RoomData) IsJieTou(pAnalyseItem *TagAnalyseItem, cbCurrentCard int) bool {
	cardValue := cbCurrentCard & MASK_VALUE
	for k, v := range pAnalyseItem.WeaveKind {
		if v&(WIK_LEFT|WIK_CENTER|WIK_RIGHT) == 0 || !pAnalyseItem.IsAnalyseGet[k] {
			continue
		} else {
			//1-2 胡3
			if cardValue == 3 && cbCurrentCard == pAnalyseItem.CardData[k][2] {
				return true
			}
			//8-9 胡7
			if cardValue == 7 && cbCurrentCard == pAnalyseItem.CardData[k][0] {
				return true
			}
		}

	}
	return false
}

//空心
func (room *RoomData) IsKongXin(pAnalyseItem *TagAnalyseItem, cbCurrentCard int) bool {
	for k, v := range pAnalyseItem.WeaveKind {
		if v&(WIK_LEFT|WIK_CENTER|WIK_RIGHT) == 0 || !pAnalyseItem.IsAnalyseGet[k] {
			continue
		} else {
			if cbCurrentCard == pAnalyseItem.CardData[k][1] {
				return true
			}
		}
	}
	return false
}

//单吊
func (room *RoomData) IsDanDiao(pAnalyseItem *TagAnalyseItem) bool {

	//todo,单吊
	return false
}