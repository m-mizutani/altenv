package main

import "github.com/keybase/go-keychain"

const keychainServiceNamePrefix = "altenv."

type keychainAddItem func(keychain.Item) error
type keychainUpdateItem func(keychain.Item, keychain.Item) error
type keychainQueryItem func(keychain.Item) ([]keychain.QueryResult, error)

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
