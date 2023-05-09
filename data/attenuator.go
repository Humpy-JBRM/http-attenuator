package data

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Attenuated interface {
	Fire(params ...interface{}) error
}

// Attenuator is used to regulate operations
//
// TODO(john): Poisson process
type Attenuator interface {
	Wait()
	Start() error
	Pause()
	WeaponsFree()
}

type attenuatorImpl struct {
	name        string
	hertz       int64
	barrier     chan interface{}
	pause       sync.Mutex
	weaponsFree bool
	trigger     Attenuated
}

func (a *attenuatorImpl) Wait() {
	<-a.barrier
}

func (a *attenuatorImpl) Pause() {
	a.pause.Lock()
}

func (a *attenuatorImpl) Resume() {
	a.pause.Unlock()
}

func (a *attenuatorImpl) WeaponsFree() {

}

func (a *attenuatorImpl) tickTrigger(waitTime time.Duration) error {
	for {
		if a.weaponsFree {
			a.barrier <- true
			continue
		}

		select {
		case <-time.After(waitTime):
			a.pause.Lock()
			log.Printf("%s: TICK", a.name)
			a.trigger.Fire()
			a.pause.Unlock()
		}
	}
}

func (a *attenuatorImpl) Start() error {
	waitTimeMs := float64(1000) / float64(a.hertz)
	if waitTimeMs <= 0 {
		return fmt.Errorf("ERROR|attenuator.Start()|Cannot create new attenuator|Wait time (%.2f) is <= 0", waitTimeMs)
	}

	log.Printf("INFO|attenuator.Start()|Starting attenuator '%s' @%dHz (wait = %fms)|", a.name, a.hertz, waitTimeMs)
	go a.tickTrigger(time.Duration(int64(waitTimeMs)) * time.Millisecond)
	return nil
}

func NewAttenuator(name string, hertz int64, start bool, register bool, trigger Attenuated) (Attenuator, error) {
	attenuator := &attenuatorImpl{
		name:    name,
		hertz:   hertz,
		barrier: make(chan interface{}, 100),
		trigger: trigger,
	}

	if register {
		GetAttenuatorRegistry().RegisterAttenuator(name, attenuator)
	}

	if start {
		attenuator.Start()
	}

	return attenuator, nil
}
