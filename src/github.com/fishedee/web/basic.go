package web

import (
	"net/http"
	"os"
	"path"
	"testing"
)

type Basic struct {
	Ctx      Context
	Config   Config
	Security Security
	Session  Session
	DB       Database
	DB2      Database
	DB3      Database
	DB4      Database
	DB5      Database
	Log      Log
	Monitor  Monitor
	Timer    Timer
	Queue    Queue
	Cache    Cache
}

type basicInner struct {
	Config   *ConfigManager
	Security *SecurityManager
	Session  *SessionManager
	DB       *DatabaseManager
	DB2      *DatabaseManager
	DB3      *DatabaseManager
	DB4      *DatabaseManager
	DB5      *DatabaseManager
	Log      *LogManager
	Monitor  *MonitorManager
	Timer    *TimerManager
	Queue    *QueueManager
	Cache    *CacheManager
}

var globalBasic basicInner

func init() {
	var err error
	globalBasic.Config, err = NewConfigManager("conf/app.conf")
	if err != nil {
		panic(err)
	}
	globalBasic.Security, err = NewSecurityManagerFromConfig("fishsecurity")
	if err != nil {
		panic(err)
	}
	globalBasic.Session, err = NewSessionManagerFromConfig("fishsession")
	if err != nil {
		panic(err)
	}
	globalBasic.DB, err = NewDatabaseManagerFromConfig("fishdb")
	if err != nil {
		panic(err)
	}
	globalBasic.DB2, err = NewDatabaseManagerFromConfig("fishdb2")
	if err != nil {
		panic(err)
	}
	globalBasic.DB3, err = NewDatabaseManagerFromConfig("fishdb3")
	if err != nil {
		panic(err)
	}
	globalBasic.DB4, err = NewDatabaseManagerFromConfig("fishdb4")
	if err != nil {
		panic(err)
	}
	globalBasic.DB5, err = NewDatabaseManagerFromConfig("fishdb5")
	if err != nil {
		panic(err)
	}
	globalBasic.Log, err = NewLogManagerFromConfig("fishlog")
	if err != nil {
		panic(err)
	}
	globalBasic.Monitor, err = NewMonitorManagerFromConfig("fishmonitor")
	if err != nil {
		panic(err)
	}
	globalBasic.Timer, err = NewTimerManager()
	if err != nil {
		panic(err)
	}
	globalBasic.Queue, err = NewQueueManagerFromConfig("fishqueue")
	if err != nil {
		panic(err)
	}
	globalBasic.Cache, err = NewCacheManagerFromConfig("fishcache")
	if err != nil {
		panic(err)
	}
}
func initBasic(request *http.Request, response http.ResponseWriter, t *testing.T) *Basic {
	return &Basic{
		Ctx: Context{
			Request:  request,
			Response: response,
			Testing:  t,
		},
		Security: globalBasic.Security,
		DB:       globalBasic.DB,
		DB2:      globalBasic.DB2,
		DB3:      globalBasic.DB3,
		DB4:      globalBasic.DB4,
		DB5:      globalBasic.DB5,
		Monitor:  globalBasic.Monitor,
		Log:      NewLogManagerWithCtxAndMonitor(request, globalBasic.Monitor, globalBasic.Log),
		Timer:    NewTimerManagerWithLog(globalBasic.Log, globalBasic.Timer),
		Queue:    NewQueueManagerWithLog(globalBasic.Log, globalBasic.Queue),
		Cache:    NewCacheManagerWithLog(globalBasic.Log, globalBasic.Cache),
	}
}