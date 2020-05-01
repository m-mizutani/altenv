// +build darwin

package main

import (
	"fmt"

	"github.com/keybase/go-keychain"
	"github.com/pkg/errors"
)

const keychainServiceNamePrefix = "altenv."

func putKeyChainValues(envvars []*envvar, namespace string) error {
	for _, v := range envvars {
		item := keychain.NewItem()
		item.SetSecClass(keychain.SecClassGenericPassword)
		item.SetService(keychainServiceNamePrefix + namespace)
		item.SetAccount(v.Key)
		item.SetDescription("altenv")
		item.SetData([]byte(v.Value))
		item.SetAccessible(keychain.AccessibleWhenUnlocked)
		item.SetSynchronizable(keychain.SynchronizableNo)

		err := keychain.AddItem(item)
		if err == keychain.ErrorDuplicateItem {
			// Duplicate
			query := keychain.NewItem()
			query.SetSecClass(keychain.SecClassGenericPassword)
			query.SetService(keychainServiceNamePrefix + namespace)
			query.SetAccount(v.Key)
			// query.SetAccessGroup(keychainLabel)
			query.SetMatchLimit(keychain.MatchLimitAll)

			if err := keychain.UpdateItem(query, item); err != nil {
				return errors.Wrap(err, "Fail to update an existing item")
			}
		} else {
			return errors.Wrap(err, "Fail to add a new keychain item")
		}
	}

	return nil
}

func getKeyChainValues(namespace string) ([]*envvar, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(keychainServiceNamePrefix + namespace)
	query.SetMatchLimit(keychain.MatchLimitAll)
	query.SetReturnAttributes(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to get keychain values")
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("Keychain items not found in %s", namespace)
	}

	var envvars []*envvar
	for _, result := range results {
		q := keychain.NewItem()
		q.SetSecClass(keychain.SecClassGenericPassword)
		q.SetService(keychainServiceNamePrefix + namespace)
		q.SetMatchLimit(keychain.MatchLimitOne)
		q.SetAccount(result.Account)
		q.SetReturnData(true)

		data, err := keychain.QueryItem(q)
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
