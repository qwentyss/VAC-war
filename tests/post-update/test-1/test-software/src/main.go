package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
	"unsafe"

	"github.com/jchv/go-webview2"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

type Matrix [4][4]float32

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

func (v Vector3) Dist(other Vector3) float32 {
	return float32(math.Abs(float64(v.X-other.X)) + math.Abs(float64(v.Y-other.Y)) + math.Abs(float64(v.Z-other.Z)))
}

type Vector2 struct {
	X float32
	Y float32
}

type Rectangle struct {
	Top    float32
	Left   float32
	Right  float32
	Bottom float32
}

type Entity struct {
	Health   int32
	Team     int32
	Name     string
	Position Vector2
	Bones    map[string]Vector2
	HeadPos  Vector3
	Distance float32
	Rect     Rectangle
}

type FrameData struct {
	Entities []Entity
}

type Offset struct {
	DwViewMatrix           uintptr `json:"dwViewMatrix"`
	DwLocalPlayerPawn      uintptr `json:"dwLocalPlayerPawn"`
	DwEntityList           uintptr `json:"dwEntityList"`
	DwGameRules            uintptr `json:"dwGameRules"`
	DwBuildNumber          uintptr `json:"dwBuildNumber"`
	M_hPlayerPawn          uintptr `json:"m_hPlayerPawn"`
	M_iHealth              uintptr `json:"m_iHealth"`
	M_lifeState            uintptr `json:"m_lifeState"`
	M_iTeamNum             uintptr `json:"m_iTeamNum"`
	M_vOldOrigin           uintptr `json:"m_vOldOrigin"`
	M_pGameSceneNode       uintptr `json:"m_pGameSceneNode"`
	M_modelState           uintptr `json:"m_modelState"`
	M_boneArray            uintptr `json:"m_boneArray"`
	M_nodeToWorld          uintptr `json:"m_nodeToWorld"`
	M_sSanitizedPlayerName uintptr `json:"m_sSanitizedPlayerName"`

	// C_CSGameRules
	M_bIsValveDS             uintptr `json:"m_bIsValveDS"`
	M_bIsQueuedMatchmaking   uintptr `json:"m_bIsQueuedMatchmaking"`
	M_nQueuedMatchmakingMode uintptr `json:"m_nQueuedMatchmakingMode"`
	M_bHasMatchStarted       uintptr `json:"m_bHasMatchStarted"`
	M_totalRoundsPlayed      uintptr `json:"m_totalRoundsPlayed"`
	M_iRoundTime             uintptr `json:"m_iRoundTime"`
	M_bFreezePeriod          uintptr `json:"m_bFreezePeriod"`
	M_gamePhase              uintptr `json:"m_gamePhase"`
	M_iRoundWinStatus        uintptr `json:"m_iRoundWinStatus"`
}

type Config struct {
	Version             string `json:"version"`
	ESP                 bool   `json:"esp"`
	TeamCheck           bool   `json:"team_check"`
	HeadCircle          bool   `json:"head_circle"`
	SkeletonRendering   bool   `json:"skeleton_rendering"`
	BoxRendering        bool   `json:"box_rendering"`
	NameRendering       bool   `json:"name_rendering"`
	HealthBarRendering  bool   `json:"health_bar_rendering"`
	HealthTextRendering bool   `json:"health_text_rendering"`
	ScreenProtection    bool   `json:"screen_protection"`
	FrameDelay          uint32 `json:"frame_delay"`
	ShowDistance        bool   `json:"show_distance"`
	ShowSnapLines       bool   `json:"show_snap_lines"`
	ShowHeadDot         bool   `json:"show_head_dot"`
	KeyESP              string `json:"key_esp"`
	KeyScreenProtection string `json:"key_screen_protection"`
	KeyOpenGUI          string `json:"key_open_gui"`
}

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	ole32    = windows.NewLazySystemDLL("ole32.dll")

	getSystemMetrics           = user32.NewProc("GetSystemMetrics")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	setWindowDisplayAffinity   = user32.NewProc("SetWindowDisplayAffinity")
	setTextAlign               = gdi32.NewProc("SetTextAlign")
	createFont                 = gdi32.NewProc("CreateFontW")
	createCompatibleDC         = gdi32.NewProc("CreateCompatibleDC")
	createSolidBrush           = gdi32.NewProc("CreateSolidBrush")
	createPen                  = gdi32.NewProc("CreatePen")
	freeConsole                = kernel32.NewProc("FreeConsole")
	shell_NotifyIcon           = shell32.NewProc("Shell_NotifyIconW")
	appendMenu                 = user32.NewProc("AppendMenuW")
	createPopupMenu            = user32.NewProc("CreatePopupMenu")
	destroyMenu                = user32.NewProc("DestroyMenu")
	trackPopupMenu             = user32.NewProc("TrackPopupMenu")
	getCursorPos               = user32.NewProc("GetCursorPos")
	setForegroundWindow        = user32.NewProc("SetForegroundWindow")
	getAsyncKeyState           = user32.NewProc("GetAsyncKeyState")
	getForegroundWindow        = user32.NewProc("GetForegroundWindow")
	getWindowThreadProcessId   = user32.NewProc("GetWindowThreadProcessId")

	allowSetForegroundWindow = user32.NewProc("AllowSetForegroundWindow")
	setWindowPos             = user32.NewProc("SetWindowPos")
	bringWindowToTop         = user32.NewProc("BringWindowToTop")
	setFocus                 = user32.NewProc("SetFocus")
	attachThreadInput        = user32.NewProc("AttachThreadInput")
	getCurrentThreadId       = kernel32.NewProc("GetCurrentThreadId")
	setWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	getWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	getWindow                = user32.NewProc("GetWindow")
	enumChildWindows         = user32.NewProc("EnumChildWindows")
	enumWindows              = user32.NewProc("EnumWindows")
	isWindowVisible          = user32.NewProc("IsWindowVisible")
	showWindowProc           = user32.NewProc("ShowWindow")
	updateWindowProc         = user32.NewProc("UpdateWindow")
	getWindowTextW           = user32.NewProc("GetWindowTextW")
	getClassNameW            = user32.NewProc("GetClassNameW")
)

const (
	ASFW_ANY         = ^uintptr(0)
	HWND_TOPMOST     = ^uintptr(0)
	HWND_NOTOPMOST   = ^uintptr(1)
	SWP_NOSIZE       = 0x0001
	SWP_NOMOVE       = 0x0002
	SWP_NOZORDER     = 0x0004
	SWP_NOACTIVATE   = 0x0010
	SWP_FRAMECHANGED = 0x0020
	SWP_SHOWWINDOW   = 0x0040
	WS_EX_TOPMOST    = 0x00000008
	WS_EX_TOOLWINDOW = 0x00000080
	WS_CAPTION       = 0x00C00000
	WS_SYSMENU       = 0x00080000
	WS_THICKFRAME    = 0x00040000
	WS_MINIMIZEBOX   = 0x00020000
	WS_MAXIMIZEBOX   = 0x00010000
	GW_CHILD         = 5
	GW_HWNDNEXT      = 2
	SW_SHOW          = 5
	SW_HIDE          = 0
	SW_RESTORE       = 9
	GWL_EXSTYLE      = 0xFFFFFFEC
	GWL_STYLE        = 0xFFFFFFF0
)

var (
	cfg             *Config
	configPath      string
	currentHwnd     win.HWND
	running         bool = true
	whitePen        uintptr
	menuWindow      win.HWND
	listeningForKey int = -1
	webView         webview2.WebView
	webViewMutex    sync.Mutex
	gamePid         uint32
	gameHwnd        win.HWND
	menuOpen        bool = false

	matchLog = &matchLogger{}
)

const VERSION = "1.0"
const CurrentConfigVersion = "1.4"
const AppName = "LogitechUniversalManager"

const (
	WM_TRAYICON = win.WM_USER + 1
	NIM_ADD     = 0x00000000
	NIM_MODIFY  = 0x00000001
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	ID_TRAY_SHOW = 1001
	ID_TRAY_HIDE = 1002
	ID_TRAY_EXIT = 1003

	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	TPM_RIGHTBUTTON = 0x0002
	TPM_RETURNCMD   = 0x0100
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             win.HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            win.HICON
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         windows.GUID
	HBalloonIcon     win.HICON
}

var defaultOffsets = Offset{
	DwViewMatrix:           0x2567FA0,
	DwLocalPlayerPawn:      0x2562808,
	DwEntityList:           0x2717828,
	DwGameRules:            0x2565D50,
	DwBuildNumber:          0x61CF2C,
	M_hPlayerPawn:          0x8A0,
	M_iHealth:              0x358,
	M_lifeState:            0x35C,
	M_iTeamNum:             0x3F4,
	M_vOldOrigin:           0x1390,
	M_pGameSceneNode:       0x330,
	M_modelState:           0x160,
	M_boneArray:            0x80,
	M_nodeToWorld:          0x120,
	M_sSanitizedPlayerName: 0x610,

	M_bIsValveDS:             0xA4,
	M_bIsQueuedMatchmaking:   0x9C,
	M_nQueuedMatchmakingMode: 0xA0,
	M_bHasMatchStarted:       0xB0,
	M_totalRoundsPlayed:      0x88,
	M_iRoundTime:             0x68,
	M_bFreezePeriod:          0x40,
	M_gamePhase:              0x84,
	M_iRoundWinStatus:        0x9AC,
}

func DefaultConfig() *Config {
	return &Config{
		Version:             CurrentConfigVersion,
		ESP:                 true,
		TeamCheck:           true,
		HeadCircle:          true,
		SkeletonRendering:   false,
		BoxRendering:        true,
		NameRendering:       true,
		HealthBarRendering:  true,
		HealthTextRendering: true,
		ScreenProtection:    true,
		FrameDelay:          15,
		ShowDistance:        false,
		ShowSnapLines:       false,
		ShowHeadDot:         false,
		KeyESP:              "None",
		KeyScreenProtection: "None",
		KeyOpenGUI:          "Insert",
	}
}

func GetAppDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, AppName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func GetMenuHtmlDir() (string, error) {
	baseDir, err := GetAppDataDir()
	if err != nil {
		return "", err
	}
	htmlDir := filepath.Join(baseDir, "ui")
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		return "", err
	}
	return htmlDir, nil
}

func GetOffsetsDir() (string, error) {
	baseDir, err := GetAppDataDir()
	if err != nil {
		return "", err
	}
	offsetsDir := filepath.Join(baseDir, "offsets")
	if err := os.MkdirAll(offsetsDir, 0755); err != nil {
		return "", err
	}
	return offsetsDir, nil
}

func GetConfigPath() (string, error) {
	dir, err := GetAppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func exportConfig() error {
	dir, _ := GetAppDataDir()

	configName := "current"
	if configPath != "" {
		baseName := filepath.Base(configPath)
		if strings.HasSuffix(baseName, ".json") {
			configName = strings.TrimSuffix(baseName, ".json")
		}
	}

	exportPath := filepath.Join(dir, "config_"+configName+"_exported.json")
	return SaveConfig(cfg, exportPath)
}

func exportConfigByName(name string) error {
	if name == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	dir, _ := GetAppDataDir()
	configsDir := filepath.Join(dir, "configs")
	sourcePath := filepath.Join(configsDir, name+".json")

	loadedCfg, err := LoadConfig(sourcePath)
	if err != nil {
		return err
	}
	if loadedCfg == nil {
		return fmt.Errorf("config not found")
	}

	exportPath := filepath.Join(configsDir, "config_"+name+"_exported.json")
	return SaveConfig(loadedCfg, exportPath)
}

func resetToDefault() error {
	cfg = DefaultConfig()
	return SaveConfig(cfg, configPath)
}

func getConfigsList() []string {
	dir, _ := GetAppDataDir()
	configsDir := filepath.Join(dir, "configs")

	var configs []string
	files, err := os.ReadDir(configsDir)
	if err != nil {
		return configs
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			configs = append(configs, strings.TrimSuffix(file.Name(), ".json"))
		}
	}

	return configs
}

func saveConfigAs(name string) error {
	if name == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	dir, _ := GetAppDataDir()
	configsDir := filepath.Join(dir, "configs")
	os.MkdirAll(configsDir, 0755)

	configPath := filepath.Join(configsDir, name+".json")
	return SaveConfig(cfg, configPath)
}

func loadConfigByName(name string) error {
	if name == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	dir, _ := GetAppDataDir()
	configsDir := filepath.Join(dir, "configs")
	configPath := filepath.Join(configsDir, name+".json")

	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	if loadedCfg == nil {
		return fmt.Errorf("config not found")
	}

	cfg = loadedCfg
	return SaveConfig(cfg, configPath)
}

func deleteConfig(name string) error {
	if name == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	dir, _ := GetAppDataDir()
	configsDir := filepath.Join(dir, "configs")
	configPath := filepath.Join(configsDir, name+".json")

	return os.Remove(configPath)
}

func exportConfigToFolder(folderPath string) error {
	if folderPath == "" {
		return fmt.Errorf("folder path cannot be empty")
	}

	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return err
	}

	configName := "current"
	if configPath != "" {
		baseName := filepath.Base(configPath)
		if strings.HasSuffix(baseName, ".json") {
			configName = strings.TrimSuffix(baseName, ".json")
		}
	}

	exportPath := filepath.Join(folderPath, "config_"+configName+"_exported.json")
	return SaveConfig(cfg, exportPath)
}

