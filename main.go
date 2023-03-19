package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"golang.org/x/net/websocket"
)

// DebugData is JSON structure returned by Chromium (/json)
type DebugData struct {
	Description          string `json:"description"`
	DevtoolsFrontendURL  string `json:"devtoolsFrontendUrl"`
	FaviconURL           string `json:"faviconUrl"`
	ID                   string `json:"id"`
	Title                string `json:"title"`
	PageType             string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// DebugData is a JSON structure returned by Chromium (/json/version)
type DebugDataVersion struct {
	Browser              string `json:"Browser"`
	ProtocolVersion      string `json:"Protocol-Version"`
	UserAgent            string `json:"User-Agent"`
	V8Version            string `json:"V8-Version"`
	WebkitVersion        string `json:"WebKit-Version"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// WebsocketResponseRoot is the raw response from Chromium websocket
type WebsocketResponseRoot struct {
	ID     int                     `json:"id"`
	Result WebsocketResponseNested `json:"result"`
}

// WebsocketResponseNested is the object within the the raw response from Chromium websocket
type WebsocketResponseNested struct {
	Cookies []Cookie `json:"cookies"`
}

// Cookie is JSON structure returned by Chromium websocket
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	Size     int     `json:"size"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	Session  bool    `json:"session"`
	SameSite string  `json:"sameSite"`
	Priority string  `json:"priority"`
}

// LightCookie is a JSON structure for the cookie with only the name, value, domain, path, and (modified) expires fields
type LightCookie struct {
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Domain  string  `json:"domain"`
	Path    string  `json:"path"`
	Expires float64 `json:"expirationDate"`
}

// GetDebugData interacts with Chromium debug port to obtain the JSON response of open tabs/installed extensions
func GetDebugData(debugPort string) []DebugData {

	// Create debugURL from user input
	var debugURL = "http://localhost:" + debugPort + "/json"

	// Make GET request
	resp, err := http.Get(debugURL)
	if err != nil {
		log.Fatalln(err)
	}

	// Read GET response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Unmarshal JSON response
	var debugList []DebugData
	err = json.Unmarshal(body, &debugList)
	if err != nil {
		log.Fatalln(err)
	}

	return debugList
}

// GetDebugDataVersion interacts with the Chromium debug port to obtain the JSON response from /json/version
func GetDebugDataVersion(debugPort string) DebugDataVersion {
	// Create debugURL from user input
	var debugURL = "http://localhost:" + debugPort + "/json"

	// Make GET request
	resp, err := http.Get(debugURL)
	if err != nil {
		log.Fatalln(err)
	}

	// Read GET response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Unmarshal JSON response
	var debugVersionList DebugDataVersion
	err = json.Unmarshal(body, &debugVersionList)
	if err != nil {
		log.Fatalln(err)
	}

	return debugVersionList
}

// PrintDebugData takes the JSON response from Chromium and prints open tabs and  installed extensions
func PrintDebugData(debugList []DebugData, grep string) {

	// Check length of grep to see if filtering was requested
	var grepFlag = false

	if len(grep) > 0 {
		grepFlag = true
	}

	for _, value := range debugList {
		if grepFlag {
			if strings.Contains(value.Title, grep) || strings.Contains(value.URL, grep) {
				fmt.Printf("Title: %s\n", value.Title)
				fmt.Printf("Type: %s\n", value.PageType)
				fmt.Printf("URL: %s\n", value.URL)
				fmt.Printf("WebSocket Debugger URL: %s\n\n", value.WebSocketDebuggerURL)
			}
		} else {
			fmt.Printf("Title: %s\n", value.Title)
			fmt.Printf("Type: %s\n", value.PageType)
			fmt.Printf("URL: %s\n", value.URL)
			fmt.Printf("WebSocket Debugger URL: %s\n\n", value.WebSocketDebuggerURL)
		}

	}
}

// DumpCookies interacts with the webSocketDebuggerUrl to obtain Chromium cookies
func DumpCookies(debugVersionData DebugDataVersion, format string, grep string) {

	// Check length of grep to see if filtering was requested
	var grepFlag = false

	if len(grep) > 0 {
		grepFlag = true
	}

	// Obtain WebSocketDebuggerURL from DebugData list
	var websocketURL = debugVersionData.WebSocketDebuggerURL

	// Connect to websocket
	ws, err := websocket.Dial(websocketURL, "", "http://localhost/")
	if err != nil {
		log.Fatal(err)
	}

	// Send message to websocket
	var message = "{\"id\": 1, \"method\":\"Storage.getCookies\"}"
	websocket.Message.Send(ws, message)

	// Get cookies from websocket
	var rawResponse []byte
	websocket.Message.Receive(ws, &rawResponse)

	// Unmarshal JSON response
	var websocketResponseRoot WebsocketResponseRoot
	err = json.Unmarshal(rawResponse, &websocketResponseRoot)
	if err != nil {
		log.Fatalln(err)
	}

	// Print cookies in raw format as returned by Chromium websocket
	if format == "raw" {
		fmt.Printf("%s\n", rawResponse)
		os.Exit(0)
	}

	// Print cookies in JSON format with modified expiration date
	// Additionally, only include name, value, domain, path, and expirationDate
	if format == "modified" {
		lightCookieList := []LightCookie{}

		for _, value := range websocketResponseRoot.Result.Cookies {
			if grepFlag {
				if strings.Contains(value.Name, grep) || strings.Contains(value.Domain, grep) {
					// Turns Cookie into LightCookie with modified expires field
					var lightCookie LightCookie

					lightCookie.Name = value.Name
					lightCookie.Value = value.Value
					lightCookie.Domain = value.Domain
					lightCookie.Path = value.Path
					lightCookie.Expires = (float64)(time.Now().Unix() + (10 * 365 * 24 * 60 * 60))

					lightCookieList = append(lightCookieList, lightCookie)
				}
			} else {
				// Turns Cookie into LightCookie with modified expires field
				var lightCookie LightCookie

				lightCookie.Name = value.Name
				lightCookie.Value = value.Value
				lightCookie.Domain = value.Domain
				lightCookie.Path = value.Path
				lightCookie.Expires = (float64)(time.Now().Unix() + (10 * 365 * 24 * 60 * 60))

				lightCookieList = append(lightCookieList, lightCookie)
			}

		}

		lightCookieJSON, err := json.Marshal(lightCookieList)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%s\n", lightCookieJSON)
		os.Exit(0)
	}

	// Default to printing cookies in human format
	for _, value := range websocketResponseRoot.Result.Cookies {
		if grepFlag {
			if strings.Contains(value.Name, grep) || strings.Contains(value.Domain, grep) {
				fmt.Printf("name: %s\n", value.Name)
				fmt.Printf("value: %s\n", value.Value)
				fmt.Printf("domain: %s\n", value.Domain)
				fmt.Printf("path: %s\n", value.Path)
				fmt.Printf("expires: %f\n", value.Expires)
				fmt.Printf("size: %d\n", value.Size)
				fmt.Printf("httpOnly: %t\n", value.HTTPOnly)
				fmt.Printf("secure: %t\n", value.Secure)
				fmt.Printf("session: %t\n", value.Session)
				fmt.Printf("sameSite: %s\n", value.SameSite)
				fmt.Printf("priority: %s\n\n", value.Priority)
			}

		} else {
			fmt.Printf("name: %s\n", value.Name)
			fmt.Printf("value: %s\n", value.Value)
			fmt.Printf("domain: %s\n", value.Domain)
			fmt.Printf("path: %s\n", value.Path)
			fmt.Printf("expires: %f\n", value.Expires)
			fmt.Printf("size: %d\n", value.Size)
			fmt.Printf("httpOnly: %t\n", value.HTTPOnly)
			fmt.Printf("secure: %t\n", value.Secure)
			fmt.Printf("session: %t\n", value.Session)
			fmt.Printf("sameSite: %s\n", value.SameSite)
			fmt.Printf("priority: %s\n\n", value.Priority)
		}

	}
}

func ClearCookies(debugList []DebugData) {
	var websocketURL = debugList[0].WebSocketDebuggerURL

	// Connect to websocket
	ws, err := websocket.Dial(websocketURL, "", "http://localhost/")
	if err != nil {
		log.Fatal(err)
	}

	// Send message to websocket
	var message = "{\"id\": 1, \"method\": \"Network.clearBrowserCookies\"}"
	websocket.Message.Send(ws, message)

}

func LoadCookies(debugList []DebugData, load string) {
	// Read cookies
	content, err := ioutil.ReadFile(load)
	if err != nil {
		log.Fatal(err)
	}

	var websocketURL = debugList[0].WebSocketDebuggerURL

	// Connect to websocket
	ws, err := websocket.Dial(websocketURL, "", "http://localhost/")
	if err != nil {
		log.Fatal(err)
	}

	// Send message to websocket
	var message = fmt.Sprintf("{\"id\": 1, \"method\":\"Network.setCookies\", \"params\":{\"cookies\":%s}}", content)
	websocket.Message.Send(ws, message)
}

func main() {

	// Create new parser object
	parser := argparse.NewParser("WhiteChocolateMacademia", "Interact with Chromium-based browsers' debug port to view open tabs, installed extensions, and cookies (https://github.com/slyd0g/WhiteChocolateMacademiaNut)")

	// Create arguments
	var debugPort *string = parser.String("p", "port", &argparse.Options{Required: true, Help: "{REQUIRED} - Debug port"})
	var dump *string = parser.String("d", "dump", &argparse.Options{Required: false, Help: "{ pages || cookies } - Dump open tabs/extensions or cookies"})
	var format *string = parser.String("f", "format", &argparse.Options{Required: false, Help: "{ raw || human || modified } - Format when dumping cookies"})
	var grep *string = parser.String("g", "grep", &argparse.Options{Required: false, Help: "Narrow scope of dumping to specific name/domain"})
	var load *string = parser.String("l", "load", &argparse.Options{Required: false, Help: "File name for cookies to load into browser"})
	var clear *string = parser.String("c", "clear", &argparse.Options{Required: false, Help: "Clear cookies before loading new cookies"})
	// Parse arguments
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Printf("%s\n", parser.Usage(err))
		os.Exit(1)
	}

	if *dump != "" {
		// Enumerate open tabs and installed extensions
		if *dump == "pages" {
			debugList := GetDebugData(*debugPort)
			PrintDebugData(debugList, *grep)
		}

		// Dump cookies
		if *dump == "cookies" {
			debugList := GetDebugDataVersion(*debugPort)
			DumpCookies(debugList, *format, *grep)
		}
	}

	if *clear != "" {
		debugList := GetDebugData(*debugPort)
		ClearCookies(debugList)
	}

	if *load != "" {
		debugList := GetDebugData(*debugPort)
		LoadCookies(debugList, *load)
	}
}
