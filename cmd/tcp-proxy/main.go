package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/docker/go-units"
	"github.com/eirture/tcp-proxy/pkg/build"
	"github.com/eirture/tcp-proxy/pkg/log"
	"github.com/juju/ratelimit"
	"github.com/spf13/cobra"
	"golang.org/x/net/proxy"
)

var limitBucket *ratelimit.Bucket

func listen(localAddr, remoteAddr, proxyAddr string) (err error) {
	log.Infof("Forwarding from %s -> %s\n", localAddr, remoteAddr)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		log.Infoln("New connection", conn.RemoteAddr())
		if err != nil {
			log.Errorln("error accepting connection", err)
			continue
		}
		go func() {
			defer conn.Close()
			var dialer proxy.Dialer

			if proxyAddr == "" {
				dialer = proxy.FromEnvironment()
			} else {
				proxyUrl, err := url.Parse(proxyAddr)
				if err != nil {
					log.Errorf("Invalid proxy address. err: %v\n", err)
				}
				dialer, err = proxy.FromURL(proxyUrl, proxy.Direct)
				if err != nil {
					log.Errorf("error dialing from proxy. %v\n")
				}
			}
			conn2, err := dialer.Dial("tcp", remoteAddr)
			if err != nil {
				log.Errorln("error dialing remote addr", err)
				return
			}
			defer conn2.Close()
			closer := make(chan struct{}, 2)
			var r1, r2 io.Reader = conn, conn2
			if limitBucket != nil {
				r1 = ratelimit.Reader(r1, limitBucket)
				r2 = ratelimit.Reader(r2, limitBucket)
			}
			go copyWithCloser(closer, conn2, r1)
			go copyWithCloser(closer, conn, r2)
			<-closer
			log.Infoln("Connection complete", conn.RemoteAddr())
		}()
	}
}

func copyWithCloser(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}

type RootOptions struct {
	address string
	proxy   string

	version   bool
	rateLimit string
}

func NewRootCmd() *cobra.Command {
	ops := RootOptions{}

	cmd := &cobra.Command{
		Use:           "tcp-proxy REMOTE_IP [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT:]REMOTE_PORT_N]",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          ops.Run,
	}

	cmd.Flags().StringVar(&ops.address, "address", "127.0.0.1", "Addresses to listen on.")
	cmd.Flags().StringVarP(&ops.proxy, "proxy", "x", "", "Use the specified proxy (format: [protocol://]host[:port]).")
	cmd.Flags().BoolVarP(&ops.version, "version", "v", false, "Print the version information.")
	cmd.Flags().StringVar(&ops.rateLimit, "rate-limit", "", "")

	return cmd
}

func (o *RootOptions) Run(cmd *cobra.Command, args []string) (err error) {

	if o.version {
		build.PrintVersion()
		return nil
	}

	if len(args) < 2 {
		return fmt.Errorf("requires at least 2 arg(s), only received %d", len(args))
	}

	if bps, err := units.FromHumanSize(o.rateLimit); err == nil {
		limitBucket = ratelimit.NewBucketWithRate(float64(bps), bps)
	}

	remote := args[0]
	ports := args[1:]

	var wg sync.WaitGroup
	for _, port := range ports {
		ps := strings.Split(port, ":")
		if len(ps) > 2 {
			log.Errorf("invalid port %s", port)
		}
		if len(ps) == 1 {
			ps = append(ps, ps[0])
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := listen(
				fmt.Sprintf("%s:%s", o.address, ps[0]),
				fmt.Sprintf("%s:%s", remote, ps[1]),
				o.proxy,
			); err != nil {
				log.Error(err)
			}
		}()
	}

	wg.Wait()

	return
}

func main() {
	rootCmd := NewRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		log.Errorln(err)
	}
}
