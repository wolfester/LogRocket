package main

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Statuses struct {
	Statuses map[string]*Status
}
func NewStatuses() (returnStatuses *Statuses) {
	returnStatuses = &Statuses{
		Statuses : make(map[string]*Status),
	}
	
	return
}
func (this *Statuses) AddStatus(status *Status) {
	this.Statuses[status.Path] = status
}
func (this *Statuses) GetStatus(path string) *Status {
	return this.Statuses[path]
}
func (this *Statuses) HasStatus(path string) bool {
	_, exists := this.Statuses[path]
	return exists
}
func (this *Statuses) Len() int {
	return len(this.Statuses)
}
func (this *Statuses) Dump(path string) {
	for key, _ := range this.Statuses {
		this.Statuses[key].SetLastOffsetToCurrentOffset()
	}
	
	bytes, err := json.MarshalIndent(this, "", "\t")
	if err != nil {
		fmt.Println("Error encoding statuses to json.  This probably shouldn't happen.")
		return
	}
	
	err = ioutil.WriteFile(path, bytes, 0666)
	if err != nil {
		fmt.Println("Error writing statuses to file: " + path + ".")
		fmt.Println(err)
	}
}

type Status struct {
	Path string
	LastOffset int64
	CurrentOffset int64
}
func NewStatus(path string, lastOffset int64, currentOffset int64) (returnStatus *Status) {
	returnStatus = &Status{
		Path : path,
		LastOffset : lastOffset,
		CurrentOffset : currentOffset,
	}
	
	return
}
func (this *Status) SetCurrentOffset(offset int64) {
	this.CurrentOffset = offset
}
func (this *Status) SetLastOffsetToCurrentOffset() {
	this.LastOffset = this.CurrentOffset
}
