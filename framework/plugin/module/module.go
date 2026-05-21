package module

import (
	"context"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var moduleMgr = common.NewPluginMgr("module")

func Register(module any) {
	_ = moduleMgr.Register(module)
}

func Setup(ctx context.Context, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error {
	return moduleMgr.Setup(ctx, eventHub, backgroundRoutine)
}

func Run(ctx context.Context) *cd.Error {
	return moduleMgr.Run(ctx)
}

func Teardown(ctx context.Context) {
	moduleMgr.Teardown(ctx)
}
