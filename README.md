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
      --address string   Addresses to listen on. (default "127.0.0.1")
  -h, --help             help for tcp-proxy
  -x, --proxy string     Use the specified proxy (format: [protocol://]host[:port]).
  -v, --version          Print the version information.
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

## License

tcp-proxy is released under the Apache 2.0 license. See [LICENSE.txt](/LICENSE.txt)
