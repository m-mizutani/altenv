package main

const keychainServiceNamePrefix = "altenv."

type putKeyChainValuesArgs struct {
	envvars       []*envvar
	namespace     string
	servicePrefix string
	addItem       keychainAddItem
	updateItem    keychainUpdateItem
}

type getKeyChainValuesArgs struct {
	namespace     string
	servicePrefix string
	queryItem     keychainQueryItem
}
