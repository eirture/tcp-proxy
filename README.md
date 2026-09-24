# tcp-proxy

A TCP proxy command line tool.

## Install

```sh
$ go install github.com/eirture/tcp-proxy/cmd/tcp-proxy@latest
```

Or clone the repo and execute the `make install` command. You can also download binary executor from [release page](https://github.com/eirture/tcp-proxy/releases).

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

### Arguments

- `REMOTE_IP`: the remote host (IP or domain name) to forward to.
- `[LOCAL_PORT:]REMOTE_PORT`: the port mapping. If `LOCAL_PORT` is omitted, the same port as `REMOTE_PORT` is used locally. Multiple mappings can be given.

### Examples

Proxy `192.168.1.2:80` on `127.0.0.1:80`:

```sh
$ tcp-proxy 192.168.1.2 80
```

Use another port on localhost:

```sh
$ tcp-proxy 192.168.1.2 8080:80
```

Forward multiple ports:

```sh
$ tcp-proxy 192.168.1.2 8080:80 8443:443
```

Listen on all interfaces instead of only `127.0.0.1`:

```sh
$ tcp-proxy 192.168.1.2 8080:80 --address 0.0.0.0
```

Connect to the remote host through a proxy (`-x`/`--proxy`). Supported protocols are `http`, `socks5` and `socks5h` (`http` is used when the protocol is omitted), and credentials can be given as `protocol://user:password@host:port`. If `--proxy` is not set, the `ALL_PROXY` (and `NO_PROXY`) environment variables are used.

```sh
$ tcp-proxy 192.168.1.2 8080:80 -x socks5://127.0.0.1:1080
$ tcp-proxy 192.168.1.2 8080:80 -x 127.0.0.1:8080  # same as http://127.0.0.1:8080
$ tcp-proxy 192.168.1.2 8080:80 --proxy http://user:pass@127.0.0.1:8080
```

Limit the transfer rate (shared by sending and receiving), using human-readable sizes such as `500KB` or `1MB`:

```sh
$ tcp-proxy 192.168.1.2 8080:80 --rate-limit 1MB
```

Log the transferred byte counts of each connection as raw numbers instead of human-readable sizes:

```sh
$ tcp-proxy 192.168.1.2 8080:80 --raw-bytes
```

Copy the sent (`--tee-sen`) and received (`--tee-rec`) data to a file, or to standard output with `-`:

```sh
$ tcp-proxy 192.168.1.2 8080:80 8443:443 --tee-sen - --tee-rec -
$ tcp-proxy 192.168.1.2 8080:80 --tee-sen sent.log --tee-rec received.log
```

Print the version information:

```sh
$ tcp-proxy -v
```

## License

tcp-proxy is released under the Apache 2.0 license. See [LICENSE.txt](/LICENSE.txt)