func exportConfigByNameToFolder(name string, folderPath string) error {
	if name == "" || folderPath == "" {
		return fmt.Errorf("name and folder path cannot be empty")
	}

	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return err
	}

	dir, _ := GetAppDataDir()
	configsDir := filepath.Join(dir, "configs")
	sourcePath := filepath.Join(configsDir, name+".json")

	loadedCfg, err := LoadConfig(sourcePath)
	if err != nil {
		return err
	}
	if loadedCfg == nil {
		return fmt.Errorf("config not found")
	}

	exportPath := filepath.Join(folderPath, "config_"+name+"_exported.json")
	return SaveConfig(loadedCfg, exportPath)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Version == "" || cfg.Version != CurrentConfigVersion {
		defaultCfg := DefaultConfig()
		if cfg.KeyESP == "" {
			cfg.KeyESP = defaultCfg.KeyESP
		}
		if cfg.KeyScreenProtection == "" {
			cfg.KeyScreenProtection = defaultCfg.KeyScreenProtection
		}
		if cfg.KeyOpenGUI == "" {
			cfg.KeyOpenGUI = defaultCfg.KeyOpenGUI
		}
		cfg.Version = CurrentConfigVersion
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	cfg.Version = CurrentConfigVersion
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func fetchURL(url string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func SaveOffsets(offsets Offset) error {
	dir, err := GetOffsetsDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "offsets.json")
	data, err := json.MarshalIndent(offsets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadOffsets() (Offset, bool) {
	var offsets Offset
	dir, err := GetOffsetsDir()
	if err != nil {
		return offsets, false
	}
	path := filepath.Join(dir, "offsets.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return offsets, false
	}
	if err := json.Unmarshal(data, &offsets); err != nil {
		return offsets, false
	}
	return offsets, true
}

func UpdateOffsets() Offset {
	result := defaultOffsets

	if cached, ok := LoadOffsets(); ok {
		result = cached
	}

	offsetsDir, err := GetOffsetsDir()
	if err != nil {
		return result
	}

	offsetsURL := "https://raw.githubusercontent.com/a2x/cs2-dumper/main/output/offsets.json"
	clientDllURL := "https://raw.githubusercontent.com/a2x/cs2-dumper/main/output/client_dll.json"

	offsetsData, err1 := fetchURL(offsetsURL, 10*time.Second)
	if err1 != nil {
		offsetsData, err1 = os.ReadFile(filepath.Join(offsetsDir, "_raw_offsets.json"))
	} else {
		os.WriteFile(filepath.Join(offsetsDir, "_raw_offsets.json"), offsetsData, 0644)
	}

	clientData, err2 := fetchURL(clientDllURL, 10*time.Second)
	if err2 != nil {
		clientData, err2 = os.ReadFile(filepath.Join(offsetsDir, "_raw_client_dll.json"))
	} else {
		os.WriteFile(filepath.Join(offsetsDir, "_raw_client_dll.json"), clientData, 0644)
	}

	if err1 != nil || err2 != nil {
		fmt.Println("[offsets] failed to fetch, using cached/defaults")
		return result
	}

	var offsetsRoot struct {
		ClientDLL  map[string]json.Number `json:"client.dll"`
		Engine2DLL map[string]json.Number `json:"engine2.dll"`
	}
	if err := json.Unmarshal(offsetsData, &offsetsRoot); err != nil {
		fmt.Println("[offsets] failed to parse offsets.json:", err)
		return result
	}

	if v, ok := offsetsRoot.ClientDLL["dwViewMatrix"]; ok {
		if n, err := v.Int64(); err == nil {
			result.DwViewMatrix = uintptr(n)
		}
	}
	if v, ok := offsetsRoot.ClientDLL["dwLocalPlayerPawn"]; ok {
		if n, err := v.Int64(); err == nil {
			result.DwLocalPlayerPawn = uintptr(n)
		}
	}
	if v, ok := offsetsRoot.ClientDLL["dwEntityList"]; ok {
		if n, err := v.Int64(); err == nil {
			result.DwEntityList = uintptr(n)
		}
	}
	if v, ok := offsetsRoot.ClientDLL["dwGameRules"]; ok {
		if n, err := v.Int64(); err == nil {
			result.DwGameRules = uintptr(n)
		}
	}
	if v, ok := offsetsRoot.Engine2DLL["dwBuildNumber"]; ok {
		if n, err := v.Int64(); err == nil {
			result.DwBuildNumber = uintptr(n)
		}
	}

	var clientRoot struct {
		ClientDLL struct {
			Classes map[string]struct {
				Fields map[string]json.Number `json:"fields"`
			} `json:"classes"`
		} `json:"client.dll"`
	}
	if err := json.Unmarshal(clientData, &clientRoot); err != nil {
		fmt.Println("[offsets] failed to parse client_dll.json:", err)
		return result
	}

	readField := func(className, fieldName string) (uintptr, bool) {
		classObj, ok := clientRoot.ClientDLL.Classes[className]
		if !ok {
			return 0, false
		}
		val, ok := classObj.Fields[fieldName]
		if !ok {
			return 0, false
		}
		n, err := val.Int64()
		if err != nil {
			return 0, false
		}
		return uintptr(n), true
	}

	if v, ok := readField("CCSPlayerController", "m_hPlayerPawn"); ok {
		result.M_hPlayerPawn = v
	}
	if v, ok := readField("C_BaseEntity", "m_iHealth"); ok {
		result.M_iHealth = v
	}
	if v, ok := readField("C_BaseEntity", "m_lifeState"); ok {
		result.M_lifeState = v
	}
	if v, ok := readField("C_BaseEntity", "m_iTeamNum"); ok {
		result.M_iTeamNum = v
	}
	if v, ok := readField("C_BaseEntity", "m_pGameSceneNode"); ok {
		result.M_pGameSceneNode = v
	}
	if v, ok := readField("C_BasePlayerPawn", "m_vOldOrigin"); ok {
		result.M_vOldOrigin = v
	} else if v, ok := readField("C_BaseEntity", "m_vOldOrigin"); ok {
		result.M_vOldOrigin = v
	}
	if v, ok := readField("CSkeletonInstance", "m_modelState"); ok {
		result.M_modelState = v
	}
	if v, ok := readField("CGameSceneNode", "m_nodeToWorld"); ok {
		result.M_nodeToWorld = v
	}
	if v, ok := readField("CCSPlayerController", "m_sSanitizedPlayerName"); ok {
		result.M_sSanitizedPlayerName = v
	}

	if v, ok := readField("C_CSGameRules", "m_bIsValveDS"); ok {
		result.M_bIsValveDS = v
	}
	if v, ok := readField("C_CSGameRules", "m_bIsQueuedMatchmaking"); ok {
		result.M_bIsQueuedMatchmaking = v
	}
	if v, ok := readField("C_CSGameRules", "m_nQueuedMatchmakingMode"); ok {
		result.M_nQueuedMatchmakingMode = v
	}
	if v, ok := readField("C_CSGameRules", "m_bHasMatchStarted"); ok {
		result.M_bHasMatchStarted = v
	}
	if v, ok := readField("C_CSGameRules", "m_totalRoundsPlayed"); ok {
		result.M_totalRoundsPlayed = v
	}
	if v, ok := readField("C_CSGameRules", "m_iRoundTime"); ok {
		result.M_iRoundTime = v
	}
	if v, ok := readField("C_CSGameRules", "m_bFreezePeriod"); ok {
		result.M_bFreezePeriod = v
	}
	if v, ok := readField("C_CSGameRules", "m_gamePhase"); ok {
		result.M_gamePhase = v
	}
	if v, ok := readField("C_CSGameRules", "m_iRoundWinStatus"); ok {
		result.M_iRoundWinStatus = v
	}

	result.M_boneArray = 0x80

	if err := SaveOffsets(result); err != nil {
		fmt.Println("[offsets] failed to save offsets.json:", err)
	} else {
		fmt.Println("[offsets] saved offsets.json")
	}

	return result
}

func getOffsets() Offset {
	return UpdateOffsets()
}

type matchLogger struct {
	mu          sync.Mutex
	exeDir      string
	testDir     string
	counterFile string
	counter     int
	currentFile string
	currentNum  int
	startTime   time.Time
	active      bool
	lastSeen    time.Time
}

func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(exePath)
}

func initMatchLogger() error {
	exeDir := getExeDir()
	testDir := filepath.Join(exeDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return err
	}

	matchLog.exeDir = exeDir
	matchLog.testDir = testDir
	matchLog.counterFile = filepath.Join(testDir, ".counter")

	data, err := os.ReadFile(matchLog.counterFile)
	if err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			matchLog.counter = n
		}
	}
	fmt.Printf("[matchlog] test dir: %s, counter=%d\n", testDir, matchLog.counter)
	return nil
}

func (ml *matchLogger) saveCounter() {
	_ = os.WriteFile(ml.counterFile, []byte(strconv.Itoa(ml.counter)), 0644)
}

func getCs2Build(procHandle windows.Handle, offsets Offset) string {
	engine2Base, err := getModuleBaseAddress(int(gamePid), "engine2.dll")
	if err != nil {
		return "unknown"
	}
	var build int32
	if err := read(procHandle, engine2Base+offsets.DwBuildNumber, &build); err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%d", build)
}

func getServerType(isValveDS, isQueued bool) string {
	switch {
	case isValveDS && isQueued:
		return "Valve Matchmaking"
	case isValveDS:
		return "Valve Official"
	case isQueued:
		return "Community Matchmaking"
	default:
		return "Community Server"
	}
}

func (ml *matchLogger) StartMatch(procHandle windows.Handle, offsets Offset) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.active {
		return
	}

	ml.counter++
	ml.currentNum = ml.counter
	ml.currentFile = filepath.Join(ml.testDir, fmt.Sprintf("match%02d.txt", ml.currentNum))
	ml.startTime = time.Now()
	ml.lastSeen = time.Now()
	ml.active = true
	ml.saveCounter()

	ml.writeHeaderLocked(procHandle, offsets)
	fmt.Printf("[matchlog] started match %02d -> %s\n", ml.currentNum, ml.currentFile)
}

func (ml *matchLogger) writeHeaderLocked(procHandle windows.Handle, offsets Offset) {
	cs2Build := getCs2Build(procHandle, offsets)

	var gameRules uintptr
	var isValveDS, isQueued bool
	var hasStarted bool
	var totalRounds int32

	clientDll, _ := getModuleBaseAddress(int(gamePid), "client.dll")
	if err := read(procHandle, clientDll+offsets.DwGameRules, &gameRules); err == nil && gameRules != 0 {
		var b byte
		if err := read(procHandle, gameRules+offsets.M_bIsValveDS, &b); err == nil {
			isValveDS = b != 0
		}
		if err := read(procHandle, gameRules+offsets.M_bIsQueuedMatchmaking, &b); err == nil {
			isQueued = b != 0
		}
		if err := read(procHandle, gameRules+offsets.M_bHasMatchStarted, &b); err == nil {
			hasStarted = b != 0
		}
		_ = read(procHandle, gameRules+offsets.M_totalRoundsPlayed, &totalRounds)
	}

	serverType := getServerType(isValveDS, isQueued)
	matchCompleted := "No"
	if hasStarted && totalRounds >= 16 {
		matchCompleted = "Yes"
	}

	header := fmt.Sprintf(`# Match %02d/15

Date: %s
Start / end time: %s / -
CS2 build: %s
Steam client version: 
Mode / map: 
Server type: %s
Match completed: %s
Banned immediately after match: Yes / No
Ban status at follow-up: Yes / No
`,
		ml.currentNum,
		ml.startTime.Format("2006-01-02"),
		ml.startTime.Format("15:04:05"),
		cs2Build,
		serverType,
		matchCompleted,
	)

	_ = os.WriteFile(ml.currentFile, []byte(header), 0644)
}

func (ml *matchLogger) Touch() {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	if !ml.active {
		return
	}
	ml.lastSeen = time.Now()
}

func (ml *matchLogger) EndMatch(procHandle windows.Handle, offsets Offset) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	if !ml.active {
		return
	}

	endTime := time.Now()
	data, err := os.ReadFile(ml.currentFile)
	if err != nil {
		ml.active = false
		return
	}

	content := string(data)
	oldStart := ml.startTime.Format("15:04:05")
	newStart := fmt.Sprintf("%s / %s", oldStart, endTime.Format("15:04:05"))
	content = strings.Replace(content, oldStart+" / -", newStart, 1)

	var gameRules uintptr
	clientDll, _ := getModuleBaseAddress(int(gamePid), "client.dll")
	if err := read(procHandle, clientDll+offsets.DwGameRules, &gameRules); err == nil && gameRules != 0 {
		var hasStarted byte
		var totalRounds int32
		if err := read(procHandle, gameRules+offsets.M_bHasMatchStarted, &hasStarted); err == nil {
			_ = read(procHandle, gameRules+offsets.M_totalRoundsPlayed, &totalRounds)
			if hasStarted != 0 && totalRounds >= 16 {
				content = strings.Replace(content, "Match completed: No", "Match completed: Yes", 1)
			}
		}
	}

	_ = os.WriteFile(ml.currentFile, []byte(content), 0644)

	fmt.Printf("[matchlog] ended match %02d (duration %s)\n",
		ml.currentNum, endTime.Sub(ml.startTime).Round(time.Second))

	ml.active = false
	ml.currentFile = ""
}

func (ml *matchLogger) CheckTimeout(procHandle windows.Handle, offsets Offset) {
	ml.mu.Lock()
	stillActive := ml.active
	lastSeen := ml.lastSeen
	ml.mu.Unlock()

	if stillActive && time.Since(lastSeen) > 30*time.Second {
		ml.EndMatch(procHandle, offsets)
	}
}

func init() {
	runtime.LockOSThread()
}

func findWebViewWindowFast() win.HWND {
	var found win.HWND = 0
	ourPid := uint32(windows.GetCurrentProcessId())

	callback := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		var pid uint32
		getWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
		if pid != ourPid {
			return 1
		}

		visible, _, _ := isWindowVisible.Call(uintptr(hwnd))
		if visible == 0 {
			return 1
		}

		var buf [256]uint16
		getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), 256)
		title := windows.UTF16ToString(buf[:])

		if title == "Logitech Universal Manager" {
			found = hwnd
			return 0
		}
		return 1
	})

	enumWindows.Call(callback, 0)
	return found
}

func findRenderChild(parent win.HWND) win.HWND {
	if parent == 0 {
		return 0
	}

	var found win.HWND = 0

	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		var buf [256]uint16
		getClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), 256)
		className := windows.UTF16ToString(buf[:])

		if className == "Chrome_WidgetWin_1" ||
			className == "Chrome_RenderWidgetHostHWND" ||
			className == "Chrome_WidgetWin_0" {
			vis, _, _ := isWindowVisible.Call(uintptr(hwnd))
			if vis != 0 {
				found = hwnd
				return 0
			}
		}
		return 1
	})

	enumChildWindows.Call(uintptr(parent), cb, 0)
	return found
}

func findGameWindowByPid(pid uint32) win.HWND {
	var found win.HWND = 0

	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		var wpid uint32
		getWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&wpid)))
		if wpid != pid {
			return 1
		}

		vis, _, _ := isWindowVisible.Call(uintptr(hwnd))
		if vis == 0 {
			return 1
		}

		var buf [256]uint16
		getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), 256)
		title := windows.UTF16ToString(buf[:])

		if strings.Contains(title, "Counter-Strike") || strings.Contains(title, "CS2") {
			found = hwnd
			return 0
		}
		if found == 0 {
			found = hwnd
		}
		return 1
	})

	enumWindows.Call(cb, 0)
	return found
}

