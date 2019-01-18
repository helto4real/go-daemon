package core

import (
	"context"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-hassclient/client"
)

type Entity struct {
	id                     string
	hassEntity             *client.HassEntity
	entityChan             chan d.DaemonEntity
	listenChan             chan client.HassEntity
	daemonHelper           d.DaemonAppHelper
	passEntityChanges      bool
	passCallServiceEvents  bool
	cancelContext          context.Context
	autoRespondServiceCall bool
}

func NewEntity(id string, daemonHelper d.DaemonAppHelper, autoRespondServiceCall bool,
	changedEntityChannel chan d.DaemonEntity) d.DaemonEntity {
	currentEntity, ok := daemonHelper.GetEntity(id)

	if !ok {
		// No existing in hass, create one
		currentEntity = client.NewHassEntity(id, id, client.HassEntityState{},
			client.HassEntityState{State: "unknown", Attributes: map[string]interface{}{}})
	}

	entity := Entity{hassEntity: currentEntity, entityChan: changedEntityChannel,
		listenChan: make(chan client.HassEntity, 2), passEntityChanges: false,
		passCallServiceEvents: false, cancelContext: daemonHelper.GetCancelContext(),
		autoRespondServiceCall: autoRespondServiceCall, id: id, daemonHelper: daemonHelper}

	entity.init()

	return &entity
}

func (a *Entity) init() {
	a.daemonHelper.ListenState(a.id, a.listenChan)
	go a.messagePump()
}

func (a *Entity) ID() string {
	return a.id
}

func (a *Entity) State() interface{} {
	return a.hassEntity.New.State
}

func (a *Entity) Attributes() map[string]interface{} {
	return a.hassEntity.New.Attributes
}

func (a *Entity) Entity() *client.HassEntity {
	return a.hassEntity
}

func (a *Entity) messagePump() {

	for {
		select {
		case entity, ok := <-a.listenChan:
			if !ok {
				return
			}
			a.hassEntity.New.State = entity.New.State
			for key, attribute := range entity.New.Attributes {
				a.hassEntity.New.Attributes[key] = attribute
			}
			a.hassEntity.Old = entity.Old

			if a.entityChan != nil {
				a.entityChan <- a
			}
		case <-a.cancelContext.Done():
			return
		}
	}
}
