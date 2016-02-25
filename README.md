# LogRocket

##Synopsis
The purpose of the LogRocket is to read lines/logs from log files and send them to a Kafka cluster with as little performance hit as possible.  

This code is running in production and has proven to be quite stable.

##Compiling
The LogRocket requires the EasyKafkaProducer which was developed alongside the LogRocket to assist in providing a lightweight client.

##Installation
The LogRocket simply needs to be compiled and run.  No runtime dependencies exist.

##Running LogRocket
logrocket /path/to/config/directory

##Config file format
Config files are simple json objects.  Config files can take any name other than "readme.txt".  
Capitalization of properties in the config files must be as shown.

###Config properties are:

Path (string, required):  																		The wildcard (*) capable path for logfiles you wish to forward 

Topic (string, required): 																		The Kafka topic that log messages will be sent to

Multiline (bool, optional default false): 														Flag to indicate whether log messages are multiline

FirstLineRegex (string, required when Multiline is true): 										The regex pattern that the first line of a log message will be.  This must meet the Go regex specification.

ReplaceNewline (bool, optional default false): 													Flag to indicate whether newline characters should be replaced

NewlineReplacement (string, required if ReplaceNewline is true): 								The string to replace newline characters with

FromBeginning (bool, optional default false): 													Flag to indicate whether historic log messages should be harvested or only new ones

PrependHostname (bool, optional default false): 												Flag to indicate whether log messages should have the current machines hostname prepended to it

HostnameOverride (string, optional default null, only applicable if PrependHostname is true): 	String to override the hostname of the current machine



## License

MIT License:

Copyright (c) 2016, Charles Wolfe and contributors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.