func forceForeground(hwnd win.HWND) {
	if hwnd == 0 {
		return
	}

	allowSetForegroundWindow.Call(ASFW_ANY)

	setWindowPos.Call(
		uintptr(hwnd),
		HWND_TOPMOST,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_SHOWWINDOW,
	)

	bringWindowToTop.Call(uintptr(hwnd))

	ret, _, _ := setForegroundWindow.Call(uintptr(hwnd))
	fmt.Printf("[forceForeground] SetForegroundWindow = %d\n", ret)

	if ret == 0 {
		fg, _, _ := getForegroundWindow.Call()
		if fg != 0 {
			fgThread, _, _ := getWindowThreadProcessId.Call(fg, 0)
			curThread, _, _ := getCurrentThreadId.Call()

			if fgThread != 0 && fgThread != curThread {
				attachThreadInput.Call(fgThread, curThread, 1)
				ret2, _, _ := setForegroundWindow.Call(uintptr(hwnd))
				attachThreadInput.Call(fgThread, curThread, 0)
				fmt.Printf("[forceForeground] SetForegroundWindow (attached) = %d\n", ret2)
			}
		}
	}

	setFocus.Call(uintptr(hwnd))

	renderChild := findRenderChild(hwnd)
	if renderChild != 0 {
		setFocus.Call(uintptr(renderChild))
	}
}

func showMenu() {
	if menuWindow == 0 {
		return
	}

	exStyle, _, _ := getWindowLongPtrW.Call(uintptr(menuWindow), uintptr(GWL_EXSTYLE))
	newStyle := exStyle | WS_EX_TOPMOST | WS_EX_TOOLWINDOW
	setWindowLongPtrW.Call(uintptr(menuWindow), uintptr(GWL_EXSTYLE), newStyle)

	showWindowProc.Call(uintptr(currentHwnd), SW_HIDE)

	showWindowProc.Call(uintptr(menuWindow), SW_SHOW)
	updateWindowProc.Call(uintptr(menuWindow))
	forceForeground(menuWindow)

	menuOpen = true
	fmt.Println("[showMenu] menu opened")
}

func hideMenu() {
	if menuWindow == 0 {
		return
	}

	showWindowProc.Call(uintptr(menuWindow), SW_HIDE)
	menuOpen = false

	if gameHwnd != 0 {
		allowSetForegroundWindow.Call(ASFW_ANY)
		setForegroundWindow.Call(uintptr(gameHwnd))
	}
	fmt.Println("[hideMenu] menu hidden, focus returned to game")
}

func isGameInFocus() bool {
	if gamePid == 0 {
		return false
	}

	foreground, _, _ := getForegroundWindow.Call()
	if foreground == 0 {
		return false
	}

	if win.HWND(foreground) == currentHwnd || win.HWND(foreground) == menuWindow {
		return false
	}

	var foregroundPid uint32
	getWindowThreadProcessId.Call(foreground, uintptr(unsafe.Pointer(&foregroundPid)))

	return foregroundPid == gamePid
}

func isAnyOfOurWindowsInFocus() bool {
	foreground, _, _ := getForegroundWindow.Call()
	if foreground == 0 {
		return false
	}

	if win.HWND(foreground) == currentHwnd || win.HWND(foreground) == menuWindow {
		return true
	}

	if gamePid != 0 {
		var fgPid uint32
		getWindowThreadProcessId.Call(foreground, uintptr(unsafe.Pointer(&fgPid)))
		if fgPid == gamePid {
			return true
		}
	}

	return false
}

func isInputAllowed() bool {
	foreground, _, _ := getForegroundWindow.Call()
	if foreground == 0 {
		return false
	}

	if win.HWND(foreground) == menuWindow {
		return true
	}

	if win.HWND(foreground) == currentHwnd {
		return false
	}

	if gamePid != 0 {
		var fgPid uint32
		getWindowThreadProcessId.Call(foreground, uintptr(unsafe.Pointer(&fgPid)))
		if fgPid == gamePid {
			return true
		}
	}

	return false
}

func worldToScreen(viewMatrix Matrix, position Vector3) (float32, float32) {
	var screenX float32
	var screenY float32
	screenX = viewMatrix[0][0]*position.X + viewMatrix[0][1]*position.Y + viewMatrix[0][2]*position.Z + viewMatrix[0][3]
	screenY = viewMatrix[1][0]*position.X + viewMatrix[1][1]*position.Y + viewMatrix[1][2]*position.Z + viewMatrix[1][3]
	w := viewMatrix[3][0]*position.X + viewMatrix[3][1]*position.Y + viewMatrix[3][2]*position.Z + viewMatrix[3][3]
	if w < 0.01 {
		return -1, -1
	}
	invw := 1.0 / w
	screenX *= invw
	screenY *= invw
	width, _, _ := getSystemMetrics.Call(0)
	height, _, _ := getSystemMetrics.Call(1)
	widthFloat := float32(width)
	heightFloat := float32(height)
	x := widthFloat / 2
	y := heightFloat / 2
	x += 0.5*screenX*widthFloat + 0.5
	y -= 0.5*screenY*heightFloat + 0.5
	return x, y
}

func getEntitiesInfo(procHandle windows.Handle, clientDll uintptr, screenWidth uintptr, screenHeight uintptr, offsets Offset) FrameData {
	var entityList uintptr
	var data FrameData
	err := read(procHandle, clientDll+offsets.DwEntityList, &entityList)
	if err != nil {
		return data
	}
	var (
		localPlayerP           uintptr
		localPlayerGameScene   uintptr
		localPlayerSceneOrigin Vector3
		localTeam              int32
		listEntry              uintptr
		gameScene              uintptr
		entityController       uintptr
		entityControllerPawn   uintptr
		entityPawn             uintptr
		entityNameAddress      uintptr
		entityBoneArray        uintptr
		entityTeam             int32
		entityHealth           int32
		entityLifeState        int32
		entityName             string
		sanitizedNameStr       string
		entityOrigin           Vector3
		viewMatrix             Matrix
	)
	bones := map[string]int{
		"head": 6, "neck_0": 5, "spine_1": 4, "spine_2": 2, "pelvis": 0,
		"arm_upper_L": 8, "arm_lower_L": 9, "hand_L": 10,
		"arm_upper_R": 13, "arm_lower_R": 14, "hand_R": 15,
		"leg_upper_L": 22, "leg_lower_L": 23, "ankle_L": 24,
		"leg_upper_R": 25, "leg_lower_R": 26, "ankle_R": 27,
	}
	var (
		currentBone      Vector3
		entityHead       Vector3
		entityHeadTop    Vector3
		entityHeadBottom Vector3
	)

	err = read(procHandle, clientDll+offsets.DwLocalPlayerPawn, &localPlayerP)
	if err != nil {
		return data
	}
	err = read(procHandle, localPlayerP+offsets.M_pGameSceneNode, &localPlayerGameScene)
	if err != nil {
		return data
	}
	err = read(procHandle, localPlayerGameScene+offsets.M_nodeToWorld, &localPlayerSceneOrigin)
	if err != nil {
		return data
	}
	err = read(procHandle, clientDll+offsets.DwViewMatrix, &viewMatrix)
	if err != nil {
		return data
	}

	for i := 0; i < 64; i++ {
		err = read(procHandle, entityList+uintptr((8*(i&0x7FFF)>>9)+16), &listEntry)
		if err != nil || listEntry == 0 {
			continue
		}
		err = read(procHandle, listEntry+uintptr(112)*uintptr(i&0x1FF), &entityController)
		if err != nil || entityController == 0 {
			continue
		}
		err = read(procHandle, entityController+offsets.M_hPlayerPawn, &entityControllerPawn)
		if err != nil || entityControllerPawn == 0 {
			continue
		}
		pawnListEntry := uintptr(0)
		err = read(procHandle, entityList+uintptr(0x8*((entityControllerPawn&0x7FFF)>>9)+16), &pawnListEntry)
		if err != nil || pawnListEntry == 0 {
			continue
		}
		err = read(procHandle, pawnListEntry+uintptr(112)*uintptr(entityControllerPawn&0x1FF), &entityPawn)
		if err != nil || entityPawn == 0 {
			continue
		}
		if entityPawn == localPlayerP {
			continue
		}
		err = read(procHandle, entityPawn+offsets.M_iHealth, &entityHealth)
		if err != nil {
			continue
		}
		err = read(procHandle, entityPawn+offsets.M_lifeState, &entityLifeState)
		if err != nil {
			continue
		}
		if entityLifeState != 256 {
			continue
		}
		err = read(procHandle, entityPawn+offsets.M_iTeamNum, &entityTeam)
		if err != nil {
			continue
		}
		if entityTeam == 0 {
			continue
		}
		if cfg != nil && cfg.TeamCheck {
			err = read(procHandle, localPlayerP+offsets.M_iTeamNum, &localTeam)
			if err != nil {
				continue
			}
			if localTeam == entityTeam {
				continue
			}
		}
		if entityHealth < 1 || entityHealth > 100 {
			continue
		}
		err = read(procHandle, entityController+offsets.M_sSanitizedPlayerName, &entityNameAddress)
		if err != nil {
			continue
		}
		err = read(procHandle, entityNameAddress, &entityName)
		if err != nil {
			continue
		}
		if entityName == "" {
			continue
		}
		var sanitizedName strings.Builder
		for _, c := range entityName {
			if unicode.IsLetter(c) || unicode.IsDigit(c) || unicode.IsPunct(c) || unicode.IsSpace(c) {
				sanitizedName.WriteRune(c)
			}
		}
		sanitizedNameStr = sanitizedName.String()
		err = read(procHandle, entityPawn+offsets.M_pGameSceneNode, &gameScene)
		if err != nil {
			continue
		}
		if gameScene == 0 {
			continue
		}
		err = read(procHandle, gameScene+offsets.M_modelState+offsets.M_boneArray, &entityBoneArray)
		if err != nil {
			continue
		}
		if entityBoneArray == 0 {
			continue
		}
		err = read(procHandle, entityPawn+offsets.M_vOldOrigin, &entityOrigin)
		if err != nil {
			continue
		}
		var entityBones map[string]Vector2 = make(map[string]Vector2)
		for boneName, boneIndex := range bones {
			err = read(procHandle, entityBoneArray+uintptr(boneIndex)*32, &currentBone)
			if err != nil {
				continue
			}
			if boneName == "head" {
				entityHead = currentBone
			}
			boneX, boneY := worldToScreen(viewMatrix, currentBone)
			entityBones[boneName] = Vector2{boneX, boneY}
		}
		entityHeadTop = Vector3{entityHead.X, entityHead.Y, entityHead.Z + 7}
		entityHeadBottom = Vector3{entityHead.X, entityHead.Y, entityHead.Z - 5}
		screenPosHeadX, screenPosHeadTopY := worldToScreen(viewMatrix, entityHeadTop)
		_, screenPosHeadBottomY := worldToScreen(viewMatrix, entityHeadBottom)
		screenPosFeetX, screenPosFeetY := worldToScreen(viewMatrix, entityOrigin)
		entityBoxTop := Vector3{entityOrigin.X, entityOrigin.Y, entityOrigin.Z + 70}
		_, screenPosBoxTop := worldToScreen(viewMatrix, entityBoxTop)
		if screenPosHeadX <= -1 || screenPosFeetY <= -1 || screenPosHeadX >= float32(screenWidth) || screenPosHeadTopY >= float32(screenHeight) {
			continue
		}
		boxHeight := screenPosFeetY - screenPosBoxTop

		var tempEntity Entity
		tempEntity.Health = entityHealth
		tempEntity.Team = entityTeam
		tempEntity.Name = sanitizedNameStr
		tempEntity.Distance = entityOrigin.Dist(localPlayerSceneOrigin) / 100.0
		tempEntity.Position = Vector2{screenPosFeetX, screenPosFeetY}
		tempEntity.Bones = entityBones
		tempEntity.HeadPos = Vector3{screenPosHeadX, screenPosHeadTopY, screenPosHeadBottomY}
		tempEntity.Rect = Rectangle{screenPosBoxTop, screenPosFeetX - boxHeight/4, screenPosFeetX + boxHeight/4, screenPosFeetY}
		data.Entities = append(data.Entities, tempEntity)
	}
	return data
}

func drawSkeleton(hdc win.HDC, pen uintptr, bones map[string]Vector2) {
	win.SelectObject(hdc, win.HGDIOBJ(pen))
	win.MoveToEx(hdc, int(bones["head"].X), int(bones["head"].Y), nil)
	win.LineTo(hdc, int32(bones["neck_0"].X), int32(bones["neck_0"].Y))
	win.LineTo(hdc, int32(bones["spine_1"].X), int32(bones["spine_1"].Y))
	win.LineTo(hdc, int32(bones["spine_2"].X), int32(bones["spine_2"].Y))
	win.LineTo(hdc, int32(bones["pelvis"].X), int32(bones["pelvis"].Y))
	win.LineTo(hdc, int32(bones["leg_upper_L"].X), int32(bones["leg_upper_L"].Y))
	win.LineTo(hdc, int32(bones["leg_lower_L"].X), int32(bones["leg_lower_L"].Y))
	win.LineTo(hdc, int32(bones["ankle_L"].X), int32(bones["ankle_L"].Y))
	win.MoveToEx(hdc, int(bones["pelvis"].X), int(bones["pelvis"].Y), nil)
	win.LineTo(hdc, int32(bones["leg_upper_R"].X), int32(bones["leg_upper_R"].Y))
	win.LineTo(hdc, int32(bones["leg_lower_R"].X), int32(bones["leg_lower_R"].Y))
	win.LineTo(hdc, int32(bones["ankle_R"].X), int32(bones["ankle_R"].Y))
	win.MoveToEx(hdc, int(bones["spine_1"].X), int(bones["spine_1"].Y), nil)
	win.LineTo(hdc, int32(bones["arm_upper_L"].X), int32(bones["arm_upper_L"].Y))
	win.LineTo(hdc, int32(bones["arm_lower_L"].X), int32(bones["arm_lower_L"].Y))
	win.LineTo(hdc, int32(bones["hand_L"].X), int32(bones["hand_L"].Y))
	win.MoveToEx(hdc, int(bones["spine_1"].X), int(bones["spine_1"].Y), nil)
	win.LineTo(hdc, int32(bones["arm_upper_R"].X), int32(bones["arm_upper_R"].Y))
	win.LineTo(hdc, int32(bones["arm_lower_R"].X), int32(bones["arm_lower_R"].Y))
	win.LineTo(hdc, int32(bones["hand_R"].X), int32(bones["hand_R"].Y))
}

