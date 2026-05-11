package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const BaseURL = "https://authvaultix.com/api/1.0/"

// ==========================================
// DTOs
// ==========================================
type DtoBasic struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AppInfo struct {
	Version           string `json:"version"`
	CustomerPanelLink string `json:"customerPanelLink"`
}

type DtoInit struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	SessionID string  `json:"sessionid"`
	AppInfo   AppInfo `json:"appinfo"`
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

type DtoAuth struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Info      *UserInfo `json:"info"`
	SessionID string    `json:"sessionid"`
}

type DtoData struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Contents string `json:"contents"`
}

type DtoVar struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Response string `json:"response"`
}

type OnlineUser struct {
	Credential string `json:"credential"`
}

type DtoOnline struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Users   []OnlineUser `json:"users"`
}

type DtoChat struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	Code             int    `json:"code"`
	RemainingSeconds int    `json:"remaining_seconds"`
	MutedUntil       string `json:"muted_until"`
	RemainingHuman   string `json:"remaining_human"`
}

type ChatMessage struct {
	Author    string `json:"author"`
	Role      string `json:"role"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type DtoChatHistory struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
	Messages []ChatMessage `json:"messages"`
}

type UpgradeUser struct {
	Name string `json:"name"`
}

type DtoUpgrade struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Users   []UpgradeUser `json:"users"`
}

// ==========================================
// NetworkAgent
// ==========================================
type NetworkAgent struct{}

func (na *NetworkAgent) Post(url string, payload map[string]string) ([]byte, error) {
	form := ""
	for k, v := range payload {
		form += fmt.Sprintf("%s=%s&", k, v)
	}
	form = strings.TrimSuffix(form, "&")

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(form))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "AuthVaultixClient/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("You're connecting too fast, slow down.")
	}

	return io.ReadAll(resp.Body)
}

// ==========================================
// PayloadBuilder
// ==========================================
type PayloadBuilder struct {
	Payload map[string]string
}

func NewPayloadBuilder(actionType string) *PayloadBuilder {
	return &PayloadBuilder{
		Payload: map[string]string{"type": actionType},
	}
}

func (pb *PayloadBuilder) WithContext(appName, ownerId, sessionId string) *PayloadBuilder {
	pb.Payload["name"] = appName
	pb.Payload["ownerid"] = ownerId
	if sessionId != "" {
		pb.Payload["sessionid"] = sessionId
	}
	return pb
}

func (pb *PayloadBuilder) WithValue(key, value string) *PayloadBuilder {
	pb.Payload[key] = value
	return pb
}

func (pb *PayloadBuilder) Compile() map[string]string {
	return pb.Payload
}

// ==========================================
// HardwareIdentifier
// ==========================================
func GetHWID() string {
	cmd := exec.Command("powershell", "-Command", "[System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value")
	out, err := cmd.Output()
	if err != nil {
		return "UNKNOWN_HWID"
	}
	return strings.TrimSpace(string(out))
}

// ==========================================
// AuthVaultixCore
// ==========================================
type AuthVaultixCore struct {
	AppName     string
	OwnerID     string
	Secret      string
	Version     string
	SessionID   string
	Initialized bool
	CurrentUser *UserInfo
	Agent       *NetworkAgent
}

func NewAuthVaultixCore(appName, ownerId, secret, version string) *AuthVaultixCore {
	if appName == "" || ownerId == "" || secret == "" || version == "" {
		fmt.Println("Application not setup correctly.")
		os.Exit(1)
	}
	return &AuthVaultixCore{
		AppName: appName,
		OwnerID: ownerId,
		Secret:  secret,
		Version: version,
		Agent:   &NetworkAgent{},
	}
}

func (c *AuthVaultixCore) EnsureReady() {
	if !c.Initialized {
		fmt.Println("SDK not initialized. Call Init() before using any API.")
		os.Exit(1)
	}
}

func (c *AuthVaultixCore) InitializeContext() bool {
	if c.Initialized {
		return true
	}

	payload := NewPayloadBuilder("init").
		WithValue("ver", c.Version).
		WithValue("name", c.AppName).
		WithValue("ownerid", c.OwnerID).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		fmt.Println("Init Error:", err)
		os.Exit(1)
	}

	var dto DtoInit
	if err := json.Unmarshal(respBytes, &dto); err != nil {
		fmt.Println("Invalid server response")
		os.Exit(1)
	}

	if !dto.Success {
		fmt.Println("Init Failed:", dto.Message)
		os.Exit(1)
	}

	c.SessionID = dto.SessionID
	c.Initialized = true
	fmt.Println("Initialized Successfully! Session ID:", c.SessionID)
	return true
}

func (c *AuthVaultixCore) AuthenticateUser(username, password string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("login").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("username", username).
		WithValue("pass", password).
		WithValue("hwid", GetHWID()).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		fmt.Println("Login Error:", err)
		return false
	}

	var dto DtoAuth
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Login Failed:", dto.Message)
		return false
	}

	c.CurrentUser = dto.Info
	if dto.SessionID != "" {
		c.SessionID = dto.SessionID
	}
	fmt.Println("Logged in!")
	PrintUserInfo(c.CurrentUser)
	return true
}

func (c *AuthVaultixCore) ValidateSession() bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("check").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		fmt.Println("Session Valid!")
	} else {
		fmt.Println("Session Invalid:", dto.Message)
	}
	return dto.Success
}

func (c *AuthVaultixCore) RegisterAccount(username, password, licenseKey, email string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("register").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("username", username).
		WithValue("pass", password).
		WithValue("key", licenseKey).
		WithValue("email", email).
		WithValue("hwid", GetHWID()).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		fmt.Println("Register Error:", err)
		return false
	}

	var dto DtoAuth
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Register Failed:", dto.Message)
		return false
	}

	c.CurrentUser = dto.Info
	if dto.SessionID != "" {
		c.SessionID = dto.SessionID
	}
	fmt.Println("Registered Successfully!")
	PrintUserInfo(c.CurrentUser)
	return true
}

func (c *AuthVaultixCore) LicenseAccess(licenseKey string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("license").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("key", licenseKey).
		WithValue("hwid", GetHWID()).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		fmt.Println("License Login Error:", err)
		return false
	}

	var dto DtoAuth
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("License Login Failed:", dto.Message)
		return false
	}

	c.CurrentUser = dto.Info
	if dto.SessionID != "" {
		c.SessionID = dto.SessionID
	}
	fmt.Println("License Login Successful!")
	PrintUserInfo(c.CurrentUser)
	return true
}

func (c *AuthVaultixCore) SendLog(message string) bool {
	c.EnsureReady()
	currentUser, _ := user.Current()
	pcuser := "Unknown"
	if currentUser != nil {
		pcuser = currentUser.Username
	}

	payload := NewPayloadBuilder("log").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("message", message).
		WithValue("pcuser", pcuser).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Log Failed:", dto.Message)
	}
	return dto.Success
}

func (c *AuthVaultixCore) RetrieveFile(fileId string) []byte {
	c.EnsureReady()
	payload := NewPayloadBuilder("file").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("fileid", fileId).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return nil
	}
	var dto DtoData
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Download Failed:", dto.Message)
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(dto.Contents)
	if err != nil {
		fmt.Println("Base64 Decode Error")
		return nil
	}
	fmt.Println("Download successful")
	return decoded
}

func (c *AuthVaultixCore) GetOnlineClients() []OnlineUser {
	c.EnsureReady()
	payload := NewPayloadBuilder("fetchonline").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return nil
	}
	var dto DtoOnline
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Fetch Online Failed:", dto.Message)
		return nil
	}
	return dto.Users
}

func (c *AuthVaultixCore) EnforceBan(reason string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("ban").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("reason", reason).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		fmt.Println("Banned successfully")
	} else {
		fmt.Println("Ban Failed:", dto.Message)
	}
	return dto.Success
}

func (c *AuthVaultixCore) TerminateSession() {
	c.EnsureReady()
	payload := NewPayloadBuilder("logout").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err == nil {
		var dto DtoBasic
		json.Unmarshal(respBytes, &dto)
		if dto.Success {
			c.SessionID = ""
			c.Initialized = false
			fmt.Println("Logged out successfully")
			return
		}
	}
	fmt.Println("Logout Error")
}

func (c *AuthVaultixCore) UpdateUsername(newUsername string) {
	c.EnsureReady()
	payload := NewPayloadBuilder("changeusername").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("newUsername", newUsername).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err == nil {
		var dto DtoBasic
		json.Unmarshal(respBytes, &dto)
		if dto.Success {
			c.SessionID = ""
			c.Initialized = false
			fmt.Println("Username changed successfully. Please login again.")
			return
		}
	}
	fmt.Println("Change Username Error")
}

func (c *AuthVaultixCore) VerifyBlacklist() bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("checkblacklist").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("hwid", GetHWID()).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Client is blacklisted:", dto.Message)
		return false
	}
	return true
}

func (c *AuthVaultixCore) TriggerPasswordReset(username, email string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("forgot").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("username", username).
		WithValue("email", email).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		fmt.Println("Reset email sent successfully")
	} else {
		fmt.Println("Forgot Password Failed:", dto.Message)
	}
	return dto.Success
}

func (c *AuthVaultixCore) ApplyUpgrade(username, licenseKey string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("upgrade").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("username", username).
		WithValue("key", licenseKey).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoUpgrade
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		fmt.Println("Upgrade successful")
	} else {
		fmt.Println("Upgrade Failed:", dto.Message)
	}
	return dto.Success
}

func (c *AuthVaultixCore) FetchGlobalVariable(varId string) string {
	c.EnsureReady()
	payload := NewPayloadBuilder("var").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("varid", varId).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return ""
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		return dto.Message
	}
	fmt.Println("Fetch Global Var Failed:", dto.Message)
	return ""
}

func (c *AuthVaultixCore) FetchUserVariable(varName string) string {
	c.EnsureReady()
	payload := NewPayloadBuilder("getvar").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("var", varName).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return ""
	}
	var dto DtoVar
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		return dto.Response
	}
	fmt.Println("Fetch User Var Failed:", dto.Message)
	return ""
}

func (c *AuthVaultixCore) UpdateUserVariable(varName, value string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("setvar").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("var", varName).
		WithValue("data", value).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoBasic
	json.Unmarshal(respBytes, &dto)
	if !dto.Success {
		fmt.Println("Set User Var Failed:", dto.Message)
	}
	return dto.Success
}

func (c *AuthVaultixCore) TransmitChatMessage(message, channel string) bool {
	c.EnsureReady()
	payload := NewPayloadBuilder("chatsend").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("message", message).
		WithValue("channel", channel).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return false
	}
	var dto DtoChat
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		fmt.Println("Message sent.")
		return true
	}
	if dto.Code == 403 && dto.RemainingSeconds > 0 {
		fmt.Printf("Muted till %s (wait %s)\n", dto.MutedUntil, dto.RemainingHuman)
	} else {
		fmt.Println("Chat Send Failed:", dto.Message)
	}
	return false
}

func (c *AuthVaultixCore) RetrieveChatHistory(channel string) []ChatMessage {
	c.EnsureReady()
	payload := NewPayloadBuilder("chatfetch").
		WithContext(c.AppName, c.OwnerID, c.SessionID).
		WithValue("channel", channel).
		Compile()

	respBytes, err := c.Agent.Post(BaseURL, payload)
	if err != nil {
		return nil
	}
	var dto DtoChatHistory
	json.Unmarshal(respBytes, &dto)
	if dto.Success {
		return dto.Messages
	}
	fmt.Println("Chat Fetch Failed:", dto.Message)
	return nil
}

// ==========================================
// AuthVaultixClient
// ==========================================
type AuthVaultix struct {
	Core *AuthVaultixCore
}

func NewAuthVaultix(appName, ownerId, secret, version string) *AuthVaultix {
	return &AuthVaultix{
		Core: NewAuthVaultixCore(appName, ownerId, secret, version),
	}
}

func (a *AuthVaultix) Init() bool                                                 { return a.Core.InitializeContext() }
func (a *AuthVaultix) Login(username, password string) bool                       { return a.Core.AuthenticateUser(username, password) }
func (a *AuthVaultix) Check() bool                                                { return a.Core.ValidateSession() }
func (a *AuthVaultix) Register(username, password, licenseKey, email string) bool { return a.Core.RegisterAccount(username, password, licenseKey, email) }
func (a *AuthVaultix) LicenseLogin(licenseKey string) bool                        { return a.Core.LicenseAccess(licenseKey) }
func (a *AuthVaultix) Log(message string) bool                                    { return a.Core.SendLog(message) }
func (a *AuthVaultix) Download(fileId string) []byte                              { return a.Core.RetrieveFile(fileId) }
func (a *AuthVaultix) FetchOnline() []OnlineUser                                  { return a.Core.GetOnlineClients() }
func (a *AuthVaultix) Ban(reason string) bool                                     { return a.Core.EnforceBan(reason) }
func (a *AuthVaultix) Logout()                                                    { a.Core.TerminateSession() }
func (a *AuthVaultix) ChangeUsername(newUsername string)                          { a.Core.UpdateUsername(newUsername) }
func (a *AuthVaultix) CheckBlacklist() bool                                       { return a.Core.VerifyBlacklist() }
func (a *AuthVaultix) Upgrade(username, licenseKey string) bool                   { return a.Core.ApplyUpgrade(username, licenseKey) }
func (a *AuthVaultix) ForgotPassword(username, email string) bool                 { return a.Core.TriggerPasswordReset(username, email) }
func (a *AuthVaultix) GetGlobalVar(varId string) string                           { return a.Core.FetchGlobalVariable(varId) }
func (a *AuthVaultix) GetVar(varName string) string                               { return a.Core.FetchUserVariable(varName) }
func (a *AuthVaultix) SetVar(varName, value string) bool                          { return a.Core.UpdateUserVariable(varName, value) }
func (a *AuthVaultix) ChatSend(message, channel string) bool                      { return a.Core.TransmitChatMessage(message, channel) }
func (a *AuthVaultix) ChatFetch(channel string) []ChatMessage                     { return a.Core.RetrieveChatHistory(channel) }

// ==========================================
// Helper
// ==========================================
func formatUnix(tsStr string) string {
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return tsStr
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func formatTimeLeft(seconds int64) string {
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%dd %dh %dm", d, h, m)
}

func PrintUserInfo(user *UserInfo) {
	if user == nil {
		return
	}
	fmt.Println("\n=== User Data ===")
	fmt.Println("Username:", user.Username)
	if user.IP != "" {
		fmt.Println("IP:", user.IP)
	}
	if user.HWID != "" {
		fmt.Println("HWID:", user.HWID)
	}
	if user.CreateDate != "" {
		fmt.Println("Created:", formatUnix(user.CreateDate))
	}
	if user.LastLogin != "" {
		fmt.Println("Last Login:", formatUnix(user.LastLogin))
	}

	if len(user.Subscriptions) > 0 {
		fmt.Println("\nSubscriptions:")
		for i, sub := range user.Subscriptions {
			fmt.Printf("[%d] %s | Expiry: %s | Timeleft: %s\n", i+1, sub.Subscription, formatUnix(sub.Expiry), formatTimeLeft(sub.TimeLeft))
		}
	}
	fmt.Println()
}
