package main

import (
	"log"
	"sync"
	"time"

	"github.com/CloudPassenger/rnm-go/config"
	"github.com/CloudPassenger/rnm-go/dispatcher"
	_ "github.com/CloudPassenger/rnm-go/dispatcher/tcp"
)

const HttpClientTimeout = 10 * time.Second

type MapPortDispatcher map[int]*[len(protocols)]dispatcher.Dispatcher

type SyncMapPortDispatcher struct {
	sync.Mutex
	Map MapPortDispatcher
}

func NewSyncMapPortDispatcher() *SyncMapPortDispatcher {
	return &SyncMapPortDispatcher{Map: make(MapPortDispatcher)}
}

var (
	protocols       = [...]string{"tcp"}
	groupWG         sync.WaitGroup
	mPortDispatcher = NewSyncMapPortDispatcher()
)

func listenGroup(group *config.Group) {
	err := listenProtocols(group, protocols[:])
	if err != nil {
		mPortDispatcher.Lock()
		// error but listening
		if _, ok := mPortDispatcher.Map[group.Port]; ok {
			log.Fatalln(err)
		}
		mPortDispatcher.Unlock()
	}
}

func listenProtocols(group *config.Group, protocols []string) error {
	mPortDispatcher.Lock()
	if _, ok := mPortDispatcher.Map[group.Port]; !ok {
		mPortDispatcher.Map[group.Port] = new([1]dispatcher.Dispatcher)
	}
	t := mPortDispatcher.Map[group.Port]
	mPortDispatcher.Unlock()

	ch := make(chan error, len(protocols))
	for i, protocol := range protocols {
		d, _ := dispatcher.New(protocol, group)
		(*t)[i] = d
		go func() {
			err := d.Listen()
			ch <- err
		}()
	}
	return <-ch
}

func main() {
	conf := config.NewConfig()

	// handle reload
	go signalHandler(conf)

	mPortDispatcher.Lock()
	for i := range conf.Groups {
		groupWG.Add(1)
		go func(group *config.Group) {
			listenGroup(group)
			groupWG.Done()
		}(&conf.Groups[i])
	}
	mPortDispatcher.Unlock()
	groupWG.Wait()
}
