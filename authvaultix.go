// ================================================
// 🔹 authvaultix Go SDK (HTTPS + JSON + HWID Support)
// ================================================
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const baseURL = "https://api.authvaultix.com/api/1.2/"

// Core Structs
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Subscription struct {
	Subscription string `json:"subscription"`
	Key          string `json:"key"`
	Expiry       string `json:"expiry"`
	TimeLeft     int64  `json:"timeleft"`
}

type UserInfo struct {
	Username      string         `json:"username"`
	IP            string         `json:"ip"`
	HWID          string         `json:"hwid"`
	CreateDate    string         `json:"createdate"`
	LastLogin     string         `json:"lastlogin"`
	Subscriptions []Subscription `json:"subscriptions"`
}

type ApiResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	SessionID string    `json:"sessionid"`
	AppInfo   AppInfo   `json:"appinfo"`
	Info      *UserInfo `json:"info"`
}

// ================================================
// 🔹 Main authvaultix Client
// ================================================
type Authvaultix struct {
	Name      string
	OwnerID   string
	Secret    string
	Version   string
	SessionID string
}

// Set credentials
func (a *Authvaultix) Api(name, ownerid, secret, version string) {
	a.Name = name
	a.OwnerID = ownerid
	a.Secret = secret
	a.Version = version
}

// Send HTTPS Request
func sendRequest(payload map[string]string) (*ApiResponse, error) {
	form := ""
	for k, v := range payload {
		form += fmt.Sprintf("%s=%s&", k, v)
	}
	form = strings.TrimSuffix(form, "&")

	resp, err := http.Post(baseURL, "application/x-www-form-urlencoded", bytes.NewBufferString(form))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result ApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON parse error: %v\nRaw: %s", err, string(body))
	}

	return &result, nil
}

// Get HWID (Windows SID)
func getHWID() string {
	cmd := exec.Command("powershell", "-Command", "[System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value")
	out, err := cmd.Output()
	if err != nil {
		return "UNKNOWN_HWID"
	}
	return strings.TrimSpace(string(out))
}

// ================================================
// 🔹 Init
// ================================================
func (a *Authvaultix) Init() {
	fmt.Println("Connecting...")

	payload := map[string]string{
		"type":    "init",
		"name":    a.Name,
		"ownerid": a.OwnerID,
		"ver":     a.Version,
		"secret":  a.Secret,
	}

	resp, err := sendRequest(payload)
	if err != nil {
		fmt.Println("❌ Init Error:", err)
		os.Exit(1)
	}

	if resp.Success {
		a.SessionID = resp.SessionID
		fmt.Println("✅ Initialized Successfully!")
	} else {
		fmt.Println("❌ Init Failed:", resp.Message)
		os.Exit(1)
	}
}

// ================================================
// 🔹 Login
// ================================================
func (a *Authvaultix) Login(username, password string) {
	hwid := getHWID()
	payload := map[string]string{
		"type":      "login",
		"sessionid": a.SessionID,
		"username":  username,
		"pass":      password,
		"hwid":      hwid,
		"name":      a.Name,
		"ownerid":   a.OwnerID,
	}

	resp, err := sendRequest(payload)
	if err != nil {
		fmt.Println("❌ Login Error:", err)
		return
	}

	if resp.Success {
		fmt.Println("✅ Logged in!")
		printUserInfo(resp.Info)
	} else {
		fmt.Println("❌ Login Failed:", resp.Message)
	}
}

// ================================================
// 🔹 Register
// ================================================
func (a *Authvaultix) Register(username, password, license string) {
	hwid := getHWID()
	payload := map[string]string{
		"type":      "register",
		"sessionid": a.SessionID,
		"username":  username,
		"pass":      password,
		"key":       license,
		"hwid":      hwid,
		"name":      a.Name,
		"ownerid":   a.OwnerID,
	}

	resp, err := sendRequest(payload)
	if err != nil {
		fmt.Println("❌ Register Error:", err)
		return
	}

	if resp.Success {
		fmt.Println("✅ Registered Successfully!")
		printUserInfo(resp.Info)
	} else {
		fmt.Println("❌ Register Failed:", resp.Message)
	}
}

// ================================================
// 🔹 License Login
// ================================================
func (a *Authvaultix) License(license string) {
	hwid := getHWID()
	payload := map[string]string{
		"type":      "license",
		"sessionid": a.SessionID,
		"key":       license,
		"hwid":      hwid,
		"name":      a.Name,
		"ownerid":   a.OwnerID,
	}

	resp, err := sendRequest(payload)
	if err != nil {
		fmt.Println("❌ License Error:", err)
		return
	}

	if resp.Success {
		fmt.Println("✅ License Login Successful!")
		printUserInfo(resp.Info)
	} else {
		fmt.Println("❌ License Login Failed:", resp.Message)
	}
}

// ================================================
// 🔹 Display User Info
// ================================================
func printUserInfo(user *UserInfo) {
	if user == nil {
		return
	}
	fmt.Println("\n👤 User Info:")
	fmt.Println(" Username:", user.Username)
	if user.IP != "" {
		fmt.Println(" IP:", user.IP)
	}
	if user.HWID != "" {
		fmt.Println(" HWID:", user.HWID)
	}
	for _, sub := range user.Subscriptions {
		fmt.Printf(" → %s | Expiry: %d | Left: %d\n", sub.Subscription, sub.Expiry, sub.TimeLeft)
	}
}

// ================================================
