package main

import (
	"os"
	"fmt"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
	"os/signal"
)

func main() {
	versionNumber := "0.3"
	
	confPath := "./conf/"
	varPath := "./var/"
	kafkaPath := "./kafka/"
	clientId, err := os.Hostname()
	if err != nil {
		clientId = ""
	}
	
	//parse command line arguments
	for x := 1; x < len(os.Args); x++ {
		value := os.Args[x]
		
		if value == "--version" {
			fmt.Println("MessageRocket Version " + versionNumber)
			fmt.Println("Written by Taylor Wolfe (taylor.wolfe@realpage.com)")
			return
		} else if value == "--help" {
			fmt.Println("Command line arguements are: ")
			fmt.Println("--confPath (string directory path)")
			fmt.Println("--varPath (string directory path)")
			fmt.Println("--kafkaPath (string directory path)")
			fmt.Println("--version")
			return
		} else if value == "--confPath" {
			if len(os.Args) > x + 1 {
				x++
				confPath = os.Args[x]
			} else {
				fmt.Println("Argument --confPath must be followed by a path.")
				return
			}
		} else if value == "--varPath" {
			if len(os.Args) > x + 1 {
				x++
				varPath = os.Args[x]
			} else {
				fmt.Println("Argument --varPath must be followed by a path.")
				return
			}
		} else if value == "--clientId" {
			if len(os.Args) > x + 1 {
				x++
				clientId = os.Args[x]
			} else {
				fmt.Println("Argument --clientId must be followed by a path.")
				return
			}
		} else if value == "--kafkaPath" {
			if len(os.Args) > x + 1 {
				x++
				clientId = os.Args[x]
			} else {
				fmt.Println("Argument --kafkaPath must be followed by a path.")
				return
			}
		} else {
			fmt.Println("Unknown command line argument: " + value)
			return
		}
	}
	
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Unable to get hostname from OS. Defaulting to empty string.")
		hostname = ""
	}
	
	//append trailing / if the path doesn't have it
	if confPath[len(confPath) - 1] != '/' {
		confPath += "/"
	}
	if varPath[len(varPath) - 1] != '/' {
		varPath += "/"
	}
	if kafkaPath[len(kafkaPath) - 1] != '/' {
		kafkaPath += "/"
	}
	
	//get kafkahosts
	hosts := GetHosts(kafkaPath)
	
	//get configs
	configs := GetConfigs(confPath)
	
	//get statuses
	statuses := GetStatuses(varPath)
	
	//start pathwatchers
	pathWatchers := make(map[string]*PathWatcher)
	watcherRunning := false
	for key, _ := range configs {
		if _, statusesExists := statuses[key]; !statusesExists {
			statuses[key] = NewStatuses()
		}
		
		if configs[key].PrependHostname {
			configs[key].hostname = hostname
		}
		
		fmt.Println("Creating watcher for config: " + configs[key].filename)
		watcher, err := NewPathWatcher(configs[key], statuses[key], varPath, hosts, clientId)
		if err != nil {
			fmt.Println("Unable to start PathWatcher for config: " + configs[key].filename + ". Skipping.")
			continue
		}
		watcherRunning = true
		pathWatchers[configs[key].filename] = watcher
	}	
	
	if !watcherRunning {
		fmt.Println("Unable to begin any path watchers.  Terminating application.")
		os.Exit(1)
	}
	
	//listen for interrupt signal	
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	select {
		case <- interruptChannel:
			for key, _ := range pathWatchers {
				pathWatchers[key].Close()
			}
	}
}

func GetConfigs(confPath string) (returnConfigs map[string]*Config) {
	returnConfigs = make(map[string]*Config)
	
	paths, err := filepath.Glob(confPath + "*")
	if err != nil {
		fmt.Println("Unable to get config file paths.  Terminating application.")
		os.Exit(1)
	}
	
	fmt.Printf("Number of files in config folder: %d\n", len(paths))	
	
	for _, path := range paths {
		_, filename := filepath.Split(path)
		if filename == "readme.txt" {
			continue
		}		
		
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("Unable to read config file: " + path + ".  Skipping.")
			continue
		}
		
		config := new(Config)
		err = json.Unmarshal(bytes, config)
		if err != nil {
			fmt.Println("Unable to unmarshal json from config file: " + path + ". Skipping.")
			fmt.Println(err)
			continue
		}
		config.filename = filename
		
		returnConfigs[filename] = config
	}
	
	if len(returnConfigs) < 1 {
		fmt.Println("No readable config files in the given config path.  Nothing to do.")
		os.Exit(0)
	}
	
	return
}
func GetHosts(kafkaPath string) (returnHosts []string) {
	bytes, err := ioutil.ReadFile(kafkaPath + "hosts.txt")
	if err != nil {
		fmt.Println("Unable to read kafka hosts file.  Terminating application.")
		os.Exit(1)
	}
	
	err = json.Unmarshal(bytes, &returnHosts)
	if err != nil {
		fmt.Println("Unable to unmarshal kafka hosts.  This usually means the json is malformed.  Terminating application.")
		fmt.Println(err)
		os.Exit(1)
	}
	
	return
}
func GetStatuses(varPath string) (returnStatuses map[string]*Statuses) {
	returnStatuses = make(map[string]*Statuses)
	
	paths, err := filepath.Glob(varPath + "*")
	if err != nil {
		fmt.Println("Unable to get paths for status files.")
		fmt.Println(err)
	}
	
	for _, path := range paths {
		_, filename := filepath.Split(path)
		if filename == "readme.txt" {
			continue
		}
		
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("Unable to read status file: " + path)
			continue
		}
		
		statuses := new(Statuses)
		err = json.Unmarshal(bytes, statuses)
		if err != nil {
			fmt.Println("Unable to unmarshal json from status file.  This shouldn't happen.")
			continue
		}
		
		returnStatuses[filename] = statuses
	}
	
	return
}
