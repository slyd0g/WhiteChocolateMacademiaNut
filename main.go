package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"golang.org/x/net/websocket"
)

// DebugData is JSON structure returned by Chrome
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

// WebsocketResponseRoot is the raw response from Chrome websocket
type WebsocketResponseRoot struct {
	ID     int                     `json:"id"`
	Result WebsocketResponseNested `json:"result"`
}

// WebsocketResponseNested is the object within the the raw response from Chrome websocket
type WebsocketResponseNested struct {
	Cookies []Cookie `json:"cookies"`
}

// Cookie is JSON structure returned by Chrome websocket
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

// GetDebugData interacts with Chrome debug port to obtain the JSON response of open tabs/installed extensions
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

// PrintDebugData takes the JSON response from Google Chrome and prints open tabs and extensions
func PrintDebugData(debugList []DebugData) {
	for _, value := range debugList {
		fmt.Printf("Title: %s\n", value.Title)
		fmt.Printf("Type: %s\n", value.PageType)
		fmt.Printf("URL: %s\n\n", value.URL)
	}
}

// DumpCookies interacts with the webSocketDebuggerUrl to obtain Chrome cookies
func DumpCookies(debugList []DebugData, format string) {

	// Obtain WebSocketDebuggerURL from DebugData list
	var websocketURL = debugList[0].WebSocketDebuggerURL

	// Connect to websocket
	ws, err := websocket.Dial(websocketURL, "", "http://localhost/")
	if err != nil {
		log.Fatal(err)
	}

	// Send message to websocket
	var message = "{\"id\": 1, \"method\":\"Network.getAllCookies\"}"
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

	// Print cookies in raw format as returned by Chrome websocket
	if format == "raw" {
		fmt.Printf("%s\n", rawResponse)
		os.Exit(0)
	}

	// Print cookies in JSON format with modified expiration date
	// Additionally, only include name, value, domain, path, and expirationDate
	if format == "modified" {
		lightCookieList := []LightCookie{}

		for _, value := range websocketResponseRoot.Result.Cookies {
			// Turns Cookie into LightCookie with modified expires field
			var lightCookie LightCookie

			lightCookie.Name = value.Name
			lightCookie.Value = value.Value
			lightCookie.Domain = value.Domain
			lightCookie.Path = value.Path
			lightCookie.Expires = (float64)(time.Now().Unix() + (10 * 365 * 24 * 60 * 60))

			lightCookieList = append(lightCookieList, lightCookie)
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

func main() {

	// Create new parser object
	parser := argparse.NewParser("WhiteChocolateMacademia", "Interact with Google Chrome debug port to view open tabs, installed Chrome extensions, and cookies")

	// Create arguments
	var debugPort *string = parser.String("p", "port", &argparse.Options{Required: true, Help: "Debug port"})
	var dump *string = parser.String("d", "dump", &argparse.Options{Required: true, Help: "{ pages || cookies } - Dump open tabs/extensions or cookies"})
	var format *string = parser.String("f", "format", &argparse.Options{Required: false, Help: "{ raw || human || modified } - Format when dumping cookies"})

	// Parse arguments
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Printf("%s\n", parser.Usage(err))
		os.Exit(1)
	}

	// Enumerate open tabs and installed Chrome extensions
	if *dump == "pages" {
		debugList := GetDebugData(*debugPort)
		PrintDebugData(debugList)
	}

	// Dump cookies
	if *dump == "cookies" {
		debugList := GetDebugData(*debugPort)
		DumpCookies(debugList, *format)
	}
}
