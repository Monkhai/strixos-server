package server

import (
	"log"
	"sync"

	"github.com/Monkhai/strixos-server.git/internal/game"
)

type PlayerNode struct {
	Player *game.Player
	Prev   *PlayerNode
	Next   *PlayerNode
}

type PlayerQueue struct {
	Head *PlayerNode
	Tail *PlayerNode
	Map  map[string]*PlayerNode
	Mux  *sync.RWMutex
}

func (q *PlayerQueue) Enqueue(p *game.Player) {
	q.Mux.Lock()
	defer q.Mux.Unlock()

	node := &PlayerNode{Player: p}

	if q.Head == nil {
		q.Head = node
		q.Tail = node
	} else {
		q.Tail.Next = node
		node.Prev = q.Tail
		q.Tail = node
	}

	q.Map[p.Identity.ID] = node
}

func (q *PlayerQueue) Dequeue() *game.Player {
	q.Mux.Lock()
	defer q.Mux.Unlock()

	if q.Head == nil {
		return nil
	}

	node := q.Head
	q.Head = node.Next

	if q.Head == nil {
		q.Tail = nil
	} else {
		q.Head.Prev = nil
	}

	delete(q.Map, node.Player.Identity.ID)
	return node.Player
}

func (q *PlayerQueue) RemovePlayer(p *game.Player) {
	q.Mux.Lock()
	defer q.Mux.Unlock()

	node, exists := q.Map[p.Identity.ID]
	if !exists {
		return
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		q.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		q.Tail = node.Prev
	}

	delete(q.Map, p.Identity.ID)
}

func (q *PlayerQueue) GetTwoPlayers() ([2]*game.Player, bool) {
	q.Mux.RLock()

	if q.Head == nil || q.Head.Next == nil {
		q.Mux.RUnlock()
		return [2]*game.Player{nil, nil}, false
	}
	q.Mux.RUnlock()

	playerOne := q.Dequeue()
	playerTwo := q.Dequeue()

	return [2]*game.Player{playerOne, playerTwo}, true
}

func (q *PlayerQueue) IsPlayerInQueue(id string) bool {
	q.Mux.RLock()
	defer q.Mux.RUnlock()

	_, ok := q.Map[id]
	return ok
}

func NewPlayerQueue() *PlayerQueue {
	return &PlayerQueue{
		Head: nil,
		Tail: nil,
		Map:  make(map[string]*PlayerNode),
		Mux:  &sync.RWMutex{},
	}
}

func (q *PlayerQueue) printQueue() {
	q.Mux.RLock()
	defer q.Mux.RUnlock()

	for node := q.Head; node != nil; node = node.Next {
		log.Println(node.Player.Identity.ID)
	}
}
