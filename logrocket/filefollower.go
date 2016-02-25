package main

import (
	"os"
	"bufio"
	"time"
	"io"
	"regexp"
	"ekp" 
	"fmt"
)

type FileFollower struct {
	path string
	offset int64
	config *Config
	regex *regexp.Regexp
	file *os.File
	reader *bufio.Reader
	
	messageChannel chan *FollowerMessage
	
	closeFlag chan bool
	closedFlag chan bool
	continueFlag chan bool
	
	running bool
	
	message string
	partialMessage string
}
func NewFileFollower(path string, offset int64, config *Config, msgChannel chan *FollowerMessage, regex *regexp.Regexp) (returnFileFollower *FileFollower, err error) {
	//fmt.Println("Starting filefollower on path " + path)
	
	returnFileFollower = &FileFollower{
		path : path,
		offset : offset,
		config : config,
		regex : regex,
		
		messageChannel : msgChannel,
		
		closeFlag : make(chan bool),
		closedFlag : make(chan bool),
		continueFlag : make(chan bool),
		
		running : false,
		
		message : "",
		partialMessage : "",
	}
	
	err = returnFileFollower.init()
	if err != nil {
		return
	}
	
	return
}
func (this *FileFollower) Close() {
	if !this.running {
		return
	}
	
	this.closeFlag <- true
	
	select {
		case <- this.closedFlag:
			return
	}
}
func (this *FileFollower) cont() {
	this.continueFlag <- true
}
func (this *FileFollower) follow() {
	for this.running {
		select {
			case <- this.continueFlag:
				this.readMessage()
			case <- this.closeFlag:
				this.running = false
				this.file.Close()
				this.closedFlag <- true
		}
	}
}
func (this *FileFollower) init() (err error){
	this.file, err = os.Open(this.path)
	if err != nil {
		return
	}
	
	_, err = this.file.Seek(this.offset, 0)
	if err != nil {
		return
	}
	
	this.reader = bufio.NewReader(this.file)
	
	this.running = true
	
	go this.follow()
	go this.cont()
	
	return
}
func (this *FileFollower) clearMessage() {
	this.message = ""
	if this.config.PrependHostname {
		if this.config.HostnameOverride == "" {
			this.message += this.config.hostname + " "
		} else {
			this.message += this.config.HostnameOverride + " "
		}
	}
}
func (this *FileFollower) readMessage() {
	for reading := true; reading; {
		line, err := this.reader.ReadString('\n')
		if len(line) > 0 {
			if line[len(line) -1] == '\n' {
				if this.partialMessage != "" {
					line = this.partialMessage + line
					this.partialMessage = ""
				}
				
				offset, offsetErr := this.file.Seek(0, 1)
				if offsetErr != nil {
					fmt.Println("Unable to get current offset in file.  This normally shouldn't happen.  Terminating file follower for path: " + this.path)
					this.running = false
					return
				}
				this.offset = offset
				
				if len(line) > 1 && line[len(line) - 2] == '\r' {
					line = line[:len(line) - 2]
				} else {
					line = line[:len(line) - 1]
				}
				
				if this.config.Multiline {
					if this.regex.Match([]byte(line)) {
						this.sendMessage()
						reading = false
					}
					
					if this.config.ReplaceNewline {
						line += this.config.NewlineReplacement
					} else {
						line += "\n"
					}
					
					this.message += line
				} else {
					this.message += line
					this.sendMessage()
					reading = false
				}
				
			} else {
				this.partialMessage += line
			}
		}
		
		if err != nil {
			if err == io.EOF {
				fileinfo, err := this.file.Stat()
				if err != nil {
					fmt.Println("Error getting info for file.  This shouldn't happen.  Killing file follower for path: " + this.path)
					this.running = false
					return
				}
				
//				fmt.Println("Reached end of file.")
//				fmt.Println("Filename: " + fileinfo.Name())
//				fmt.Printf("Filesize: %d\n", fileinfo.Size())
//				fmt.Println("File I'm supposed to be watching: " + this.path)
				
				if fileinfo.Size() < this.offset {
					//file has been truncated
					this.offset = 0
					this.file.Seek(0, 0)
					this.partialMessage = ""
					this.clearMessage()
				} else if fi, err := os.Stat(this.path); !os.SameFile(fileinfo, fi) {
					if err != nil {
						fmt.Println("File has been renamed and there doesn't appear to be a file with the same name.  Stopping file follower.")
						this.running = false
						reading = false
						return
					}
					
					//File has been renamed and a new file was created.
					this.file.Close()
					this.file, err = os.Open(this.path)
					if err != nil {
						fmt.Println("File has been renamed and I can't open the new file.")
						this.running = false
						reading = false
						return
					}
					this.offset = 0
					
					this.reader = bufio.NewReader(this.file)
					
				} else {
					go this.sleep()
					return
				}
				
				reading = false
			} else {
				//check for additional errors
				fmt.Println("Error reading line from file.")
				fmt.Println(err)
				fmt.Println("Stopping file follower on path: " + this.path)
				this.running = false
				return
			}
		}
	}
	
	go this.cont()
}
func (this *FileFollower) sendMessage() {
	this.messageChannel <- &FollowerMessage{
		offset : this.offset,
		path : this.path,
		msg : ekp.NewMessage(
			[]byte(""),
			[]byte(this.message),
		),
	}
	this.clearMessage()
}
func (this *FileFollower) sleep() {
	time.Sleep(time.Second)
	this.cont()
}

type FollowerMessage struct {
	offset int64
	path string
	msg *ekp.Message
}
