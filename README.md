# How's Sophie?

A bot written in go to inform me about Sophie - my hackintosh PC running in my room. It posts tweets whether it has come online or went offline. Should help me to remotely monitor the current status, which is useful if I forget to turn her off for the night.


## Requirements

You will need the following services for this to run:
- [go^1.13](https://golang.org/dl/)
- [Redis](https://redis.io/download)
This has been tested on Linux only and the commands are given for a bash environment.


## Installation

To download the Go source code simply run
```bash
go get github.com/rubinda/hows-sophie
```

First you will need to provide some parameters inside the config file (check out the [example](configs/config.example.yml)). You will also need a running instance of Redis and some applications, which will publish messages to the Redis message broker. As an example, we have the [script](boot-shutdown.sh) that handles power on / off messages on macOS, included is also the [plist](boot-shutdown-script.plist) that makes it run at startup. You should now be ready to build and run the program:
```bash
go build ./cmd/sophie && ./sophie
```