func renderEntityInfo(hdc win.HDC, tPen uintptr, gPen uintptr, oPen uintptr, hPen uintptr, rect Rectangle, hp int32, name string, headPos Vector3, distance float32) {
	if cfg != nil && cfg.BoxRendering {
		win.SelectObject(hdc, win.HGDIOBJ(tPen))
		win.MoveToEx(hdc, int(rect.Left), int(rect.Top), nil)
		win.LineTo(hdc, int32(rect.Right), int32(rect.Top))
		win.LineTo(hdc, int32(rect.Right), int32(rect.Bottom))
		win.LineTo(hdc, int32(rect.Left), int32(rect.Bottom))
		win.LineTo(hdc, int32(rect.Left), int32(rect.Top))
	}
	if cfg != nil && cfg.HeadCircle {
		radius := int32((int32(headPos.Z) - int32(headPos.Y)) / 2)
		win.SelectObject(hdc, win.HGDIOBJ(oPen))
		win.Ellipse(hdc, int32(headPos.X)-radius-1, int32(headPos.Y)-1, int32(headPos.X)+radius+1, int32(headPos.Z)+1)
		win.SelectObject(hdc, win.HGDIOBJ(hPen))
		win.Ellipse(hdc, int32(headPos.X)-radius, int32(headPos.Y), int32(headPos.X)+radius, int32(headPos.Z))
		win.SelectObject(hdc, win.HGDIOBJ(oPen))
		win.Ellipse(hdc, int32(headPos.X)-radius+1, int32(headPos.Y)+1, int32(headPos.X)+radius-1, int32(headPos.Z)-1)
	}
	if cfg != nil && cfg.ShowHeadDot {
		win.SelectObject(hdc, win.HGDIOBJ(whitePen))
		win.Ellipse(hdc, int32(headPos.X)-3, int32(headPos.Y)-3, int32(headPos.X)+3, int32(headPos.Y)+3)
	}
	if cfg != nil && cfg.ShowSnapLines {
		win.SelectObject(hdc, win.HGDIOBJ(whitePen))
		width, _, _ := getSystemMetrics.Call(0)
		height, _, _ := getSystemMetrics.Call(1)
		centerX := int32(width) / 2
		centerY := int32(height) / 2

		win.MoveToEx(hdc, int(rect.Left), int(rect.Top), nil)
		win.LineTo(hdc, centerX, centerY)

		win.MoveToEx(hdc, int(rect.Left), int(rect.Bottom), nil)
		win.LineTo(hdc, centerX, centerY)
	}
	if cfg != nil && cfg.HealthBarRendering {
		boxHeight := rect.Bottom - rect.Top
		healthPercentage := float64(hp) / 100.0
		greenBoxLeft := rect.Left - 6
		greenBoxRight := rect.Left - 2
		greenBoxBottom := rect.Bottom
		greenBoxTop := rect.Bottom - boxHeight*float32(healthPercentage)
		win.SelectObject(hdc, win.HGDIOBJ(gPen))
		for y := int(greenBoxTop); y < int(greenBoxBottom); y++ {
			win.MoveToEx(hdc, int(greenBoxLeft), y, nil)
			win.LineTo(hdc, int32(greenBoxRight), int32(y))
		}
		win.SelectObject(hdc, win.HGDIOBJ(oPen))
		win.MoveToEx(hdc, int(greenBoxLeft), int(greenBoxTop), nil)
		win.LineTo(hdc, int32(greenBoxRight), int32(greenBoxTop))
		win.LineTo(hdc, int32(greenBoxRight), int32(greenBoxBottom))
		win.LineTo(hdc, int32(greenBoxLeft), int32(greenBoxBottom))
		win.LineTo(hdc, int32(greenBoxLeft), int32(greenBoxTop))
	}
	if cfg != nil && cfg.HealthTextRendering {
		healthText := fmt.Sprintf("%d HP", hp)
		text, _ := windows.UTF16PtrFromString(healthText)
		win.SetTextColor(hdc, win.RGB(byte(0), byte(255), byte(50)))
		setTextAlign.Call(uintptr(hdc), 0x00000002)
		textX := rect.Right
		textY := rect.Top - 20
		win.TextOut(hdc, int32(textX), int32(textY), text, int32(len(healthText)))
	}
	if cfg != nil && cfg.NameRendering {
		displayName := name
		if cfg.ShowDistance {
			displayName = fmt.Sprintf("%s [%.1fm]", name, distance)
		}
		text, _ := windows.UTF16PtrFromString(displayName)
		win.SetTextColor(hdc, win.RGB(byte(255), byte(255), byte(255)))
		setTextAlign.Call(uintptr(hdc), 0x00000006)
		win.TextOut(hdc, int32(rect.Left)+int32((int32(rect.Right)-int32(rect.Left))/2), int32(rect.Top)-14, text, int32(len(displayName)))
	}
}

func createTrayIcon(hwnd win.HWND) {
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = 1
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_TRAYICON
	nid.HIcon = win.LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(win.IDI_APPLICATION))))

	tipText := "Logitech Universal Manager v" + VERSION
	for i, c := range tipText {
		if i < 127 {
			nid.SzTip[i] = uint16(c)
		}
	}

	shell_NotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
}

func removeTrayIcon(hwnd win.HWND) {
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = 1

	shell_NotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

func showTrayMenu(hwnd win.HWND) {
	menu, _, _ := createPopupMenu.Call()
	if menu == 0 {
		return
	}
	defer destroyMenu.Call(menu)

	showText, _ := syscall.UTF16PtrFromString("show Menu")
	hideText, _ := syscall.UTF16PtrFromString("Hide Menu")
	exitText, _ := syscall.UTF16PtrFromString("Exit")

	appendMenu.Call(menu, MF_STRING, ID_TRAY_SHOW, uintptr(unsafe.Pointer(showText)))
	appendMenu.Call(menu, MF_STRING, ID_TRAY_HIDE, uintptr(unsafe.Pointer(hideText)))
	appendMenu.Call(menu, MF_SEPARATOR, 0, 0)
	appendMenu.Call(menu, MF_STRING, ID_TRAY_EXIT, uintptr(unsafe.Pointer(exitText)))

	var pos struct {
		X int32
		Y int32
	}
	getCursorPos.Call(uintptr(unsafe.Pointer(&pos)))

	setForegroundWindow.Call(uintptr(hwnd))
	cmd, _, _ := trackPopupMenu.Call(menu, TPM_RIGHTBUTTON|TPM_RETURNCMD, uintptr(pos.X), uintptr(pos.Y), 0, uintptr(hwnd), 0)
	win.PostMessage(hwnd, win.WM_NULL, 0, 0)

	switch cmd {
	case ID_TRAY_SHOW:
		if menuWindow != 0 {
			showMenu()
		}
	case ID_TRAY_HIDE:
		if menuWindow != 0 {
			hideMenu()
		}
	case ID_TRAY_EXIT:
		running = false
		win.PostQuitMessage(0)
	}
}

func windowProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_TIMER:
		return 0
	case WM_TRAYICON:
		if lParam == win.WM_RBUTTONUP || lParam == win.WM_LBUTTONUP {
			showTrayMenu(hwnd)
		}
		return 0
	case win.WM_DESTROY:
		removeTrayIcon(hwnd)
		win.PostQuitMessage(0)
		return 0
	case win.WM_PAINT:
		var ps win.PAINTSTRUCT
		win.BeginPaint(hwnd, &ps)
		win.EndPaint(hwnd, &ps)
		return 0
	default:
		return win.DefWindowProc(hwnd, msg, wParam, lParam)
	}
}

func initWindow(screenWidth uintptr, screenHeight uintptr) win.HWND {
	className, _ := windows.UTF16PtrFromString("chairWindow")
	windowTitle, _ := windows.UTF16PtrFromString("chair")

	wc := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		Style:         win.CS_HREDRAW | win.CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(windowProc),
		HInstance:     win.GetModuleHandle(nil),
		HIcon:         win.LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(win.IDI_APPLICATION)))),
		HCursor:       win.LoadCursor(0, (*uint16)(unsafe.Pointer(uintptr(win.IDC_ARROW)))),
		HbrBackground: 0,
		LpszClassName: className,
	}

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		return 0
	}

	hInstance := win.GetModuleHandle(nil)
	hwnd := win.CreateWindowEx(
		win.WS_EX_TOPMOST|win.WS_EX_NOACTIVATE|win.WS_EX_LAYERED|win.WS_EX_TRANSPARENT|win.WS_EX_TOOLWINDOW,
		className,
		windowTitle,
		win.WS_POPUP,
		0, 0,
		int32(screenWidth), int32(screenHeight),
		0, 0, hInstance, nil,
	)
	if hwnd == 0 {
		return 0
	}

	setLayeredWindowAttributes.Call(uintptr(hwnd), 0x000000, 0, 0x00000001)

	if cfg != nil && cfg.ScreenProtection {
		const WDA_EXCLUDEFROMCAPTURE = 0x00000011
		setWindowDisplayAffinity.Call(uintptr(hwnd), WDA_EXCLUDEFROMCAPTURE)
	}

	return hwnd
}

func keyToString(keyCode int) string {
	switch keyCode {
	case 0x01:
		return "LMB"
	case 0x02:
		return "RMB"
	case 0x04:
		return "MMB"
	case 0x05:
		return "X1"
	case 0x06:
		return "X2"
	case 0x08:
		return "Backspace"
	case 0x09:
		return "Tab"
	case 0x0D:
		return "Enter"
	case 0x10:
		return "Shift"
	case 0x11:
		return "Ctrl"
	case 0x12:
		return "Alt"
	case 0x13:
		return "Pause"
	case 0x14:
		return "CapsLock"
	case 0x1B:
		return "Escape"
	case 0x20:
		return "Space"
	case 0x21:
		return "PageUp"
	case 0x22:
		return "PageDown"
	case 0x23:
		return "End"
	case 0x24:
		return "Home"
	case 0x25:
		return "Left"
	case 0x26:
		return "Up"
	case 0x27:
		return "Right"
	case 0x28:
		return "Down"
	case 0x2D:
		return "Insert"
	case 0x2E:
		return "Delete"
	case 0x30:
		return "0"
	case 0x31:
		return "1"
	case 0x32:
		return "2"
	case 0x33:
		return "3"
	case 0x34:
		return "4"
	case 0x35:
		return "5"
	case 0x36:
		return "6"
	case 0x37:
		return "7"
	case 0x38:
		return "8"
	case 0x39:
		return "9"
	case 0x41:
		return "A"
	case 0x42:
		return "B"
	case 0x43:
		return "C"
	case 0x44:
		return "D"
	case 0x45:
		return "E"
	case 0x46:
		return "F"
	case 0x47:
		return "G"
	case 0x48:
		return "H"
	case 0x49:
		return "I"
	case 0x4A:
		return "J"
	case 0x4B:
		return "K"
	case 0x4C:
		return "L"
	case 0x4D:
		return "M"
	case 0x4E:
		return "N"
	case 0x4F:
		return "O"
	case 0x50:
		return "P"
	case 0x51:
		return "Q"
	case 0x52:
		return "R"
	case 0x53:
		return "S"
	case 0x54:
		return "T"
	case 0x55:
		return "U"
	case 0x56:
		return "V"
	case 0x57:
		return "W"
	case 0x58:
		return "X"
	case 0x59:
		return "Y"
	case 0x5A:
		return "Z"
	case 0x70:
		return "F1"
	case 0x71:
		return "F2"
	case 0x72:
		return "F3"
	case 0x73:
		return "F4"
	case 0x74:
		return "F5"
	case 0x75:
		return "F6"
	case 0x76:
		return "F7"
	case 0x77:
		return "F8"
	case 0x78:
		return "F9"
	case 0x79:
		return "F10"
	case 0x7A:
		return "F11"
	case 0x7B:
		return "F12"
	default:
		return fmt.Sprintf("0x%X", keyCode)
	}
}

func stringToKey(key string) int {
	switch key {
	case "LMB":
		return 0x01
	case "RMB":
		return 0x02
	case "MMB":
		return 0x04
	case "X1":
		return 0x05
	case "X2":
		return 0x06
	case "Backspace":
		return 0x08
	case "Tab":
		return 0x09
	case "Enter":
		return 0x0D
	case "Shift":
		return 0x10
	case "Ctrl":
		return 0x11
	case "Alt":
		return 0x12
	case "Pause":
		return 0x13
	case "CapsLock":
		return 0x14
	case "Escape":
		return 0x1B
	case "Space":
		return 0x20
	case "PageUp":
		return 0x21
	case "PageDown":
		return 0x22
	case "End":
		return 0x23
	case "Home":
		return 0x24
	case "Left":
		return 0x25
	case "Up":
		return 0x26
	case "Right":
		return 0x27
	case "Down":
		return 0x28
	case "Insert":
		return 0x2D
	case "Delete":
		return 0x2E
	case "0":
		return 0x30
	case "1":
		return 0x31
	case "2":
		return 0x32
	case "3":
		return 0x33
	case "4":
		return 0x34
	case "5":
		return 0x35
	case "6":
		return 0x36
	case "7":
		return 0x37
	case "8":
		return 0x38
	case "9":
		return 0x39
	case "A":
		return 0x41
	case "B":
		return 0x42
	case "C":
		return 0x43
	case "D":
		return 0x44
	case "E":
		return 0x45
	case "F":
		return 0x46
	case "G":
		return 0x47
	case "H":
		return 0x48
	case "I":
		return 0x49
	case "J":
		return 0x4A
	case "K":
		return 0x4B
	case "L":
		return 0x4C
	case "M":
		return 0x4D
	case "N":
		return 0x4E
	case "O":
		return 0x4F
	case "P":
		return 0x50
	case "Q":
		return 0x51
	case "R":
		return 0x52
	case "S":
		return 0x53
	case "T":
		return 0x54
	case "U":
		return 0x55
	case "V":
		return 0x56
	case "W":
		return 0x57
	case "X":
		return 0x58
	case "Y":
		return 0x59
	case "Z":
		return 0x5A
	case "F1":
		return 0x70
	case "F2":
		return 0x71
	case "F3":
		return 0x72
	case "F4":
		return 0x73
	case "F5":
		return 0x74
	case "F6":
		return 0x75
	case "F7":
		return 0x76
	case "F8":
		return 0x77
	case "F9":
		return 0x78
	case "F10":
		return 0x79
	case "F11":
		return 0x7A
	case "F12":
		return 0x7B
	default:
		return 0
	}
}

