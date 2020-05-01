// +build darwin

package main

import (
	"fmt"

	"github.com/keybase/go-keychain"
	"github.com/pkg/errors"
)

type keychainAddItem func(keychain.Item) error
type keychainUpdateItem func(keychain.Item, keychain.Item) error
type keychainQueryItem func(keychain.Item) ([]keychain.QueryResult, error)

func setupKeychainFunc(extIO *ExtIOFunc) {
	extIO.KeychainAddItem = keychain.AddItem
	extIO.KeychainUpdateItem = keychain.UpdateItem
	extIO.KeychainQueryItem = keychain.QueryItem
}

func putKeyChainValues(args putKeyChainValuesArgs) error {
	prefix := keychainServiceNamePrefix
	if args.servicePrefix != "" {
		prefix = args.servicePrefix
	}

	for _, v := range args.envvars {
		item := keychain.NewItem()
		item.SetSecClass(keychain.SecClassGenericPassword)
		item.SetService(prefix + args.namespace)
		item.SetAccount(v.Key)
		item.SetDescription("altenv")
		item.SetData([]byte(v.Value))
		item.SetAccessible(keychain.AccessibleWhenUnlocked)
		item.SetSynchronizable(keychain.SynchronizableNo)

		err := args.addItem(item)
		if err == keychain.ErrorDuplicateItem {
			// Duplicate
			query := keychain.NewItem()
			query.SetSecClass(keychain.SecClassGenericPassword)
			query.SetService(keychainServiceNamePrefix + args.namespace)
			query.SetAccount(v.Key)
			// query.SetAccessGroup(keychainLabel)
			query.SetMatchLimit(keychain.MatchLimitAll)

			if err := args.updateItem(query, item); err != nil {
				return errors.Wrap(err, "Fail to update an existing item")
			}
		} else if err != nil {
			return errors.Wrap(err, "Fail to add a new keychain item")
		}
	}

	return nil
}

func getKeyChainValues(args getKeyChainValuesArgs) ([]*envvar, error) {
	prefix := keychainServiceNamePrefix
	if args.servicePrefix != "" {
		prefix = args.servicePrefix
	}

	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(prefix + args.namespace)
	query.SetMatchLimit(keychain.MatchLimitAll)
	query.SetReturnAttributes(true)

	results, err := args.queryItem(query)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to get keychain values")
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("Keychain items not found in %s", args.namespace)
	}

	var envvars []*envvar
	for _, result := range results {
		q := keychain.NewItem()
		q.SetSecClass(keychain.SecClassGenericPassword)
		q.SetService(prefix + args.namespace)
		q.SetMatchLimit(keychain.MatchLimitOne)
		q.SetAccount(result.Account)
		q.SetReturnData(true)

		data, err := args.queryItem(q)
		if err != nil {
			return nil, fmt.Errorf("Fail to get keychain value: `%s`", result.Account)
		}
		envvars = append(envvars, &envvar{
			Key:   result.Account,
			Value: string(data[0].Data),
		})
	}

	return envvars, nil
}
