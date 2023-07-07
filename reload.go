package main

import (
	"log"

	"github.com/CloudPassenger/rnm-go/config"
)

func ReloadConfig(oldConf *config.Config) {
	log.Println("Reloading configuration")
	mPortDispatcher.Lock()
	defer mPortDispatcher.Unlock()

	// rebuild config
	confPath := oldConf.ConfPath
	// httpClient := oldConf.HttpClient
	newConf, err := config.BuildConfig(confPath)
	if err != nil {
		log.Printf("failed to reload configuration: %v", err)
		return
	}
	// check if there is any net error when pulling the upstream configurations
	config.SetConfig(newConf)
	c := newConf

	// update dispatchers
	newConfPortSet := make(map[int]struct{})
	for i := range c.Groups {
		newConfPortSet[c.Groups[i].Port] = struct{}{}

		if t, ok := mPortDispatcher.Map[c.Groups[i].Port]; ok {
			// update the existing dispatcher
			for j := range protocols {
				t[j].UpdateGroup(&c.Groups[i])
			}
		} else {
			// add a new port dispatcher
			groupWG.Add(1)
			go func(group *config.Group) {
				listenGroup(group)
				groupWG.Done()
			}(&c.Groups[i])
		}
	}
	// close all removed port dispatcher
	for port := range mPortDispatcher.Map {
		if _, ok := newConfPortSet[port]; !ok {
			t := mPortDispatcher.Map[port]
			delete(mPortDispatcher.Map, port)
			for j := range protocols {
				_ = (*t)[j].Close()
			}
		}
	}
	log.Println("Reloaded configuration")
}