func getKeyState(keyCode int) bool {
	if keyCode == 0 || keyCode == -1 {
		return false
	}
	ret, _, _ := getAsyncKeyState.Call(uintptr(keyCode))
	return ret&0x8000 != 0
}

func removeKeyBindFromOtherFunctions(keyName string, excludeIndex int) {
	if keyName == "None" || keyName == "" {
		return
	}

	if excludeIndex != 11 && cfg.KeyESP == keyName {
		cfg.KeyESP = "None"
	}
	if excludeIndex != 7 && cfg.KeyScreenProtection == keyName {
		cfg.KeyScreenProtection = "None"
	}
	if excludeIndex != 12 && cfg.KeyOpenGUI == keyName {
		cfg.KeyOpenGUI = "None"
	}
}

func updateAllUI() {
	webViewMutex.Lock()
	defer webViewMutex.Unlock()

	if webView == nil || cfg == nil {
		return
	}

	js := fmt.Sprintf(`
		window.updateUI('toggle_esp', %v);
		window.updateUI('toggle_team_check', %v);
		window.updateUI('toggle_head_circle', %v);
		window.updateUI('toggle_skeleton', %v);
		window.updateUI('toggle_box', %v);
		window.updateUI('toggle_health_bar', %v);
		window.updateUI('toggle_health_text', %v);
		window.updateUI('toggle_name', %v);
		window.updateUI('toggle_screen_protection', %v);
		window.updateUI('toggle_head_dot', %v);
		window.updateUI('toggle_snap_lines', %v);
		window.updateUI('toggle_distance', %v);
		window.updateKeyBindDirect(11, %q);
		window.updateKeyBindDirect(7, %q);
		window.updateKeyBindDirect(12, %q);
	`,
		cfg.ESP,
		cfg.TeamCheck,
		cfg.HeadCircle,
		cfg.SkeletonRendering,
		cfg.BoxRendering,
		cfg.HealthBarRendering,
		cfg.HealthTextRendering,
		cfg.NameRendering,
		cfg.ScreenProtection,
		cfg.ShowHeadDot,
		cfg.ShowSnapLines,
		cfg.ShowDistance,
		cfg.KeyESP,
		cfg.KeyScreenProtection,
		cfg.KeyOpenGUI,
	)

	webView.Dispatch(func() {
		webView.Eval(js)
	})
}

