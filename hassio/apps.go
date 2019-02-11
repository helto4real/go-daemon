package main

// Since go is a compiled language the deamon have to know
// what applications that are available and their configuration
// name that will be referenced in yaml config "app: appname"
//
// Everytime you add a new app in go-daemon this list must be added
var apps = map[string]interface{}{}
