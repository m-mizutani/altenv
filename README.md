# altenv [![Travis-CI](https://travis-ci.org/m-mizutani/altenv.svg)](https://travis-ci.org/m-mizutani/altenv) [![Report card](https://goreportcard.com/badge/github.com/m-mizutani/altenv)](https://goreportcard.com/report/github.com/m-mizutani/altenv)

**NOTE: This product has been deprecated and archived. Use [zenv](https://github.com/m-mizutani/zenv) instead.**

CLI Environment Variable Controller in Go

```sh
$ cat setting.env
DB_NAME = AWESOME_DATABASE
$ cat sample.py
import os
print("my database is {0}".format(os.environ["DB_NAME"]))
$ altenv -e setting.env python sample.py
my database is AWESOME_DATABASE
```

## Install

```
$ go get -u github.com/m-mizutani/altenv
```

## Usage

### Read variables from env file

Sample `yourfile.env`
```
# Comment out is available
KEY1 = ABC
KEY2 = BCD
```

```sh
$ altenv -e yourfile.env <command> [arg1, [arg2, [...]]]
```

### Read variables from JSON file

Sample `yourfile.json`. JSON file must be map format with pairs of key(string) and value(string).
```json
{
    "KEY1": "ABC",
    "KEY2": "BCD"
}
```

```sh
$ altenv -j yourfile.json <command> [arg1, [arg2, [...]]]
```

### Confirm new environment variables

`altenv` provides dryrun feature to confirm conputed environment variables.

```sh
$ altenv -i yourfile.env -r dryrun
KEY1=ABC
KEY2=BCD
```

### Input from prompt

If you want to hide input value, you can use `--prompt` option for no-echo input.

```sh
$ altenv --prompt TOKEN ./deploy.sh
Enter TOKEN value:
```

### Use Keychain (only for macOS)

`altenv` can saves environment variable to macOS Keychain and loads saved variable from Keychain. This feature is appropriate to manage secret values e.g. credential key, token, etc. `altenv` can have multiple namespaces. This is inspired by [envchain](https://github.com/sorah/envchain).

If you already have plain text credential, you can use stdin option, `-i env`. `-i` option required input format and `env` means `KEY1=ABC` style.

```sh
$ cat credentials.env | altenv -r update-keychain -w your_namespace -i env
```

Or you can set secrets with `--prompt` option. Prompt input works with no echo, then you can put secrets more securely.

```sh
$ altenv -r update-keychain -w your_namespace --prompt AWS_SECRET_ACCESS_KEY
Enter AWS_SECRET_ACCESS_KEY value:
```

After that, you can import saved secret values by `-k <namespace>` option.

```sh
$ altenv -k your_namespace aws s3 ls
```

## Configuration

altenv imports config file when invoked. `$HOME/.altenv` is default config path (ignore if not exists) and config path can be specified `-c` option. Config file can be written in toml format.

### Sections

There are 3 types of section in configuration file.

- `global`: The section's configurations are always imported.
- `profile.xxx`: Profile can be switched by CLI option `-p`. `profile.xxx` is imported when `-p xxx` is given in CLI.
  - `profile.default`: The section is imported by default. If you specifiy profile name other than `default`, this section is not imported.
- `workdir.xxx`: WorkDir section is enabled by your current working directory. Directory path can be specified by `dirpath` (See *Configuration fields* part). `dirpath` works as directory path prefix. If multiple `dirpath` are matcehd with current working directory, all matched configurations are imported. NOTE: `xxx` is just label in WorkDir section.

### Configuration fields

- `envfile` (array of string): Specify envfile foramt file(s). (multiple lines with `KEY1=ABC` style)
- `jsonfile` (array of string): Specify json format file(S). Only map format of string key and value is acceptable.
- `define` (array of string): Specify environment variable(s) directly with `KEY1=ABC` style.
- `keychain` (array of string): Specify namespace(s) for environment variables stored in Keychain. See *Use Keychain* part.
- `overwrite` (string, [`deny`|`warn`|`allow`]): Specify Overwrite policy. Default is `deny` and `altenv` abort program when environment variable key conflict. `warn` is only output warning message. `allow` allows overwrite when collision.
- `keychainServicePrefix`: Specify prefix of service name of Keychain. Default is `altenv.`
- `dirpath` (string): Required in only `workdir` section. Specify prefix of working directoy.

### Example configuration

Example of `$HOME/.altenv` is following.

```toml
[global]
define = ["ALWAYS=ENABLED"]

[profile.default]
define = ["PROFILE_IS=DEFAULT"]

[profile.mytest]
define = ["PRFILE_IS=MY_TEST"]

[workdir.proj1]
dirpath = "/Users/mizutani/works/proj1"
define = ["DBNAME=proj1"]

[workdir.proj2]
dirpath = "/Users/mizutani/works/proj2"
define = ["DBNAME=proj2"]
```

Run with the configuration.

```sh
$ cd /Users/mizutani
$ altenv -r dryrun
ALWAYS=ENABLED
PROFILE_IS=DEFAULT
$ cd /Users/mizutani/works/proj1
$ altenv -r dryrun
ALWAYS=ENABLED
DBNAME=proj1
PROFILE_IS=DEFAULT
$ altenv -r dryrun -p mytest
ALWAYS=ENABLED
DBNAME=proj1
PRFILE_IS=MY_TEST
```

## License

- MIT License
- Author: mizutani@hey.com
