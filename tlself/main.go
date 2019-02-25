package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path"
	"sync"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cant figure out current user: %v", err)
		os.Exit(2)
	}

	if usr.Uid != "0" {
		fmt.Fprintf(os.Stderr, "run tlself with sudo")
		os.Exit(2)
	}

	listenStr := os.Getenv("LISTEN")
	if listenStr == "" {
		listenStr = "127.0.0.1:443"
	}

	backendStr := os.Getenv("BACKEND")
	if backendStr == "" {
		backendStr = "127.0.0.1:80"
	}

	certDir := os.Getenv("CERTDIR")
	if certDir == "" {
		certDir = path.Join(usr.HomeDir, ".tlself")
	}

	err = os.MkdirAll(certDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create cert dir %s: %v", certDir, err)
		os.Exit(2)
	}

	certFile := path.Join(certDir, "cert.pem")
	// trustCertFile is the path to the cert.pem on the host, which is different from
	// certFile because we are assuming we are running in a docker container
	trustCertDir := os.Getenv("TRUST_CERTDIR")
	trustCertFile := path.Join(trustCertDir, "cert.pem")
	keyFile := path.Join(certDir, "key.pem")
	root, loaded := LoadOrCreateRootCA(certFile, keyFile)

	if !loaded {
		yellow := func(s string) string {
			return "\033[32m" + s + "\033[0m"
		}
		cmd := systemTrustCommand(trustCertFile)
		fmt.Fprintf(os.Stderr, yellow("--------------------------------------------------\n"))
		fmt.Fprintf(os.Stderr, yellow("ATTENTION USER! Manual action required\n"))
		fmt.Fprintf(os.Stderr, yellow("Trust the root cert by running following cmd:\n%s\n"), cmd)
		fmt.Fprintf(os.Stderr, yellow("--------------------------------------------------\n"))
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
		},
		// Optional, for requesting certificates on the fly from Let's Encrypt
		// and stpling OCSP
		GetCertificate: root.GetCertificate,
	}

	ln, err := tls.Listen("tcp", listenStr, tlsConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to listen on 127.0.0.1:443: %v", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "TLS proxy running: %s => %s", listenStr, backendStr)

	p, err := newProxy(backendStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to make proxy to %s: %v", backendStr, err)
		os.Exit(2)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accepting connection: %v", err)
			continue
		}
		go p.proxy(conn)
	}
}

type proxy struct {
	backend *net.TCPAddr
}

func newProxy(backendStr string) (proxy, error) {
	var p proxy

	rAddr, err := net.ResolveTCPAddr("tcp", backendStr)
	if err != nil {
		return p, err
	}

	p = proxy{
		backend: rAddr,
	}
	return p, nil
}

func (p proxy) proxy(conn net.Conn) {

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing frontend connection: %v", err)
		}
	}()

	bConn, err := net.DialTCP("tcp", nil, p.backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to %s: %v", p.backend, err)
		return
	}
	defer func() {
		err := bConn.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing backend connection: %v", err)
		}
	}()

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		_, err := io.Copy(bConn, conn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending data to backend connection: %v", err)
		}
		wg.Done()
	}()

	go func() {
		_, err := io.Copy(conn, bConn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending data to frontend connection: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

func systemTrustCommand(certFile string) string {
	// checkSystemTrusted: security verify-cert -c certFile
	return "sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain " + certFile
}
