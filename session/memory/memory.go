package memory

import (
	"mux/session"
	"sync"
)

func init() {
	session.Register("memory",NewMemroy())
}

type Memory struct {
	m sync.Map
}

func NewMemroy() *Memory {
	return nil
}

func (m *Memory) Create(sid string) (session.Sessioner, error) {
	panic("")
}

func (*Memory) Read(sid string) (session.Sessioner, error) {
	panic("implement me")
}

func (*Memory) Delete(sid string) error {
	panic("implement me")
}

func (*Memory) GC(maxLifeTime int64) {
	panic("implement me")
}

