package main

import (
	//"encoding/json"
)

type Config struct {
	Path string
	Topic string
	Multiline bool
	FirstLineRegex string
	ReplaceNewline bool
	NewlineReplacement string
	FromBeginning bool
	PrependHostname bool
	HostnameOverride string
	
	filename string
	hostname string
}
