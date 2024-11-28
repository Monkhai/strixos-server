package game

import "sync"

type InviteGameManager struct {
	Map map[string]*Game
	Mux *sync.RWMutex
}

func NewInviteGameManager() *InviteGameManager {
	return &InviteGameManager{
		Map: make(map[string]*Game),
		Mux: &sync.RWMutex{},
	}
}

func (i *InviteGameManager) AddGame(newGame *Game) {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	i.Map[newGame.ID] = newGame
}

func (i *InviteGameManager) RemoveGame(gameID string) {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	delete(i.Map, gameID)
}

func (i *InviteGameManager) GetGame(gameID string) (*Game, bool) {
	i.Mux.RLock()
	defer i.Mux.RUnlock()
	game, ok := i.Map[gameID]
	return game, ok
}
