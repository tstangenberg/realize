package realize

import (
	"fmt"
	"go/build"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	RPrefix  = "realize"
	RVersion = "2.0"
	RExt     = ".yaml"
	RFile    = RPrefix + RExt
	RDir     = "." + RPrefix
	RExtWin  = ".exe"
)

type (
	LogWriter struct{}

	Realize struct {
		Settings Settings `yaml:"settings" json:"settings"`
		Server   Server   `yaml:"server" json:"server"`
		Schema   `yaml:",inline"`
		sync     chan string
		exit     chan os.Signal
		Err      Func `yaml:"-"`
		After    Func `yaml:"-"`
		Before   Func `yaml:"-"`
		Change   Func `yaml:"-"`
		Reload   Func `yaml:"-"`
	}

	Context struct {
		Path    string
		Project Project
		Stop    <-chan bool
		Watcher FileWatcher
		Event   fsnotify.Event
	}

	Func func(Context)
)

// init check
func init() {
	// custom log
	log.SetFlags(0)
	log.SetOutput(LogWriter{})
	if build.Default.GOPATH == "" {
		log.Fatal("$GOPATH isn't set properly")
	}
	if err := os.Setenv("GOBIN", filepath.Join(build.Default.GOPATH, "bin")); err != nil {
		log.Fatal(err)
	}
}

// Stop realize workflow
func (r *Realize) Stop() {
	close(r.exit)
}

// Run realize workflow
func (r *Realize) Start() {
	r.exit = make(chan os.Signal, 2)
	signal.Notify(r.exit, os.Interrupt, syscall.SIGTERM)
	for k := range r.Schema.Projects {
		r.Schema.Projects[k].parent = r
		go r.Schema.Projects[k].Watch(r.exit)
	}
	for {
		select {
		case <-r.exit:
			return
		}
	}
}

// Prefix a given string with tool name
func (r *Realize) Prefix(input string) string {
	if len(input) > 0 {
		return fmt.Sprint(Yellow.Bold("["), strings.ToUpper(RPrefix), Yellow.Bold("]"), " : ", input)
	}
	return input
}

// Rewrite the layout of the log timestamp
func (w LogWriter) Write(bytes []byte) (int, error) {
	if len(bytes) > 0 {
		return fmt.Fprint(Output, Yellow.Regular("["), time.Now().Format("15:04:05"), Yellow.Regular("]"), string(bytes))
	}
	return 0, nil
}
