package db

import (
	"strings"
	"time"
	"github.com/lovelly/leaf/log"
	"mj/hallServer/conf"
	"mj/common"
)

type loader interface {
	LoadAll()
}

var (
	BaseDataCaches = make(map[string]loader)
	NeedReloadBaseDB bool
)

const (
	refreshInterval = 10 * time.Second // 单位：秒
)

// 仅用于运行时刷新Base库
func RefreshInTime() {
	 go func() {
	 	ticker := time.NewTicker(refreshInterval)

	 	for {
			if NeedReloadBaseDB == false {
				continue
			}
	 		select {
	 		case <-ticker.C:
	 			refreshBase()
	 		}
	 	}
	 }()
}

func refreshBase() {

	log.Debug("refreshBase")
	tableListStr := ""
	cnt := 0
	row := BaseDB.QueryRowx("select refresh_table_list, cnt from  refresh_in_time where node_id=?;",conf.Server.NodeId)
	log.Debug("node_id : %d", conf.Server.NodeId)
	err := row.Scan(&tableListStr, &cnt)
	if err != nil {
		log.Error("Query refresh_table_list encounter a error.")
		return
	}

	if tableListStr == "" {
		return
	}

	tableList := strings.Split(tableListStr, ",")
	log.Debug("Some table is need to refresh.Names:%v, %D", tableList, cnt)

	for key := range BaseDataCaches {
		log.Debug("%s", key)
	}
	for _, tableName := range tableList {
		key, err := common.TranslatePascal(tableName)
		if err != nil {
			log.Error("TranslatePascal is failed.tableName: %v", tableName)
			continue
		}

		if l, find := BaseDataCaches[key]; find {
			l.LoadAll()
		} else {
			log.Error("tableName is not exists.tableName: %v", tableName)
		}
	}

	cnt--
	if cnt < 1 {
		BaseDB.Exec("update refresh_in_time set refresh_table_list = '', cnt = 0 where node_id =?;", conf.Server.NodeId)
	} else {
		BaseDB.Exec("update refresh_in_time set cnt = ? where node_id =?;", cnt, conf.Server.NodeId)
	}
}
