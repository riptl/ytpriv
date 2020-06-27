package main

import (
	"bufio"
	"encoding/json"
	"expvar"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/rpc"
	jsonrpc "github.com/gorilla/rpc/json"

	"github.com/sirupsen/logrus"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/data"
	"github.com/terorie/ytwrk/net"
	"github.com/valyala/fasthttp"
)

var basePath string

func main() {
	bind := flag.String("bind", ":8080", "RPC Bind")
	flag.StringVar(&basePath, "path", "./livechat-dump", "Livechat folder")
	flag.Parse()
	if err := os.MkdirAll(basePath, 0777); err != nil {
		logrus.WithError(err).Fatal("Failed to create livechat folder")
	}
	s := rpc.NewServer()
	s.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	s.RegisterCodec(jsonrpc.NewCodec(), "application/json;charset=UTF-8")
	daemon := &Daemon{
		active: make(map[string]*Progress),
	}
	if err := s.RegisterService(daemon, ""); err != nil {
		logrus.Fatal(err)
	}
	if err := http.ListenAndServe(*bind, s); err != nil {
		logrus.Fatal(err)
	}
}

type Progress struct {
	Started  time.Time
	Requests expvar.Int
	Messages expvar.Int
}

func (p *Progress) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Requests int64
		Messages int64
		Started  time.Time
	}{
		Requests: p.Requests.Value(),
		Messages: p.Messages.Value(),
		Started:  p.Started,
	})
}

type Daemon struct {
	lock   sync.Mutex
	active map[string]*Progress
}

func (d *Daemon) AddJobs(r *http.Request, jobs *[]string, reply *bool) error {
	// Parse and register
	d.lock.Lock()
	penalty := time.Duration(0)
	for _, job := range *jobs {
		videoID, err := api.GetVideoID(job)
		if err != nil {
			logrus.WithError(err).Error("Add failed")
			continue
		}
		if _, ok := d.active[videoID]; ok {
			continue
		}
		d.active[videoID] = &Progress{
			Started: time.Now(),
		}
		penalty += 3 * time.Second
		go func() {
			time.Sleep(penalty)
			e := logrus.WithField("id", videoID)
			if err := d.dump(videoID); err != nil {
				e.WithError(err).
					Error("Livestream failed")
			} else {
				e.Info("Livestream finished")
			}
		}()
	}
	d.lock.Unlock()

	*reply = true
	return nil
}

func (d *Daemon) Status(r *http.Request, dummy *bool, reply *map[string]json.RawMessage) error {
	*reply = make(map[string]json.RawMessage)
	d.lock.Lock()
	defer d.lock.Unlock()
	for id, prog := range d.active {
		buf, err := prog.MarshalJSON()
		if err != nil {
			panic(err)
		}
		(*reply)[id] = buf
	}
	return nil
}

func (d *Daemon) dump(id string) (err error) {
	d.lock.Lock()
	prog := d.active[id]
	d.lock.Unlock()

	defer func() {
		d.lock.Lock()
		delete(d.active, id)
		d.lock.Unlock()
	}()

	filePath := filepath.Join(basePath, id+".ndjson")
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	wr := bufio.NewWriter(f)
	defer wr.Flush()

	startReq := api.GrabLiveChatStart(id)
	startRes := fasthttp.AcquireResponse()
	err = net.Client.Do(startReq, startRes)
	if err != nil {
		return err
	}
	prog.Requests.Add(1)
	fasthttp.ReleaseRequest(startReq)

	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)

	var out []data.LiveChatMessage
	var cont data.LiveChatContinuation
	cont, err = api.ParseLiveChatStart(&out, startRes)
	if err != nil {
		return err
	}
	fasthttp.ReleaseResponse(startRes)
	for _, msg := range out {
		prog.Messages.Add(1)
		if err := enc.Encode(&msg); err != nil {
			logrus.WithError(err).Fatal("Output failed")
		}
	}

	for cont.Continuation != "" {
		time.Sleep(time.Duration(cont.Timeout) * time.Millisecond)
		pageReq := api.GrabLiveChatContinuation(cont.Continuation)
		pageRes := fasthttp.AcquireResponse()
		err = net.Client.Do(pageReq, pageRes)
		if err != nil {
			return err
		}
		prog.Requests.Add(1)
		fasthttp.ReleaseRequest(pageReq)
		cont, err = api.ParseLiveChatPage(&out, pageRes)
		if err != nil {
			return err
		}
		fasthttp.ReleaseResponse(pageRes)
		for _, msg := range out {
			prog.Messages.Add(1)
			if err := enc.Encode(&msg); err != nil {
				panic(err)
			}
		}
	}

	return nil
}
