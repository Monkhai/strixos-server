package game

import (
	"context"
	"sync"

	"github.com/Monkhai/strixos-server.git/pkg/utils"
)

func NewInviteGame(player *Player, parentCtx context.Context) *Game {
	ctx, cancel := context.WithCancel(parentCtx)
	player.SetIsInGame(true)
	return &Game{
		Board:   NewBoard(),
		Player1: player,
		MsgChan: make(chan interface{}, 10),
		Ctx:     ctx,
		Cancel:  cancel,
		ID:      utils.GenerateUniqueID(),
		Mux:     &sync.RWMutex{},
	}
}

func (g *Game) AddSecondPlayer(p *Player) {
	g.Mux.Lock()
	g.Player2 = p
	g.Mux.Unlock()
	p.SetIsInGame(true)
}
