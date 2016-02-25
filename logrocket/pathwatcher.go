package main

import (
	"time"
	"path/filepath"
	"fmt"
	"os"
	"ekp"
	"regexp"
	"sync"
)

type PathWatcher struct {
	config *Config
	statuses *Statuses
	statusesLock *sync.Mutex
	
	varPath string
	
	followers map[string]*FileFollower
	messageChannel chan *FollowerMessage
	regex *regexp.Regexp
	
	messageSet *ekp.MessageSet
	ekp *ekp.SKP
	hosts []string
	clientId string
	
	closeFlag chan bool
	closedFlag chan bool
	continueFlag chan bool
	
	flushFlag chan bool
	flushedFlag chan bool
	
	running bool
	spooling bool
	firstRun bool
}
func NewPathWatcher(config *Config, statuses *Statuses, varPath string, hosts []string, clientId string) (returnPathWatcher *PathWatcher, err error) {
	//fmt.Println("Starting pathwatcher on path " + config.Path)
	
	returnPathWatcher = &PathWatcher {
		config : config,
		statuses : statuses,
		statusesLock : &sync.Mutex{},
		
		varPath : varPath,
		
		followers : make(map[string]*FileFollower),
		messageChannel : make(chan *FollowerMessage, 10000),
		
		messageSet : ekp.NewMessageSet(),
		hosts : hosts,
		clientId : clientId,
		
		closeFlag : make(chan bool),
		closedFlag : make(chan bool),
		continueFlag : make(chan bool),
		
		flushFlag : make(chan bool),
		flushedFlag : make(chan bool),
		
		running : false,
		spooling : false,
	}
	
	returnPathWatcher.init()
	
	return
}
func (this *PathWatcher) Close() {
	if !this.running {
		return
	}
	this.closeFlag <- true
	
	select {
		case <- this.closedFlag:
			return
	}
}
func (this *PathWatcher) checkPath() {
	paths, err := filepath.Glob(this.config.Path)
	if err != nil {
		fmt.Println("Malformed path in configuration file.  Path: " + this.config.Path)
		fmt.Println("Terminating PathWatcher.")
		this.running = false
		return
	}
	
	noStatuses := false
	if this.statuses.Len() < 1 {
		noStatuses = true
	}
	
	for _, path := range paths {
		if _, followerExists := this.followers[path]; !followerExists {
			//fmt.Println("Creating file follower")
			if this.statuses.HasStatus(path) {
				follower, err := NewFileFollower(path, this.statuses.GetStatus(path).LastOffset, this.config, this.messageChannel, this.regex)
				if err != nil {
					fmt.Println("Error creating file follower for path: " + path + ".  Skipping.")
					continue
				}
				this.followers[path] = follower
			} else {
				//status doesn't exist yet... figure out what to do
				if this.config.FromBeginning || !noStatuses || !this.firstRun {
					//create new status and make follower from beginning
					this.statusesLock.Lock()
					this.statuses.AddStatus(
						NewStatus(path, int64(0), int64(0)),
					)
					this.statusesLock.Unlock()
					follower, err := NewFileFollower(path, this.statuses.GetStatus(path).LastOffset, this.config, this.messageChannel, this.regex)
					if err != nil {
						fmt.Println("Error creating file follower for path: " + path + ".  Skipping.")
						continue
					}
					this.followers[path] = follower
				} else {
					//create new status and make follower from end
					fileinfo, err := os.Stat(path)
					if err != nil {
						fmt.Println("Error getting file info.")
					}
					this.statuses.AddStatus(
						NewStatus(path, fileinfo.Size(), fileinfo.Size()),
					)
					follower, err := NewFileFollower(path, this.statuses.GetStatus(path).LastOffset, this.config, this.messageChannel, this.regex)
					if err != nil {
						fmt.Println("Error creating file follower for path: " + path + ".  Skipping.")
						continue
					}
					this.followers[path] = follower
				}
			}
		} else if !this.followers[path].running {
			//restart follower
		}
	}
	
	this.statusesLock.Lock()
	this.dumpStatuses()
	this.statusesLock.Unlock()
	
	this.firstRun = false
	go this.sleep()
}
func (this *PathWatcher) dumpStatuses() {
	this.statuses.Dump(this.varPath + this.config.filename)
}
func (this *PathWatcher) flush() {
	for this.running {
		time.Sleep(time.Second)
		this.flushFlag <- true
		select {
			case <- this.flushedFlag:
				continue
		}
	}
}
func (this *PathWatcher) init() {
	this.ekp = ekp.NewSKP(this.hosts, this.clientId)
	
	if this.config.Multiline {
		regex, err := regexp.Compile(this.config.FirstLineRegex)
		if err != nil {
			fmt.Println("Unable to compile regex for config: " + this.config.filename)
			fmt.Println("Not watching path.")
			return
		}
		this.regex = regex
	}
	
	this.running = true
	this.spooling = true
	this.firstRun = true	
	
	go this.watch()
	go this.spool()
	go this.flush()
	
	go func(){
		this.continueFlag <- true
	}()
}
func (this *PathWatcher) sendMessages() {
	err := this.ekp.ProduceMessageSet(this.config.Topic, this.messageSet)
	if err != nil {
		//fmt.Println("Error producing to kafka.")
		time.Sleep(time.Second / 2)
		this.sendMessages()
	} else {
		this.messageSet.Reset()
		this.statusesLock.Lock()
		this.dumpStatuses()
		this.statusesLock.Unlock()
	}
}
func (this *PathWatcher) sleep() {
	time.Sleep(time.Minute)
	this.continueFlag <- true
}
func (this *PathWatcher) spool() {
	for this.spooling {
		select {
			case msg := <- this.messageChannel:
				this.statusesLock.Lock()
				this.statuses.GetStatus(msg.path).SetCurrentOffset(msg.offset)
				this.statusesLock.Unlock()
				numMessages := this.messageSet.AddMessage(msg.msg)
				if numMessages > 1000 {
					this.sendMessages()
				}
			case <- this.flushFlag:
				if this.messageSet.Position() > 0 {
					this.sendMessages()
				}
				
				this.flushedFlag <- true
		}
	}
}
func (this *PathWatcher) watch() {
	for this.running {
		select {
			case <- this.continueFlag:
				this.checkPath()
			case <- this.closeFlag:
				for key, _ := range this.followers {
					this.followers[key].Close()
				}
				this.running = false
				this.spooling = false
				this.flushFlag <- true
				select {
					case <- this.flushedFlag:
					this.closedFlag <- true
				}
		}
	}
}
