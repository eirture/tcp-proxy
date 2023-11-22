# tcp-proxy

A TCP proxy command line tool.

## Install

```sh
$ go get github.com/eirture/tcp-proxy/cmd/tcp-proxy
```

You can also download binary executor from [release page](https://github.com/eirture/tcp-proxy/releases).

## Usage

Print the help:

```sh
$ tcp-proxy -h
Usage:
  tcp-proxy REMOTE_IP [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT:]REMOTE_PORT_N] [flags]

Flags:
      --address string      Addresses to listen on. (default "127.0.0.1")
  -h, --help                help for tcp-proxy
  -x, --proxy string        Use the specified proxy (format: [protocol://]host[:port]).
      --rate-limit string   Set the send and receive rate limit to n per second. eg: 1MB
      --raw-bytes           Log bytes as raw number
      --tee-rec string      tee path of received data
      --tee-sen string      tee path of sent data
  -v, --version             Print the version information.
```

This command will proxy the `192.168.1.2:80` on `localhost:80`

```sh
$ tcp-proxy 192.168.1.2 80
```

or with another port on localhost

```sh
$ tcp-proxy 192.168.1.2 8080:80
```

with multiple ports

```sh
$ tcp-proxy 192.168.1.2 8080:80 8443:443
```

You can print the sent and received data to standard output

```sh
$ tcp-proxy 192.168.1.2 8080:80 8443:443 --tee-sen - --tee-rec -
```

## License

tcp-proxy is released under the Apache 2.0 license. See [LICENSE.txt](/LICENSE.txt)
