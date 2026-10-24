package controller

import (
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
	"github.com/mhsanaei/3x-ui/v3/internal/web/session"

	"github.com/gin-gonic/gin"
)

type ExitPoolController struct {
	service service.ExitPoolService
}

func NewExitPoolController(g *gin.RouterGroup) *ExitPoolController {
	a := &ExitPoolController{}
	g = g.Group("/exit-pool")
	g.GET("", a.state)
	g.POST("/sync", a.sync)
	g.POST("/delete/:subId", a.delete)
	return a
}

func (a *ExitPoolController) state(c *gin.Context) {
	state, err := a.service.State(session.GetLoginUser(c).Id)
	jsonObj(c, state, err)
}

func (a *ExitPoolController) sync(c *gin.Context) {
	var request service.ExitPoolSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		jsonObj(c, nil, err)
		return
	}
	group, err := a.service.Sync(session.GetLoginUser(c).Id, request)
	if err != nil {
		jsonObj(c, nil, err)
		return
	}
	websocketNotifyXrayAndClients()
	jsonObj(c, group, nil)
}

func (a *ExitPoolController) delete(c *gin.Context) {
	if err := a.service.Delete(c.Param("subId")); err != nil {
		jsonObj(c, nil, err)
		return
	}
	websocketNotifyXrayAndClients()
	jsonMsg(c, "Exit-pool subscription group deleted", nil)
}

func websocketNotifyXrayAndClients() {
	notifyClientsChanged()
}
