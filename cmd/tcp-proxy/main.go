package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
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
var (
	logBytesAsRawNumber bool
)

var (
	teeSentWriter     io.WriteCloser
	teeReceivedWriter io.WriteCloser
)

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
			cch := make(chan struct{}, 2)
			rch := make(chan int64, 1) // receive
			sch := make(chan int64, 1) // send
			defer close(cch)
			defer close(rch)
			defer close(sch)
			defer func() {
				log.Infof("Connection done %s: ↑%s ↓%s", conn.RemoteAddr(), formatBytes(<-sch), formatBytes(<-rch))
			}()

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
					log.Errorf("error dialing from proxy. %v\n", err)
				}
			}
			conn2, err := dialer.Dial("tcp", remoteAddr)
			if err != nil {
				log.Errorln("error dialing remote addr", err)
				return
			}
			defer conn2.Close()
			var laReader, raReader io.Reader = conn, conn2
			var laWriter, raWriter io.Writer = conn, conn2
			if limitBucket != nil {
				laReader = ratelimit.Reader(laReader, limitBucket)
				raReader = ratelimit.Reader(raReader, limitBucket)
			}
			if teeSentWriter != nil {
				laReader = io.TeeReader(laReader, teeSentWriter)
			}
			if teeReceivedWriter != nil {
				raReader = io.TeeReader(raReader, teeReceivedWriter)
			}

			go func() {
				rn, _ := io.Copy(laWriter, raReader)
				cch <- struct{}{}
				rch <- rn
			}()
			go func() {
				wn, _ := io.Copy(raWriter, laReader)
				cch <- struct{}{}
				sch <- wn
			}()
			<-cch
		}()
	}
}

type CloseFunc func() error

var NopCloseFn CloseFunc = func() error { return nil }

type WriteCloser struct {
	io.Writer
	cfn CloseFunc
}

func NewWriteCloser(w io.Writer, cfn CloseFunc) io.WriteCloser {
	return &WriteCloser{
		Writer: w,
		cfn:    cfn,
	}
}

func (nwc *WriteCloser) Write(p []byte) (int, error) {
	return nwc.Writer.Write(p)
}

func (nwc *WriteCloser) Close() error {
	if nwc.cfn == nil {
		return nil
	}
	return nwc.cfn()
}

func openTeePath(path string) (io.WriteCloser, error) {
	switch path {
	case "-":
		return NewWriteCloser(os.Stdout, NopCloseFn), nil
	case "":
		return nil, nil
	default:
		fm := os.FileMode(0666)
		if fi, err := os.Lstat(path); err == nil {
			fm = fi.Mode()
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, fm)
	}
}

var (
	sizeUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}
)

func formatBytes(n int64) string {
	if logBytesAsRawNumber {
		return strconv.FormatInt(n, 10)
	}
	return humanSize(n)
}

func humanSize(n int64) string {
	var decimal int64
	var i int
	for n > 1024 && i < len(sizeUnits) {
		decimal = n % 1024
		n /= 1024
		i++
	}
	if decimal > 100 {
		return fmt.Sprintf("%d.%d%s", n, decimal/10, sizeUnits[i])
	}
	return fmt.Sprintf("%d%s", n, sizeUnits[i])
}

type RootOptions struct {
	address     string
	proxy       string
	teeSent     string
	teeReceived string

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
	cmd.Flags().BoolVar(&logBytesAsRawNumber, "raw-bytes", false, "log bytes as raw number")
	cmd.Flags().StringVar(&ops.teeSent, "tee-sen", "", "tee path of sent data")
	cmd.Flags().StringVar(&ops.teeReceived, "tee-rec", "", "tee path of received data")

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

	if teeReceivedWriter, err = openTeePath(o.teeReceived); err != nil {
		return
	} else if teeReceivedWriter != nil {
		defer teeReceivedWriter.Close()
	}
	if o.teeReceived == o.teeSent {
		teeSentWriter = teeReceivedWriter
	} else if teeSentWriter, err = openTeePath(o.teeSent); err != nil {
		return err
	} else if teeSentWriter != nil {
		defer teeSentWriter.Close()
	}

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
