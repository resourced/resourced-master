package watcher_actions

import (
	"github.com/resourced/resourced-master/dal"
)

type IWatcherActions interface {
	SetSettings(map[string]interface{}) error
	SetWatcher(*dal.WatcherRow)
}
