# tgo

A minimal, high-performance and low-resource CLI tool for managing task lists in an open `.json` files, perfect for file sync (e.g. NextCloud).

## Commands

- `tgo set-dir <path>`: Set the directory for your task lists.
- `tgo`: Open interactive mode to view and manage tasks.
- `tgo done <number>`: Mark a task as done or undone.
- `tgo help`: Show help info.

## Quick Start

Pre-built binaries for macOS, Linux, and Windows are available on the [Releases page](https://github.com/salernoelia/tgo/releases).


## Building yourself

```sh
./tgo set-dir ~/Tasks
./tgo
./tgo done 2
```

## Build

```sh
go mod tidy
go build -o tgo .
```

## Install (macOS)

```sh
sudo mv tgo /usr/local/bin/
```