func main() {
	freeConsole.Call()
	runtime.LockOSThread()

	var err error
	configPath, err = GetConfigPath()
	if err != nil {
		return
	}

	cfg, err = LoadConfig(configPath)
	if err != nil || cfg == nil {
		cfg = DefaultConfig()
		SaveConfig(cfg, configPath)
	}

	if err := initMatchLogger(); err != nil {
		fmt.Println("[main] matchLogger init failed:", err)
	}

	screenWidth, _, _ := getSystemMetrics.Call(0)
	screenHeight, _, _ := getSystemMetrics.Call(1)

	hwnd := initWindow(screenWidth, screenHeight)
	if hwnd == 0 {
		return
	}
	currentHwnd = hwnd
	defer win.DestroyWindow(hwnd)

	win.ShowWindow(hwnd, win.SW_HIDE)

	createTrayIcon(hwnd)
	defer removeTrayIcon(hwnd)

	whitePen, _, _ = createPen.Call(win.PS_SOLID, 1, 0xffffff)

	htmlDir, _ := GetMenuHtmlDir()
	htmlFile := filepath.Join(htmlDir, "menu.html")

	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        ::-webkit-scrollbar { width: 4px; height: 4px; }
        ::-webkit-scrollbar-track { background: rgba(0,50,120,0.15); border-radius: 2px; }
        ::-webkit-scrollbar-thumb { background: linear-gradient(180deg, #0066ff, #0033cc); border-radius: 2px; }
        ::-webkit-scrollbar-thumb:hover { background: linear-gradient(180deg, #3388ff, #0055dd); }
        * { scrollbar-width: thin; scrollbar-color: #0066ff rgba(0,50,120,0.15); }
        body {
            font-family: 'Segoe UI', 'Arial', sans-serif;
            background: #02030a;
            color: #e0e8ff;
            height: 100vh;
            overflow: hidden;
            user-select: none;
            display: flex;
            flex-direction: column;
        }
        .header {
            background: linear-gradient(135deg, #02040f 0%, #051030 100%);
            padding: 14px;
            text-align: center;
            border-bottom: 1px solid rgba(0,102,255,0.3);
            flex-shrink: 0;
            box-shadow: 0 2px 20px rgba(0,80,255,0.15);
        }
        .header h1 {
            font-size: 16px;
            font-weight: 800;
            background: linear-gradient(45deg, #0066ff, #00aaff, #0066ff);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            text-shadow: 0 0 30px rgba(0,102,255,0.4);
            letter-spacing: 1px;
        }
        .tabs {
            display: flex;
            background: #030618;
            border-bottom: 1px solid rgba(0,102,255,0.2);
            padding: 0 10px;
            flex-wrap: wrap;
            flex-shrink: 0;
        }
        .tab {
            flex: 1;
            padding: 14px 5px;
            text-align: center;
            cursor: pointer;
            font-size: 11px;
            font-weight: 600;
            color: #445588;
            border-bottom: 2px solid transparent;
            transition: all 0.3s ease;
            text-transform: uppercase;
            letter-spacing: 1px;
            min-width: 60px;
        }
        .tab:hover { color: #6688cc; }
        .tab.active {
            color: #0099ff;
            border-bottom: 2px solid #0066ff;
            text-shadow: 0 0 10px rgba(0,153,255,0.5);
        }
        .content {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            scroll-behavior: smooth;
            background: radial-gradient(circle at top, #050c22 0%, #02030a 60%);
        }
        .content::-webkit-scrollbar { width: 4px; }
        .content::-webkit-scrollbar-track { background: rgba(0,50,120,0.08); }
        .content::-webkit-scrollbar-thumb { background: linear-gradient(180deg, #0066ff, #0033cc); border-radius: 2px; }
        .tab-content { display: none; }
        .tab-content.active { display: block; animation: fadeIn 0.3s ease; }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .category { margin-bottom: 20px; }
        .category-title {
            font-size: 12px;
            font-weight: 700;
            color: #0099ff;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
            padding-bottom: 5px;
            border-bottom: 1px solid rgba(0,102,255,0.3);
            text-shadow: 0 0 8px rgba(0,153,255,0.3);
        }
        .item {
            display: flex;
            align-items: center;
            padding: 14px;
            margin-bottom: 10px;
            background: rgba(0,40,110,0.15);
            border-radius: 10px;
            cursor: pointer;
            transition: all 0.3s ease;
            border: 1px solid rgba(0,102,255,0.15);
        }
        .item:hover {
            background: rgba(0,60,160,0.25);
            border-color: rgba(0,153,255,0.4);
            transform: translateX(5px);
            box-shadow: 0 0 15px rgba(0,102,255,0.2);
        }
        .item.main-esp {
            background: linear-gradient(135deg, rgba(0,102,255,0.2), rgba(0,40,150,0.15));
            border: 1px solid rgba(0,153,255,0.5);
            padding: 12px;
            box-shadow: 0 0 20px rgba(0,102,255,0.15);
        }
        .item.main-esp:hover {
            background: linear-gradient(135deg, rgba(0,120,255,0.3), rgba(0,60,180,0.25));
            border-color: rgba(0,170,255,0.7);
            box-shadow: 0 0 30px rgba(0,102,255,0.35);
        }
        .icon {
            width: 30px;
            height: 30px;
            margin-right: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            background: rgba(0,102,255,0.15);
            border-radius: 8px;
            flex-shrink: 0;
            border: 1px solid rgba(0,102,255,0.2);
        }
        .icon svg { width: 16px; height: 16px; fill: #0099ff; }
        .main-esp .icon {
            background: linear-gradient(135deg, #0066ff, #0033cc);
            border-color: rgba(0,153,255,0.6);
            box-shadow: 0 0 12px rgba(0,102,255,0.5);
        }
        .main-esp .icon svg { fill: #fff; }
        .label {
            flex: 1;
            font-size: 14px;
            font-weight: 500;
            color: #c0d0ff;
        }
        .main-esp .label {
            font-weight: 700;
            font-size: 15px;
            background: linear-gradient(45deg, #0099ff, #00ddff);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .keybind {
            width: 60px;
            height: 26px;
            background: #050a1f;
            border: 1px solid rgba(0,102,255,0.4);
            border-radius: 6px;
            color: #0099ff;
            font-size: 10px;
            font-weight: 700;
            text-align: center;
            line-height: 26px;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-right: 6px;
        }
        .keybind:hover {
            border-color: #0099ff;
            background: rgba(0,102,255,0.2);
            box-shadow: 0 0 10px rgba(0,153,255,0.3);
        }
        .keybind.listening {
            animation: pulse 1s infinite;
            border-color: #ff3355;
            color: #ff3355;
            background: rgba(255,51,85,0.15);
        }
        @keyframes pulse {
            0%, 100% { box-shadow: 0 0 0 0 rgba(255,51,85,0.7); }
            50% { box-shadow: 0 0 0 10px rgba(255,51,85,0); }
        }
        .toggle {
            width: 42px;
            height: 22px;
            background: #0a1025;
            border: 1px solid rgba(0,102,255,0.3);
            border-radius: 11px;
            position: relative;
            transition: all 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
            flex-shrink: 0;
        }
        .toggle.active {
            background: linear-gradient(45deg, #0066ff, #00aaff);
            border-color: rgba(0,170,255,0.7);
            box-shadow: 0 0 15px rgba(0,153,255,0.6);
        }
        .toggle::after {
            content: '';
            position: absolute;
            width: 16px;
            height: 16px;
            background: #c0d0ff;
            border-radius: 50%;
            top: 3px;
            left: 3px;
            transition: all 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
        }
        .toggle.active::after { left: 23px; background: #fff; }
        .footer {
            padding: 10px;
            text-align: center;
            color: #334466;
            font-size: 10px;
            border-top: 1px solid rgba(0,102,255,0.1);
            flex-shrink: 0;
        }
        .footer span { color: #0099ff; text-shadow: 0 0 10px rgba(0,153,255,0.6); }

        .config-input {
            width: 100%;
            padding: 12px 14px;
            background: linear-gradient(135deg, rgba(0,30,80,0.6), rgba(0,15,50,0.6));
            border: 1px solid rgba(0,102,255,0.4);
            border-radius: 10px;
            color: #e0e8ff;
            font-size: 13px;
            margin-bottom: 10px;
            box-sizing: border-box;
            transition: all 0.3s ease;
            box-shadow: inset 0 2px 6px rgba(0,0,0,0.4);
        }
        .config-input::placeholder { color: #445588; }
        .config-input:focus {
            outline: none;
            border-color: #0099ff;
            background: linear-gradient(135deg, rgba(0,50,130,0.7), rgba(0,25,80,0.7));
            box-shadow: 0 0 20px rgba(0,153,255,0.4), inset 0 2px 6px rgba(0,0,0,0.4), inset 0 0 20px rgba(0,153,255,0.1);
        }

        .config-btn {
            position: relative;
            width: 100%;
            padding: 12px 18px;
            background: linear-gradient(135deg, #0066ff 0%, #0044cc 50%, #0066ff 100%);
            background-size: 200% 100%;
            border: 1px solid rgba(0,170,255,0.6);
            border-radius: 10px;
            color: #ffffff;
            font-size: 12px;
            font-weight: 700;
            letter-spacing: 0.8px;
            text-transform: uppercase;
            cursor: pointer;
            transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
            overflow: hidden;
            box-shadow: 0 4px 16px rgba(0,102,255,0.4), 0 0 25px rgba(0,153,255,0.2), inset 0 1px 0 rgba(0,200,255,0.3);
            text-shadow: 0 0 10px rgba(0,220,255,0.5);
        }
        .config-btn::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(0,220,255,0.4), transparent);
            transition: left 0.6s ease;
        }
        .config-btn:hover::before { left: 100%; }
        .config-btn:hover {
            background-position: 100% 0;
            border-color: #00ccff;
            transform: translateY(-3px);
            box-shadow: 0 8px 24px rgba(0,153,255,0.7), 0 0 45px rgba(0,170,255,0.5), inset 0 1px 0 rgba(0,220,255,0.5);
            text-shadow: 0 0 15px rgba(0,220,255,0.9);
        }
        .config-btn:active {
            transform: translateY(-1px);
            box-shadow: 0 3px 10px rgba(0,102,255,0.5), inset 0 2px 8px rgba(0,0,0,0.4);
        }
        .config-btn .btn-icon {
            width: 16px;
            height: 16px;
            fill: currentColor;
            filter: drop-shadow(0 0 5px rgba(0,220,255,0.7));
            transition: all 0.35s ease;
        }
        .config-btn:hover .btn-icon {
            filter: drop-shadow(0 0 10px rgba(0,220,255,1));
            transform: scale(1.1);
        }
        .config-btn.success {
            background: linear-gradient(135deg, #0088ff 0%, #0055dd 50%, #0088ff 100%);
            background-size: 200% 100%;
            border-color: rgba(0,200,255,0.7);
        }
        .config-btn.export {
            background: linear-gradient(135deg, #00aaff 0%, #0077dd 50%, #00aaff 100%);
            background-size: 200% 100%;
            border-color: rgba(0,220,255,0.7);
        }
        .config-btn.danger {
            background: linear-gradient(135deg, #0055cc 0%, #003399 50%, #0055cc 100%);
            background-size: 200% 100%;
        }

        .config-list {
            max-height: 400px;
            overflow-y: auto;
            margin-top: 10px;
            padding-right: 4px;
        }
        .config-list::-webkit-scrollbar { width: 3px; }
        .config-list::-webkit-scrollbar-track { background: rgba(0,50,120,0.08); }
        .config-list::-webkit-scrollbar-thumb { background: linear-gradient(180deg, #0066ff, #0033cc); border-radius: 2px; }
        .config-empty {
            text-align: center;
            padding: 30px 20px;
            color: #445588;
            font-size: 13px;
            background: rgba(0,40,110,0.1);
            border: 1px dashed rgba(0,102,255,0.2);
            border-radius: 10px;
        }
        .config-empty svg {
            width: 40px;
            height: 40px;
            fill: #334466;
            margin-bottom: 10px;
            display: block;
            margin-left: auto;
            margin-right: auto;
        }
        .config-card {
            position: relative;
            background: linear-gradient(135deg, rgba(0,60,160,0.2), rgba(0,30,90,0.15));
            border: 1px solid rgba(0,102,255,0.25);
            border-radius: 12px;
            padding: 12px 14px;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            gap: 12px;
            transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
            overflow: hidden;
            cursor: default;
        }
        .config-card::before {
            content: '';
            position: absolute;
            left: 0;
            top: 0;
            bottom: 0;
            width: 3px;
            background: linear-gradient(180deg, #0066ff, #00ddff, #0066ff);
            opacity: 0;
            transition: opacity 0.35s ease;
            box-shadow: 0 0 12px rgba(0,153,255,0.8);
        }
        .config-card:hover {
            background: linear-gradient(135deg, rgba(0,80,200,0.3), rgba(0,40,120,0.25));
            border-color: rgba(0,153,255,0.6);
            transform: translateX(4px);
            box-shadow: 0 6px 24px rgba(0,102,255,0.35), 0 0 30px rgba(0,153,255,0.15);
        }
        .config-card:hover::before { opacity: 1; }
        .config-card-icon {
            width: 40px;
            height: 40px;
            border-radius: 10px;
            background: linear-gradient(135deg, #0066ff, #0044cc);
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
            box-shadow: 0 0 15px rgba(0,102,255,0.5), inset 0 0 8px rgba(0,200,255,0.2);
            transition: all 0.35s ease;
        }
        .config-card:hover .config-card-icon {
            box-shadow: 0 0 25px rgba(0,153,255,0.8), inset 0 0 12px rgba(0,200,255,0.3);
            transform: scale(1.05);
        }
        .config-card-icon svg { width: 20px; height: 20px; fill: #fff; filter: drop-shadow(0 0 4px rgba(0,200,255,0.6)); }
        .config-card-info { flex: 1; min-width: 0; }
        .config-card-name {
            font-size: 14px;
            font-weight: 700;
            color: #e0e8ff;
            margin-bottom: 3px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            text-shadow: 0 0 10px rgba(0,153,255,0.2);
        }
        .config-card-date {
            font-size: 10px;
            color: #5a7acc;
            letter-spacing: 0.8px;
            text-transform: uppercase;
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .config-card-date::before {
            content: '';
            width: 5px;
            height: 5px;
            background: #00ddff;
            border-radius: 50%;
            box-shadow: 0 0 8px #00ddff;
            animation: pulseDot 2s infinite;
        }
        @keyframes pulseDot {
            0%, 100% { opacity: 1; box-shadow: 0 0 8px #00ddff; }
            50% { opacity: 0.6; box-shadow: 0 0 4px #00ddff; }
        }
        .config-card-actions { display: flex; gap: 8px; flex-shrink: 0; }
        .config-icon-btn {
            position: relative;
            width: 34px;
            height: 34px;
            border: 1px solid rgba(0,153,255,0.5);
            background: linear-gradient(135deg, #0066ff 0%, #0044cc 100%);
            border-radius: 9px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            padding: 0;
            overflow: hidden;
            box-shadow: inset 0 1px 0 rgba(0,200,255,0.4), 0 3px 10px rgba(0,102,255,0.4), 0 0 15px rgba(0,153,255,0.2);
        }
        .config-icon-btn svg {
            width: 16px;
            height: 16px;
            fill: #ffffff;
            transition: all 0.3s ease;
            filter: drop-shadow(0 0 5px rgba(0,220,255,0.7));
        }
        .config-icon-btn:hover {
            transform: translateY(-3px) scale(1.08);
            border-color: #00ddff;
            background: linear-gradient(135deg, #0088ff 0%, #0066dd 100%);
            box-shadow: 0 8px 20px rgba(0,153,255,0.7), 0 0 35px rgba(0,170,255,0.5), inset 0 1px 0 rgba(120,240,255,0.6);
        }
        .config-icon-btn:hover svg { filter: drop-shadow(0 0 10px rgba(0,240,255,1)); transform: scale(1.15); }

        .config-current-save { display: flex; gap: 8px; margin-bottom: 15px; }
        .config-current-save .config-input { flex: 1; margin-bottom: 0; }
        .config-current-save .config-btn { width: auto; padding: 12px 22px; margin-bottom: 0; white-space: nowrap; }
        .toast {
            position: fixed;
            bottom: 60px;
            left: 50%;
            transform: translateX(-50%);
            background: rgba(2,4,15,0.95);
            color: #e0e8ff;
            padding: 12px 24px;
            border-radius: 8px;
            font-size: 13px;
            z-index: 9999;
            opacity: 0;
            transition: opacity 0.3s ease;
            pointer-events: none;
            border: 1px solid rgba(0,153,255,0.4);
            box-shadow: 0 0 20px rgba(0,102,255,0.3);
        }
        .toast.show { opacity: 1; }
        .toast.success { border-color: #00bb55; box-shadow: 0 0 20px rgba(0,187,85,0.4); }
        .toast.error { border-color: #ff3355; box-shadow: 0 0 20px rgba(255,51,85,0.4); }
        .export-folder { display: flex; gap: 8px; margin-bottom: 10px; }
        .export-folder .config-input { flex: 1; margin-bottom: 0; font-size: 12px; }
        .export-folder .config-btn { width: auto; padding: 12px 18px; margin-bottom: 0; white-space: nowrap; font-size: 11px; }

        .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.8);
            backdrop-filter: blur(8px);
            z-index: 10000;
            display: none;
            justify-content: center;
            align-items: center;
        }
        .modal-overlay.show { display: flex; }
        .modal {
            background: linear-gradient(145deg, #050a1f, #02040f);
            border: 1px solid rgba(0,102,255,0.4);
            border-radius: 16px;
            padding: 30px;
            max-width: 420px;
            width: 90%;
            box-shadow: 0 20px 60px rgba(0,0,0,0.9), 0 0 40px rgba(0,102,255,0.3);
        }
        .modal-icon { text-align: center; margin-bottom: 15px; }
        .modal-icon svg { width: 48px; height: 48px; }
        .modal-title {
            font-size: 18px;
            font-weight: 700;
            text-align: center;
            margin-bottom: 10px;
            color: #e0e8ff;
            text-shadow: 0 0 15px rgba(0,153,255,0.3);
        }
        .modal-message {
            font-size: 14px;
            color: #8899bb;
            text-align: center;
            margin-bottom: 25px;
            line-height: 1.6;
        }
        .modal-actions { display: flex; gap: 10px; justify-content: center; }
        .modal-btn {
            padding: 10px 30px;
            border: none;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            min-width: 100px;
        }
        .modal-btn.confirm { background: linear-gradient(45deg, #0066ff, #0044cc); color: #fff; }
        .modal-btn.confirm:hover { transform: scale(1.05); box-shadow: 0 0 20px rgba(0,102,255,0.6); }
        .modal-btn.cancel {
            background: rgba(0,50,120,0.2);
            color: #c0d0ff;
            border: 1px solid rgba(0,102,255,0.3);
        }
        .modal-btn.cancel:hover { background: rgba(0,80,180,0.3); border-color: rgba(0,153,255,0.5); }

        .esp-subitems {
            margin-left: 0;
            border-left: 2px solid rgba(0,102,255,0.35);
            padding-left: 12px;
            overflow: hidden;
            max-height: 0;
            opacity: 0;
            transform: translateY(-8px);
            transition: max-height 0.45s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.35s cubic-bezier(0.4, 0, 0.2, 1), transform 0.35s cubic-bezier(0.4, 0, 0.2, 1), padding 0.35s ease, margin 0.35s ease;
            padding-top: 0;
            padding-bottom: 0;
            margin-top: 0;
        }
        .esp-subitems.expanded {
            max-height: 1200px;
            opacity: 1;
            transform: translateY(0);
            padding-top: 4px;
            padding-bottom: 4px;
            margin-top: 4px;
        }
        .esp-subitems .item { padding: 10px 12px; margin-bottom: 6px; background: rgba(0,40,110,0.1); border-color: rgba(0,102,255,0.1); }
        .esp-subitems .item:hover { background: rgba(0,60,160,0.2); border-color: rgba(0,153,255,0.3); }
        .esp-subitems .icon { width: 26px; height: 26px; }
        .esp-subitems .icon svg { width: 14px; height: 14px; }
        .esp-subitems .label { font-size: 13px; }
        .esp-subitems .toggle { width: 38px; height: 20px; border-radius: 10px; }
        .esp-subitems .toggle::after { width: 14px; height: 14px; }
        .esp-subitems .toggle.active::after { left: 21px; }

        .expand-arrow {
            width: 24px;
            height: 24px;
            margin-left: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
            flex-shrink: 0;
            cursor: pointer;
        }
        .expand-arrow svg {
            width: 20px;
            height: 20px;
            fill: #0099ff;
            transition: fill 0.3s ease, filter 0.3s ease;
            filter: drop-shadow(0 0 4px rgba(0,153,255,0.4));
        }
        .expand-arrow.rotated { transform: rotate(180deg); }
        .expand-arrow:hover svg { fill: #00ddff; filter: drop-shadow(0 0 8px rgba(0,221,255,0.8)); }
    </style>
</head>
<body>
    <div class="modal-overlay" id="modalOverlay">
        <div class="modal">
            <div class="modal-icon" id="modalIcon">
                <svg viewBox="0 0 24 24" fill="none" stroke="#0099ff" stroke-width="2">
                    <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                </svg>
            </div>
            <div class="modal-title" id="modalTitle">Confirm Action</div>
            <div class="modal-message" id="modalMessage">Are you sure you want to do this?</div>
            <div class="modal-actions" id="modalActions">
                <button class="modal-btn cancel" onclick="closeModal(false)">Cancel</button>
                <button class="modal-btn confirm" onclick="closeModal(true)">Confirm</button>
            </div>
        </div>
    </div>
    <div class="toast" id="toast"></div>
    <div class="header">
        <h1>LOGITECH UNIVERSAL MANAGER</h1>
    </div>
    <div class="tabs">
        <div class="tab active" onclick="switchTab(0)">Visuals</div>
        <div class="tab" onclick="switchTab(1)">Misc</div>
        <div class="tab" onclick="switchTab(2)">Configs</div>
    </div>
    <div class="content">
        <div class="tab-content active" id="tab0">
            <div class="category">
                <div class="category-title">ESP Settings</div>
                
                <div class="item main-esp">
                    <div class="icon">
                        <svg viewBox="0 0 24 24">
                            <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
                        </svg>
                    </div>
                    <div class="label" onclick="toggleESPSubItems()" style="cursor:pointer;">ESP</div>
                    <div class="keybind" id="k11" onclick="event.stopPropagation(); startKeyBind(11)">None</div>
                    <div class="toggle active" id="t11" onclick="event.stopPropagation(); toggle(11)"></div>
                    <div class="expand-arrow rotated" id="espArrow" onclick="event.stopPropagation(); toggleESPSubItems();">
                        <svg viewBox="0 0 24 24"><path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"/></svg>
                    </div>
                </div>

                <div class="esp-subitems expanded" id="espSubItems">
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg></div>
                        <div class="label">Team Check</div>
                        <div class="toggle active" id="t0" onclick="event.stopPropagation(); toggle(0)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M3 3h18v18H3z"/></svg></div>
                        <div class="label">Box ESP</div>
                        <div class="toggle active" id="t3" onclick="event.stopPropagation(); toggle(3)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 2C7.58 2 4 5.58 4 10c0 6 8 12 8 12s8-6 8-12c0-4.42-3.58-8-8-8zm0 11c-1.66 0-3-1.34-3-3s1.34-3 3-3 3 1.34 3 3-1.34 3-3 3z"/></svg></div>
                        <div class="label">Skeleton ESP</div>
                        <div class="toggle" id="t2" onclick="event.stopPropagation(); toggle(2)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/></svg></div>
                        <div class="label">Head Circle</div>
                        <div class="toggle active" id="t1" onclick="event.stopPropagation(); toggle(1)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 4c2.21 0 4 1.79 4 4s-1.79 4-4 4-4-1.79-4-4 1.79-4 4-4z"/></svg></div>
                        <div class="label">Head Dot</div>
                        <div class="toggle" id="t8" onclick="event.stopPropagation(); toggle(8)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M3 3v18h18V3H3zm16 16H5V5h14v14z"/></svg></div>
                        <div class="label">Snap Lines</div>
                        <div class="toggle" id="t9" onclick="event.stopPropagation(); toggle(9)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15h4v-6h-4v6zm2-8c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2z"/></svg></div>
                        <div class="label">Name ESP</div>
                        <div class="toggle active" id="t6" onclick="event.stopPropagation(); toggle(6)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M3 3h18v2H3V3zm0 4h18v2H3V7zm0 4h18v2H3v-2zm0 4h18v2H3v-2zm0 4h18v2H3v-2z"/></svg></div>
                        <div class="label">Distance</div>
                        <div class="toggle" id="t10" onclick="event.stopPropagation(); toggle(10)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg></div>
                        <div class="label">Health Bar</div>
                        <div class="toggle active" id="t4" onclick="event.stopPropagation(); toggle(4)"></div>
                    </div>
                    <div class="item">
                        <div class="icon"><svg viewBox="0 0 24 24"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 14H9v-4H5v-2h4V7h2v4h4v2h-4v4z"/></svg></div>
                        <div class="label">Health Text</div>
                        <div class="toggle active" id="t5" onclick="event.stopPropagation(); toggle(5)"></div>
                    </div>
                </div>
            </div>
        </div>
        <div class="tab-content" id="tab1">
            <div class="category">
                <div class="category-title">Menu</div>
                <div class="item">
                    <div class="icon"><svg viewBox="0 0 24 24"><path d="M3 5h18v2H3V5zm0 6h18v2H3v-2zm0 6h18v2H3v-2z"/></svg></div>
                    <div class="label">Open GUI</div>
                    <div class="keybind" id="k12" onclick="event.stopPropagation(); startKeyBind(12)">Insert</div>
                </div>
            </div>
            <div class="category">
                <div class="category-title">Security</div>
                <div class="item">
                    <div class="icon"><svg viewBox="0 0 24 24"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/></svg></div>
                    <div class="label">Screen Protection</div>
                    <div class="keybind" id="k7" onclick="event.stopPropagation(); startKeyBind(7)">None</div>
                    <div class="toggle active" id="t7" onclick="event.stopPropagation(); toggle(7)"></div>
                </div>
            </div>
        </div>
        <div class="tab-content" id="tab2">
            <div class="category">
                <div class="category-title">Save Current Config</div>
                <div class="config-current-save">
                    <input type="text" id="configName" class="config-input" placeholder="Enter config name..." />
                    <button class="config-btn success" onclick="saveConfig()">
                        <svg class="btn-icon" viewBox="0 0 24 24"><path d="M17 3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V7l-4-4zm-5 16c-1.66 0-3-1.34-3-3s1.34-3 3-3 3 1.34 3 3-1.34 3-3 3zm3-10H5V5h10v4z"/></svg>
                        Save
                    </button>
                </div>
            </div>
            <div class="category">
                <div class="category-title">Export Config</div>
                <div class="export-folder">
                    <input type="text" id="exportFolder" class="config-input" placeholder="Export folder path..." />
                    <button class="config-btn export" onclick="browseFolder()">
                        <svg class="btn-icon" viewBox="0 0 24 24"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
                        Browse
                    </button>
                </div>
                <button class="config-btn export" onclick="exportCurrentConfig()">
                    <svg class="btn-icon" viewBox="0 0 24 24"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                    Export Current Config to Folder
                </button>
                <button class="config-btn danger" onclick="resetCurrentConfig()">
                    <svg class="btn-icon" viewBox="0 0 24 24"><path d="M12 5V1L7 6l5 5V7c3.31 0 6 2.69 6 6s-2.69 6-6 6-6-2.69-6-6H4c0 4.42 3.58 8 8 8s8-3.58 8-8-3.58-8-8-8z"/></svg>
                    Reset Current to Default
                </button>
            </div>
            <div class="category">
                <div class="category-title">Saved Configs</div>
                <div class="config-list" id="configList"></div>
            </div>
        </div>
    </div>
    <div class="footer">Made with <span>❤</span> by Voicelayer7</div>

    <script>
        var modalResolve = null;
        window.listeningIndex = -1;
        window.espSubItemsOpen = true;
        
        function showModal(title, message, type, confirmText, cancelText) {
            return new Promise(function(resolve) {
                var overlay = document.getElementById('modalOverlay');
                document.getElementById('modalTitle').textContent = title || 'Confirm Action';
                document.getElementById('modalMessage').textContent = message || 'Are you sure?';
                
                var icon = document.getElementById('modalIcon');
                var actions = document.getElementById('modalActions');
                icon.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="#0099ff" stroke-width="2"><path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>';
                actions.innerHTML = '';
                if (cancelText !== false) {
                    var cancelBtn = document.createElement('button');
                    cancelBtn.className = 'modal-btn cancel';
                    cancelBtn.textContent = cancelText || 'Cancel';
                    cancelBtn.onclick = function() { closeModal(false); };
                    actions.appendChild(cancelBtn);
                }
                var confirmBtn = document.createElement('button');
                confirmBtn.className = 'modal-btn ' + (type || 'confirm');
                confirmBtn.textContent = confirmText || 'Confirm';
                confirmBtn.onclick = function() { closeModal(true); };
                actions.appendChild(confirmBtn);
                overlay.classList.add('show');
                modalResolve = resolve;
            });
        }
        
        function closeModal(result) {
            var overlay = document.getElementById('modalOverlay');
            overlay.classList.remove('show');
            if (modalResolve) { modalResolve(result); modalResolve = null; }
        }
        
        function showToast(message, type) {
            var toast = document.getElementById('toast');
            toast.textContent = message;
            toast.className = 'toast ' + (type || '') + ' show';
            clearTimeout(toast._timeout);
            toast._timeout = setTimeout(function() { toast.classList.remove('show'); }, 3000);
        }
        
        function browseFolder() {
            if (typeof window.openFolderDialog === 'function') {
                window.openFolderDialog().then(function(path) {
                    if (path) {
                        document.getElementById('exportFolder').value = path;
                        window.exportFolderPath = path;
                    }
                });
            }
        }
        
        function switchTab(index) {
            var tabs = document.querySelectorAll('.tab');
            var contents = document.querySelectorAll('.tab-content');
            tabs.forEach(function(tab, i) {
                if (i === index) tab.classList.add('active');
                else tab.classList.remove('active');
            });
            contents.forEach(function(content, i) {
                if (i === index) content.classList.add('active');
                else content.classList.remove('active');
            });
            if (index === 2) { refreshConfigs(); }
        }
        
        function toggle(i) {
            var t = document.getElementById('t' + i);
            if (!t) return;
            t.classList.toggle('active');
            var state = t.classList.contains('active');
            if (typeof window.toggleFunction === 'function') { window.toggleFunction(i, state); }
        }

        function setEspSubItems(open) {
            var sub = document.getElementById('espSubItems');
            var arrow = document.getElementById('espArrow');
            if (!sub) return;
            window.espSubItemsOpen = open;
            if (open) {
                sub.classList.add('expanded');
                if (arrow) arrow.classList.add('rotated');
            } else {
                sub.classList.remove('expanded');
                if (arrow) arrow.classList.remove('rotated');
            }
        }

        function toggleESPSubItems() { setEspSubItems(!window.espSubItemsOpen); }
        document.addEventListener('DOMContentLoaded', function() { setEspSubItems(true); });
        
        function startKeyBind(i) {
            if (window.listeningIndex !== -1 && window.listeningIndex !== i) {
                var oldEl = document.getElementById('k' + window.listeningIndex);
                if (oldEl) {
                    oldEl.classList.remove('listening');
                    oldEl.textContent = oldEl.getAttribute('data-original') || 'None';
                }
            }
            if (window.listeningIndex === i) {
                var el = document.getElementById('k' + i);
                if (el) {
                    el.classList.remove('listening');
                    el.textContent = el.getAttribute('data-original') || 'None';
                }
                window.listeningIndex = -1;
                if (typeof window.startKeyListening === 'function') { window.startKeyListening(-1); }
                return;
            }
            var el = document.getElementById('k' + i);
            if (el) {
                el.setAttribute('data-original', el.textContent);
                el.textContent = '...';
                el.classList.add('listening');
            }
            window.listeningIndex = i;
            if (typeof window.startKeyListening === 'function') { window.startKeyListening(i); }
        }
        
        async function saveConfig() {
            var name = document.getElementById('configName').value.trim();
            if (!name) { showToast('Please enter config name', 'error'); return; }
            var confirmed = await showModal('Save Config', 'Save current settings as "' + name + '"?', 'info', 'Save', 'Cancel');
            if (confirmed) {
                if (typeof window.saveConfigAs === 'function') {
                    window.saveConfigAs(name);
                    document.getElementById('configName').value = '';
                    showToast('Config "' + name + '" saved successfully!', 'success');
                    refreshConfigs();
                }
            }
        }
        
        async function exportCurrentConfig() {
            var folderPath = document.getElementById('exportFolder').value.trim();
            if (!folderPath) { showToast('Please select export folder first', 'error'); return; }
            var confirmed = await showModal('Export Config', 'Export current config to:\n' + folderPath + '\n\nProceed?', 'warning', 'Export', 'Cancel');
            if (confirmed) {
                if (typeof window.exportConfigToFolder === 'function') {
                    window.exportConfigToFolder(folderPath);
                    showToast('Current config exported to: ' + folderPath, 'success');
                }
            }
        }
        
        async function resetCurrentConfig() {
            var confirmed = await showModal('Reset Config', 'Reset current settings to default?\nThis will not delete saved configs.', 'danger', 'Reset', 'Cancel');
            if (confirmed) {
                if (typeof window.resetToDefault === 'function') {
                    window.resetToDefault();
                    showToast('Config reset to default', 'success');
                    updateUIAfterConfigLoad();
                }
            }
        }
        
        function refreshConfigs() {
            if (typeof window.getConfigsList === 'function') {
                window.getConfigsList().then(function(configs) {
                    var listEl = document.getElementById('configList');
                    listEl.innerHTML = '';
                    if (configs.length === 0) {
                        var emptyMsg = document.createElement('div');
                        emptyMsg.className = 'config-empty';
                        emptyMsg.innerHTML = '<svg viewBox="0 0 24 24"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14zm-8-2h2v-4h4v-2h-4V7h-2v4H7v2h4z"/></svg>No saved configs yet<br><span style="font-size:11px;color:#334466;">Save your first config above</span>';
                        listEl.appendChild(emptyMsg);
                        return;
                    }
                    configs.forEach(function(config) {
                        var card = document.createElement('div');
                        card.className = 'config-card';
                        var iconEl = document.createElement('div');
                        iconEl.className = 'config-card-icon';
                        iconEl.innerHTML = '<svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/></svg>';
                        var infoEl = document.createElement('div');
                        infoEl.className = 'config-card-info';
                        var nameEl = document.createElement('div');
                        nameEl.className = 'config-card-name';
                        nameEl.textContent = config;
                        var dateEl = document.createElement('div');
                        dateEl.className = 'config-card-date';
                        dateEl.textContent = 'CONFIG FILE';
                        infoEl.appendChild(nameEl);
                        infoEl.appendChild(dateEl);
                        var actionsEl = document.createElement('div');
                        actionsEl.className = 'config-card-actions';
                        
                        var loadBtn = document.createElement('button');
                        loadBtn.className = 'config-icon-btn load';
                        loadBtn.title = 'Load config';
                        loadBtn.innerHTML = '<svg viewBox="0 0 24 24"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>';
                        loadBtn.onclick = async function(e) {
                            e.stopPropagation();
                            var confirmed = await showModal('Load Config', 'Load config "' + config + '"?\nCurrent settings will be replaced.', 'warning', 'Load', 'Cancel');
                            if (confirmed) {
                                if (typeof window.loadConfigByName === 'function') {
                                    window.loadConfigByName(config);
                                    showToast('Config "' + config + '" loaded!', 'success');
                                    updateUIAfterConfigLoad();
                                }
                            }
                        };
                        
                        var exportBtn = document.createElement('button');
                        exportBtn.className = 'config-icon-btn export-btn';
                        exportBtn.title = 'Export config';
                        exportBtn.innerHTML = '<svg viewBox="0 0 24 24"><path d="M19 12h-4V6h-2v6H9l3 3 3-3zM5 18v2h14v-2H5zM5 4v2h14V4H5z"/></svg>';
                        exportBtn.onclick = async function(e) {
                            e.stopPropagation();
                            var folderPath = document.getElementById('exportFolder').value.trim();
                            if (!folderPath) { showToast('Please select export folder first', 'error'); return; }
                            var confirmed = await showModal('Export Config', 'Export config "' + config + '" to:\n' + folderPath + '\n\nProceed?', 'warning', 'Export', 'Cancel');
                            if (confirmed) {
                                if (typeof window.exportConfigByNameToFolder === 'function') {
                                    window.exportConfigByNameToFolder(config, folderPath);
                                    showToast('Config "' + config + '" exported to: ' + folderPath, 'success');
                                }
                            }
                        };
                        
                        var deleteBtn = document.createElement('button');
                        deleteBtn.className = 'config-icon-btn delete';
                        deleteBtn.title = 'Delete config';
                        deleteBtn.innerHTML = '<svg viewBox="0 0 24 24"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>';
                        deleteBtn.onclick = async function(e) {
                            e.stopPropagation();
                            var confirmed = await showModal('Delete Config', 'Delete config "' + config + '"?\nThis action cannot be undone!', 'danger', 'Delete', 'Cancel');
                            if (confirmed) {
                                if (typeof window.deleteConfig === 'function') {
                                    window.deleteConfig(config);
                                    showToast('Config "' + config + '" deleted', 'success');
                                    refreshConfigs();
                                }
                            }
                        };
                        
                        actionsEl.appendChild(loadBtn);
                        actionsEl.appendChild(exportBtn);
                        actionsEl.appendChild(deleteBtn);
                        card.appendChild(iconEl);
                        card.appendChild(infoEl);
                        card.appendChild(actionsEl);
                        listEl.appendChild(card);
                    });
                });
            }
        }
        
        function updateUIAfterConfigLoad() {
            if (typeof window.getUIState === 'function') {
                window.getUIState().then(function(data) { applyUIState(data); });
            }
        }
        
        function applyUIState(data) {
            var toggleMap = {
                'toggle_esp': 11, 'toggle_team_check': 0, 'toggle_head_circle': 1,
                'toggle_skeleton': 2, 'toggle_box': 3, 'toggle_health_bar': 4,
                'toggle_health_text': 5, 'toggle_name': 6, 'toggle_screen_protection': 7,
                'toggle_head_dot': 8, 'toggle_snap_lines': 9, 'toggle_distance': 10
            };
            var keyMap = { 'key_esp': 11, 'key_screen_protection': 7, 'key_open_gui': 12 };
            for (var action in toggleMap) {
                if (action in data) {
                    var t = document.getElementById('t' + toggleMap[action]);
                    if (t) {
                        if (data[action]) t.classList.add('active');
                        else t.classList.remove('active');
                    }
                }
            }
            for (var action in keyMap) {
                if (action in data) {
                    var k = document.getElementById('k' + keyMap[action]);
                    if (k) {
                        k.textContent = data[action];
                        k.setAttribute('data-original', data[action]);
                    }
                }
            }
        }
        
        window.updateKeyBindDirect = function(index, key) {
            var el = document.getElementById('k' + index);
            if (el) {
                el.textContent = key;
                el.setAttribute('data-original', key);
                el.classList.remove('listening');
            }
            window.listeningIndex = -1;
        };
        
        window.updateUI = function(action, value) {
            var toggleMap = {
                'toggle_esp': 11, 'toggle_team_check': 0, 'toggle_head_circle': 1,
                'toggle_skeleton': 2, 'toggle_box': 3, 'toggle_health_bar': 4,
                'toggle_health_text': 5, 'toggle_name': 6, 'toggle_screen_protection': 7,
                'toggle_head_dot': 8, 'toggle_snap_lines': 9, 'toggle_distance': 10
            };
            if (action in toggleMap) {
                var t = document.getElementById('t' + toggleMap[action]);
                if (t) {
                    if (value) t.classList.add('active');
                    else t.classList.remove('active');
                }
            }
        };

        if (typeof window.getUIState === 'function') {
            window.getUIState().then(function(data) { applyUIState(data); });
        }
    </script>
</body>
</html>`

	os.WriteFile(htmlFile, []byte(htmlContent), 0644)

	webViewReady := make(chan win.HWND, 1)

	go func() {
		runtime.LockOSThread()
		w := webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     false,
			AutoFocus: false,
			WindowOptions: webview2.WindowOptions{
				Title:  "Logitech Universal Manager",
				Width:  460,
				Height: 650,
				IconId: 2,
				Center: true,
			},
		})
		if w == nil {
			webViewReady <- 0
			return
		}

		webViewMutex.Lock()
		webView = w
		webViewMutex.Unlock()

		defer func() {
			w.Destroy()
			running = false
			os.Exit(0)
		}()
		w.SetSize(460, 650, webview2.HintNone)

		w.Bind("toggleFunction", func(index int, value bool) {
			if cfg == nil {
				return
			}
			switch index {
			case 0:
				cfg.TeamCheck = value
			case 1:
				cfg.HeadCircle = value
			case 2:
				cfg.SkeletonRendering = value
			case 3:
				cfg.BoxRendering = value
			case 4:
				cfg.HealthBarRendering = value
			case 5:
				cfg.HealthTextRendering = value
			case 6:
				cfg.NameRendering = value
			case 7:
				cfg.ScreenProtection = value
				if currentHwnd != 0 {
					if value {
						const WDA_EXCLUDEFROMCAPTURE = 0x00000011
						setWindowDisplayAffinity.Call(uintptr(currentHwnd), WDA_EXCLUDEFROMCAPTURE)
					} else {
						setWindowDisplayAffinity.Call(uintptr(currentHwnd), 0x00000000)
					}
				}
			case 8:
				cfg.ShowHeadDot = value
			case 9:
				cfg.ShowSnapLines = value
			case 10:
				cfg.ShowDistance = value
			case 11:
				cfg.ESP = value
			}
			SaveConfig(cfg, configPath)
			updateAllUI()
		})

		w.Bind("startKeyListening", func(index int) {
			listeningForKey = index
		})

		w.Bind("getUIState", func() map[string]interface{} {
			if cfg == nil {
				return make(map[string]interface{})
			}
			return map[string]interface{}{
				"toggle_esp":               cfg.ESP,
				"toggle_team_check":        cfg.TeamCheck,
				"toggle_head_circle":       cfg.HeadCircle,
				"toggle_skeleton":          cfg.SkeletonRendering,
				"toggle_box":               cfg.BoxRendering,
				"toggle_name":              cfg.NameRendering,
				"toggle_health_bar":        cfg.HealthBarRendering,
				"toggle_health_text":       cfg.HealthTextRendering,
				"toggle_screen_protection": cfg.ScreenProtection,
				"toggle_distance":          cfg.ShowDistance,
				"toggle_snap_lines":        cfg.ShowSnapLines,
				"toggle_head_dot":          cfg.ShowHeadDot,
				"key_esp":                  cfg.KeyESP,
				"key_screen_protection":    cfg.KeyScreenProtection,
				"key_open_gui":             cfg.KeyOpenGUI,
			}
		})

		w.Bind("getConfigsList", func() []string {
			return getConfigsList()
		})

		w.Bind("saveConfigAs", func(name string) { saveConfigAs(name) })
		w.Bind("loadConfigByName", func(name string) { loadConfigByName(name); updateAllUI() })
		w.Bind("deleteConfig", func(name string) { deleteConfig(name) })
		w.Bind("exportConfigToFolder", func(folderPath string) { exportConfigToFolder(folderPath) })
		w.Bind("exportConfigByNameToFolder", func(name string, folderPath string) { exportConfigByNameToFolder(name, folderPath) })

		w.Bind("openFolderDialog", func() string {
			type BROWSEINFOW struct {
				HwndOwner      win.HWND
				PidlRoot       uintptr
				PszDisplayName *uint16
				LpszTitle      *uint16
				UlFlags        uint32
				Lpfn           uintptr
				LParam         uintptr
				IImage         int32
			}

			var path [260]uint16
			title, _ := windows.UTF16PtrFromString("Select export folder")

			bi := BROWSEINFOW{
				HwndOwner:      0,
				PidlRoot:       0,
				PszDisplayName: &path[0],
				LpszTitle:      title,
				UlFlags:        0x00000001,
				Lpfn:           0,
				LParam:         0,
				IImage:         0,
			}

			ret, _, _ := shell32.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&bi)))
			if ret == 0 {
				return ""
			}

			shell32.NewProc("SHGetPathFromIDListW").Call(ret, uintptr(unsafe.Pointer(&path[0])))
			ole32.NewProc("CoTaskMemFree").Call(ret)

			return windows.UTF16ToString(path[:])
		})

		w.Bind("resetToDefault", func() { resetToDefault(); updateAllUI() })
		w.Bind("exportConfig", func() { exportConfig() })
		w.Bind("exportConfigByName", func(name string) { exportConfigByName(name) })

		absPath, _ := filepath.Abs(htmlFile)
		fileURL := "file:///" + strings.ReplaceAll(absPath, "\\", "/")

		w.Navigate(fileURL)

		time.Sleep(1500 * time.Millisecond)

		go func() {
			for i := 0; i < 100; i++ {
				hw := findWebViewWindowFast()
				if hw != 0 {
					win.ShowWindow(hw, win.SW_HIDE)

					style, _, _ := getWindowLongPtrW.Call(uintptr(hw), uintptr(GWL_STYLE))
					style &^= uintptr(WS_CAPTION)
					style &^= uintptr(WS_SYSMENU)
					style &^= uintptr(WS_THICKFRAME)
					style &^= uintptr(WS_MINIMIZEBOX)
					style &^= uintptr(WS_MAXIMIZEBOX)
					setWindowLongPtrW.Call(uintptr(hw), uintptr(GWL_STYLE), style)

					exStyle, _, _ := getWindowLongPtrW.Call(uintptr(hw), uintptr(GWL_EXSTYLE))
					newExStyle := exStyle | WS_EX_TOPMOST | WS_EX_TOOLWINDOW
					setWindowLongPtrW.Call(uintptr(hw), uintptr(GWL_EXSTYLE), newExStyle)

					setWindowPos.Call(
						uintptr(hw),
						HWND_TOPMOST,
						0, 0, 0, 0,
						SWP_NOMOVE|SWP_NOSIZE|SWP_NOACTIVATE|SWP_FRAMECHANGED,
					)

					webViewReady <- hw
					return
				}
				time.Sleep(20 * time.Millisecond)
			}
			webViewReady <- 0
		}()

		w.Run()
	}()

	menuWindow = <-webViewReady
	if menuWindow == 0 {
		return
	}
	win.ShowWindow(menuWindow, win.SW_HIDE)
	menuOpen = false

	time.Sleep(500 * time.Millisecond)

	pid, err := findProcessId("cs2.exe")
	if err != nil {
		return
	}

	gamePid = uint32(pid)
	gameHwnd = findGameWindowByPid(gamePid)
	fmt.Printf("[main] gamePid=%d gameHwnd=%v\n", gamePid, gameHwnd)

	clientDll, err := getModuleBaseAddress(pid, "client.dll")
	if err != nil {
		return
	}

	procHandle, err := getProcessHandle(pid)
	if err != nil {
		return
	}

	win.ShowWindow(currentHwnd, win.SW_HIDE)

	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		return
	}

	bgBrush, _, _ := createSolidBrush.Call(uintptr(0x000000))
	defer win.DeleteObject(win.HGDIOBJ(bgBrush))
	redPen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x7a78ff)
	defer win.DeleteObject(win.HGDIOBJ(redPen))
	greenPen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x7dff78)
	defer win.DeleteObject(win.HGDIOBJ(greenPen))
	bluePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0xff8e78)
	defer win.DeleteObject(win.HGDIOBJ(bluePen))
	bonePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0xffffff)
	defer win.DeleteObject(win.HGDIOBJ(bonePen))
	outlinePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x000001)
	defer win.DeleteObject(win.HGDIOBJ(outlinePen))

	font, _, _ := createFont.Call(14, 0, 0, 0, win.FW_BOLD, 0, 0, 0, win.DEFAULT_CHARSET, win.OUT_DEFAULT_PRECIS, win.CLIP_DEFAULT_PRECIS, win.DEFAULT_QUALITY, win.DEFAULT_PITCH|win.FF_DONTCARE, 0)

	offsets := getOffsets()
	fmt.Printf("[main] offsets loaded: entityList=0x%X localPlayer=0x%X viewMatrix=0x%X gameRules=0x%X buildNum=0x%X\n",
		offsets.DwEntityList, offsets.DwLocalPlayerPawn, offsets.DwViewMatrix, offsets.DwGameRules, offsets.DwBuildNumber)

	var msg win.MSG
	var frameDelay uint32 = 15
	var prevKeyStates = make(map[int]bool)
	var wasInFocus bool = false
	var wasInMatch bool = false

	for running {
		if win.PeekMessage(&msg, 0, 0, 0, win.PM_REMOVE) {
			if msg.Message == win.WM_QUIT {
				break
			}
			win.TranslateMessage(&msg)
			win.DispatchMessage(&msg)
			continue
		}

		if cfg != nil {
			frameDelay = cfg.FrameDelay
		}

		isInFocus := isGameInFocus()

		if isInFocus != wasInFocus {
			if isInFocus {
				if !menuOpen {
					win.ShowWindow(currentHwnd, win.SW_SHOWNOACTIVATE)
					win.SetWindowPos(currentHwnd, win.HWND_TOPMOST, 0, 0,
						int32(screenWidth), int32(screenHeight),
						win.SWP_NOACTIVATE|win.SWP_SHOWWINDOW)
				} else {
					showMenu()
				}
			} else {
				if isAnyOfOurWindowsInFocus() {
				} else {
					win.ShowWindow(currentHwnd, win.SW_HIDE)
					if menuOpen && menuWindow != 0 {
						showWindowProc.Call(uintptr(menuWindow), SW_HIDE)
					}
				}
			}
			wasInFocus = isInFocus
		}

		if isInFocus && !menuOpen {
			win.SetWindowPos(currentHwnd, win.HWND_TOPMOST, 0, 0, 0, 0,
				win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE)
		}

		inputAllowed := isInputAllowed()

		if inputAllowed {
			if listeningForKey != -1 {
				for keyCode := 0x01; keyCode <= 0xFE; keyCode++ {
					if keyCode == 0x01 {
						continue
					}
					if keyCode == 0x0A || keyCode == 0x0B || keyCode == 0x5E || keyCode == 0x5F {
						continue
					}
					if getKeyState(keyCode) {
						keyName := "None"
						if keyCode != 0x1B {
							keyName = keyToString(keyCode)
						}

						index := listeningForKey
						listeningForKey = -1

						removeKeyBindFromOtherFunctions(keyName, index)

						switch index {
						case 7:
							cfg.KeyScreenProtection = keyName
						case 11:
							cfg.KeyESP = keyName
						case 12:
							cfg.KeyOpenGUI = keyName
						}

						SaveConfig(cfg, configPath)
						updateAllUI()

						break
					}
				}
			}

			if isInFocus && cfg != nil && cfg.KeyOpenGUI != "None" && cfg.KeyOpenGUI != "" {
				keyCode := stringToKey(cfg.KeyOpenGUI)
				if keyCode != 0 && keyCode != 0x01 {
					currentState := getKeyState(keyCode)
					if currentState && !prevKeyStates[keyCode] {
						if menuOpen {
							hideMenu()
						} else {
							showMenu()
						}
					}
					prevKeyStates[keyCode] = currentState
				}
			}

			if cfg != nil {
				keyBindings := []struct {
					key string
					fn  func()
				}{
					{cfg.KeyESP, func() { cfg.ESP = !cfg.ESP }},
					{cfg.KeyScreenProtection, func() {
						cfg.ScreenProtection = !cfg.ScreenProtection
						if currentHwnd != 0 {
							if cfg.ScreenProtection {
								const WDA_EXCLUDEFROMCAPTURE = 0x00000011
								setWindowDisplayAffinity.Call(uintptr(currentHwnd), WDA_EXCLUDEFROMCAPTURE)
							} else {
								setWindowDisplayAffinity.Call(uintptr(currentHwnd), 0x00000000)
							}
						}
					}},
				}

				for _, binding := range keyBindings {
					if binding.key != "None" && binding.key != "" {
						keyCode := stringToKey(binding.key)
						if keyCode != 0 && keyCode != 0x01 {
							currentState := getKeyState(keyCode)
							if currentState && !prevKeyStates[keyCode] {
								binding.fn()
								SaveConfig(cfg, configPath)
								updateAllUI()
							}
							prevKeyStates[keyCode] = currentState
						}
					}
				}
			}
		}

		if !isInFocus {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if menuOpen {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		memhdc, _, _ := createCompatibleDC.Call(uintptr(hdc))
		memBitmap := win.CreateCompatibleBitmap(hdc, int32(screenWidth), int32(screenHeight))
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(memBitmap))
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(bgBrush))
		win.SetBkMode(win.HDC(memhdc), win.TRANSPARENT)
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(font))

		if cfg != nil && cfg.ESP {
			data := getEntitiesInfo(procHandle, clientDll, screenWidth, screenHeight, offsets)

			if len(data.Entities) > 0 {
				if !matchLog.active {
					matchLog.StartMatch(procHandle, offsets)
				} else {
					matchLog.Touch()
				}
				wasInMatch = true
			} else if wasInMatch {
				matchLog.CheckTimeout(procHandle, offsets)
			}

			for _, entity := range data.Entities {
				if entity.Distance < 0.35 {
					continue
				}
				if cfg.SkeletonRendering {
					drawSkeleton(win.HDC(memhdc), bonePen, entity.Bones)
				}
				if entity.Team == 2 {
					renderEntityInfo(win.HDC(memhdc), redPen, greenPen, outlinePen, bonePen, entity.Rect, entity.Health, entity.Name, entity.HeadPos, entity.Distance)
				} else {
					renderEntityInfo(win.HDC(memhdc), bluePen, greenPen, outlinePen, bonePen, entity.Rect, entity.Health, entity.Name, entity.HeadPos, entity.Distance)
				}
			}
		} else {
			matchLog.CheckTimeout(procHandle, offsets)
		}

		win.BitBlt(hdc, 0, 0, int32(screenWidth), int32(screenHeight), win.HDC(memhdc), 0, 0, win.SRCCOPY)

		win.DeleteObject(win.HGDIOBJ(memBitmap))
		win.DeleteDC(win.HDC(memhdc))

		time.Sleep(time.Duration(frameDelay) * time.Millisecond)
	}

	if matchLog.active {
		matchLog.EndMatch(procHandle, offsets)
	}
}
