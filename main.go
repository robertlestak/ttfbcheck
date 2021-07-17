package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	testUrl    string
	concurrent int
	total      int
	output     string
	logLevel   string
)

type SiteTiming struct {
	URL          string
	DNSDone      time.Duration
	tlsHandshake time.Duration
	TCPConnect   time.Duration
	TTFB         time.Duration
	TotalTime    time.Duration
}

func (t *SiteTiming) JSON() string {
	l := log.WithFields(log.Fields{
		"action": "JSON",
		"url":    t.URL,
	})
	l.Debug("Starting")
	jd, err := json.Marshal(t)
	if err != nil {
		l.WithError(err).Error("Could not marshal JSON")
		return ""
	}
	l.Debug("JSON marshalled")
	return string(jd)
}

func (t *SiteTiming) CSV() string {
	l := log.WithFields(log.Fields{
		"action": "CSV",
		"url":    t.URL,
	})
	l.Debug("Starting")
	s := fmt.Sprintf(`%s,%s,%s,%s,%s,%s`,
		t.URL,
		t.DNSDone.String(),
		t.tlsHandshake.String(),
		t.TCPConnect.String(),
		t.TTFB.String(),
		t.TotalTime.String(),
	)
	l.Debug("JSON marshalled")
	return s
}

func (t *SiteTiming) Output() string {
	l := log.WithFields(log.Fields{
		"action": "Output",
		"url":    t.URL,
	})
	l.Debug("Starting")
	if output == "jsonl" {
		return t.JSON()
	} else if output == "csv" {
		return t.CSV()
	} else {
		l.WithField("output", output).Error("Unknown output format")
		return ""
	}
}

func getSiteTiming(url string) (SiteTiming, error) {
	l := log.WithFields(log.Fields{
		"action": "timeGet",
		"url":    url,
	})
	l.Debug("Starting")
	t := &SiteTiming{
		URL: url,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		l.WithError(err).Error("Error creating request")
		return SiteTiming{}, err
	}
	var start, connect, dns, tlsHandshake time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			l.Debug("DNS Done: %v", time.Since(dns))
			t.DNSDone = time.Since(dns)
		},
		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			l.Debug("TLS Handshake: %v", time.Since(tlsHandshake))
			t.tlsHandshake = time.Since(tlsHandshake)
		},
		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			l.Debug("Connect: %v", time.Since(connect))
			t.TCPConnect = time.Since(connect)
		},
		GotFirstResponseByte: func() {
			l.Debug("TTFB: %v", time.Since(start))
			t.TTFB = time.Since(start)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		l.WithError(err).Error("Error making request")
		return SiteTiming{}, err
	}
	t.TotalTime = time.Since(start)
	l.Debug("Total time: %v\n", time.Since(start))
	return *t, nil
}

func getSiteWorker(url <-chan string, done chan<- SiteTiming) {
	for url := range url {
		l := log.WithFields(log.Fields{
			"action": "timeGet",
			"url":    url,
		})
		l.Debug("Starting")
		u := url
		if strings.TrimSpace(u) == "" {
			l.Debug("URL is empty")
			done <- SiteTiming{}
			return
		}
		t, err := getSiteTiming(u)
		if err != nil {
			l.WithError(err).Error("Error making request")
			done <- SiteTiming{}
		}
		done <- t
	}
}

func init() {
	flag.StringVar(&testUrl, "url", "", "url to test")
	flag.StringVar(&output, "out", "jsonl", "output format. (jsonl, csv)")
	flag.StringVar(&logLevel, "log", "info", "log level. (debug,info,warn,error,fatal,color,nocolor,json)")
	flag.IntVar(&concurrent, "concurrent", 10, "concurrent tests to run")
	flag.IntVar(&total, "total", 10, "total tests to run")
	flag.Parse()
	lvl, perr := log.ParseLevel(logLevel)
	if perr != nil {
		log.WithError(perr).Fatal("Invalid log level")
	} else {
		log.SetLevel(lvl)
	}
	if testUrl == "" {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	l := log.WithFields(log.Fields{
		"action": "main",
	})
	l.Debug("Starting")
	var jobs = make(chan string, total)
	var results = make(chan SiteTiming, total)
	l.Debug("Starting workers")
	for i := 0; i < concurrent; i++ {
		l.Debugf("Starting worker %d", i)
		go getSiteWorker(jobs, results)
	}
	for i := 0; i < total; i++ {
		jobs <- testUrl
	}
	close(jobs)
	for i := 0; i < total; i++ {
		r := <-results
		fmt.Println(r.Output())
	}
	close(results)
}
