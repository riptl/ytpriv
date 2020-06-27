package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var serverURL string

var app = cobra.Command{
	Use:   "livechat-client",
	Short: "Utility to control the livechat daemon",
}

func init() {
	app.AddCommand(&enqueue, &status)
	pf := app.PersistentFlags()
	pf.StringVarP(&serverURL, "connect", "c", "http://localhost:8080/", "Daemon RPC URL")
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

var enqueue = cobra.Command{
	Use:   "enqueue",
	Short: "Reads video IDs from stdin line-by-line and commits them",
	Args:  cobra.NoArgs,
	Run:   runEnqueue,
}

var status = cobra.Command{
	Use: "status",
	Run: runStatus,
}

// FIXME Really ugly
func runEnqueue(_ *cobra.Command, _ []string) {
	var lines []string
	scn := bufio.NewScanner(os.Stdin)
	for scn.Scan() {
		lines = append(lines, scn.Text())
	}
	eerere, _ := json.Marshal(lines)
	rd := strings.NewReader(
		fmt.Sprintf(`{"id":0,"method":"Daemon.AddJobs","params":[%s]}`, string(eerere)),
	)
	res, err := http.Post(serverURL, "application/json", rd)
	if err != nil {
		panic(err)
	}
	var resp struct {
		Error json.RawMessage `json:"error"`
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(buf, &resp); err != nil {
		panic(err)
	}
	res.Body.Close()
}

func runStatus(_ *cobra.Command, _ []string) {
	rd := strings.NewReader(
		`{"id":0,"method":"Daemon.Status","params":[true]}`,
	)
	res, err := http.Post(serverURL, "application/json", rd)
	if err != nil {
		panic(err)
	}
	_, _ = io.Copy(os.Stdout, res.Body)
}
