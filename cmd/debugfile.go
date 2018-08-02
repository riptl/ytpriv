package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"fmt"
	"bufio"
	"encoding/json"
	"github.com/terorie/yt-mango/net"
	c "github.com/terorie/yt-mango/pretty"
	"strings"
	"encoding/base64"
)

var DebugFile = cobra.Command{
	Use: "debug-file",
	Short: "Print a JSON-like debug file",
	Long: "Prints a human-readable version of the debug file\n" +
		"generated with flag `--debug-file`",
	Run: debugFileCmd,
	Args: cobra.ExactArgs(1),
}

func debugFileCmd(_ *cobra.Command, args []string) {
	fileName := args[0]

	file, err := os.Open(fileName)
	if err != nil { debugFileAbort(err) }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scannedOne := false
	for scanner.Scan() {
		scannedOne = true
		line := scanner.Text()
		var obj net.LogObj
		json.Unmarshal([]byte(line), &obj)
		fmt.Printf(c.Add(c.HWHITE, c.BOLD).E("%s %s%s:\n"), obj.Request.Method, obj.Request.Host, obj.Request.Path)
		fmt.Printf(" |-+- Request headers:\n")
		for key, valueArr := range obj.Request.Header {
			for _, value := range valueArr {
				fmt.Printf("   |--- %s: %s\n", key, value)
			}
		}
		fmt.Println(" | Response:")
		if obj.Err != "" {
			fmt.Println(" |--- Error:", obj.Err)
		} else {
			fmt.Printf(" |--- %s (%d)\n", obj.Response.Status, obj.Response.StatusCode)
			fmt.Println(" |-+- Response headers:")
			for key, valueArr := range obj.Response.Header {
				for _, value := range valueArr {
					fmt.Printf("   |- %s: %s\n", key, value)
				}
			}
			contentType := obj.Response.Header.Get("content-type")
			switch {
			case strings.HasPrefix(contentType, "application/json"):
				fmt.Println(" |-+- JSON data:")
				debugFileDumpBody(obj.Response.Body)
			case strings.HasPrefix(contentType, "text/html"):
				fmt.Println(" |-+- HTML document:")
				debugFileDumpBody(obj.Response.Body)
			default:
				fmt.Println(" |--- Binary data (omitted)")
			}
		}
	}
	if !scannedOne {
		fmt.Fprintln(os.Stderr, "No debug entries")
	}
}

func debugFileDumpBody(bodyB64 string) {
	bodyBytes, err := base64.StdEncoding.DecodeString(bodyB64)
	if err != nil {
		fmt.Println("   |### Broken base64 data")
		return
	}
	body := string(bodyBytes)
	/*for _, line := range strings.Split(body, "\n") {
		fmt.Println("   |-", line)
	}*/
	fmt.Println(body)
	fmt.Print("\n\n\n")
}

func debugFileAbort(err error) {
	fmt.Fprintln(os.Stderr, "Error reading debug file:", err)
	os.Exit(1)
}
