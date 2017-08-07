package pk

import (
	"mj/common/utils"
	"mj/gameServer/db/model/base"
	"mj/gameServer/user"
	"strconv"
)

type DataManager interface {
	OnCreateRoom()
	InitRoom(UserCnt int)
	GetRoomId() int
	CanOperatorRoom(uid int64) bool

	// 游戏开始
	BeforeStartGame(UserCnt int)
	StartGameing()
	AfterStartGame()

	// 游戏结束
	NormalEnd(Reason int)
	DismissEnd(Reason int)

	SendPersonalTableTip(*user.User)
	SendStatusPlay(u *user.User)
	SendStatusReady(u *user.User)

	// 叫分 加注 亮牌
	CallScore(u *user.User, scoreTimes int)
	AddScore(u *user.User, score int)
	OpenCard(u *user.User, cardType int, cardData []int)

	// 明牌
	ShowCard(u *user.User)
	// 托管
	OtherOperation(args []interface{})
}

type LogicManager interface {
	RandCardList(cbCardBuffer, OriDataArray []int)
	SortCardList(cardData []int, cardCount int)
	GetCardValue(CardData int) int
	GetCardColor(CardData int) int

	CompareCard(firstCardData []int, lastCardData []int) bool
	GetCardType(cardData []int) int
	GetCardTimes(cardType int) int

	CompareCardWithParam(firstCardData []int, lastCardData []int, args []interface{}) (int, bool)
	// 以下接口不通用
	RemoveCardList(cbRemoveCard []int, cbCardData []int) ([]int, bool)
	SetParamToLogic(args interface{}) // 设置算法必要参数

}

////////////////////////////////////////////
//全局变量
// TODO 增加 默认(错误)值 参数
func getGlobalVar(key string) string {
	if globalVar, ok := base.GlobalVarCache.Get(key); ok {
		return globalVar.V
	}
	return ""
}

func getGlobalVarFloat64(key string) float64 {
	if value := getGlobalVar(key); value != "" {
		v, _ := strconv.ParseFloat(value, 10)
		return v
	}
	return 0
}

func getGlobalVarInt64(key string, val int64) int64 {
	if value := getGlobalVar(key); value != "" {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return v
		}
	}
	return val
}

func getGlobalVarInt(key string) int {
	if value := getGlobalVar(key); value != "" {
		v, _ := strconv.Atoi(value)
		return v
	}
	return 0
}

func getGlobalVarIntList(key string) []int {
	if value := getGlobalVar(key); value != "" {
		return utils.GetStrIntList(value)
	}
	return nil
}
