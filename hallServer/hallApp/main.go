package main

import (
	"mj/common"
	"mj/common/consul"
	"github.com/lovelly/leaf"
	lconf "github.com/lovelly/leaf/conf"
	"mj/hallServer/conf"
	"mj/hallServer/gate"
	"mj/hallServer/center"
	"mj/hallServer/gameList"
	. "mj/common/cost"
	"mj/hallServer/db/model/base"
	"mj/hallServer/db"
	"mj/common/utils"
)

func main() {
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath
	lconf.ListenAddr = conf.Server.ListenAddr
	lconf.ServerName = conf.ServerName()
	lconf.ConnAddrs = conf.Server.ConnAddrs
	lconf.PendingWriteNum = conf.Server.PendingWriteNum
	lconf.HeartBeatInterval = conf.HeartBeatInterval

	common.Init()
	consul.SetConfig(&conf.ConsulConfig{})
	consul.SetSelfId(lconf.ServerName)
	consul.AddinitiativeSvr(GamePrefix)
	db.InitDB(&conf.DBConfig{})
	base.LoadBaseData()
	utils.CreatePrivateServer(conf.Server.PrivatePort)
	leaf.Run(
		gate.Module,
		center.Module,
		consul.Module,
		gameList.Module,
	)
}
