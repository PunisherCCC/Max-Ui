package singbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mhsanaei/3x-ui/v2/config"
	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/mhsanaei/3x-ui/v2/util/common"
)

func GetBinaryName() string {
	return fmt.Sprintf("sing-box-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func GetBinaryPath() string {
	return config.GetBinFolderPath() + "/" + GetBinaryName()
}

func GetConfigPath() string {
	return config.GetBinFolderPath() + "/sing-box.json"
}

type Process struct {
	cmd        *exec.Cmd
	version    string
	config     []byte
	configPath string
	logWriter  *LogWriter
	exitErr    error
	startTime  time.Time
	onlineClients []string
}

func NewProcess(cfg []byte) *Process {
	return &Process{
		version:   "Unknown",
		config:    cfg,
		logWriter: NewLogWriter(),
		startTime: time.Now(),
	}
}

func (p *Process) IsRunning() bool {
	return p != nil && p.cmd != nil && p.cmd.Process != nil && p.cmd.ProcessState == nil
}

func (p *Process) GetErr() error {
	if p == nil {
		return nil
	}
	return p.exitErr
}

func (p *Process) GetResult() string {
	if p == nil {
		return ""
	}
	if p.exitErr != nil {
		return p.exitErr.Error()
	}
	return p.logWriter.LastLine()
}

func (p *Process) GetVersion() string {
	if p == nil {
		return "Unknown"
	}
	return p.version
}

func (p *Process) GetConfig() []byte {
	if p == nil {
		return nil
	}
	return p.config
}

func (p *Process) GetAPIPort() int {
	return DefaultAPIPort
}

func (p *Process) GetOnlineClients() []string {
	if p == nil {
		return nil
	}
	return p.onlineClients
}

func (p *Process) SetOnlineClients(users []string) {
	if p != nil {
		p.onlineClients = users
	}
}

func (p *Process) GetUptime() uint64 {
	if p == nil {
		return 0
	}
	return uint64(time.Since(p.startTime).Seconds())
}

// CheckConfig validates a complete generated configuration with the exact
// sing-box binary that will run it, without touching the live config file.
func CheckConfig(data []byte) error {
	if _, err := os.Stat(GetBinaryPath()); err != nil {
		return common.NewErrorf("sing-box binary not found at %s: %v", GetBinaryPath(), err)
	}
	file, err := os.CreateTemp(config.GetBinFolderPath(), "sing-box-check-*.json")
	if err != nil {
		return common.NewErrorf("failed to create temporary sing-box config: %v", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err = file.Write(data); err != nil {
		file.Close()
		return common.NewErrorf("failed to write temporary sing-box config: %v", err)
	}
	if err = file.Close(); err != nil {
		return common.NewErrorf("failed to close temporary sing-box config: %v", err)
	}
	check := exec.Command(GetBinaryPath(), "check", "-c", path)
	if out, err := check.CombinedOutput(); err != nil {
		return common.NewErrorf("sing-box rejected the generated config: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (p *Process) refreshVersion() {
	cmd := exec.Command(GetBinaryPath(), "version")
	data, err := cmd.Output()
	if err != nil {
		p.version = "Unknown"
		return
	}
	fields := bytes.Fields(data)
	if len(fields) >= 3 {
		p.version = string(fields[2])
		return
	}
	p.version = string(bytes.TrimSpace(data))
}

func (p *Process) Start() (err error) {
	if p.IsRunning() {
		return errors.New("sing-box is already running")
	}
	if _, err := os.Stat(GetBinaryPath()); err != nil {
		return common.NewErrorf("sing-box binary not found at %s: %v", GetBinaryPath(), err)
	}
	defer func() {
		if err != nil {
			logger.Error("Failure in running sing-box process: ", err)
			p.exitErr = err
		}
	}()
	var raw any
	if err := json.Unmarshal(p.config, &raw); err != nil {
		return common.NewErrorf("failed to parse sing-box configuration: %v", err)
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return common.NewErrorf("failed to generate sing-box configuration: %v", err)
	}
	if err := os.MkdirAll(config.GetLogFolder(), 0o770); err != nil {
		logger.Warningf("Failed to create log folder: %s", err)
	}
	configPath := GetConfigPath()
	if p.configPath != "" {
		configPath = p.configPath
	}
	if err := os.WriteFile(configPath, data, fs.ModePerm); err != nil {
		return common.NewErrorf("failed to write sing-box configuration file: %v", err)
	}
	check := exec.Command(GetBinaryPath(), "check", "-c", configPath)
	if out, err := check.CombinedOutput(); err != nil {
		return common.NewErrorf("sing-box rejected the generated config: %v: %s", err, strings.TrimSpace(string(out)))
	}
	cmd := exec.Command(GetBinaryPath(), "run", "-c", configPath)
	p.cmd = cmd
	cmd.Stdout = p.logWriter
	cmd.Stderr = p.logWriter
	if err := cmd.Start(); err != nil {
		p.cmd = nil
		return common.NewErrorf("failed to start sing-box: %v", err)
	}
	exited := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		p.exitErr = err
		exited <- err
	}()
	// Binding and config initialization failures occur immediately. Do not report a
	// successful switch until the process has survived that startup window.
	select {
	case err := <-exited:
		p.cmd = nil
		if err == nil {
			err = errors.New("sing-box exited during startup")
		}
		return common.NewErrorf("sing-box failed during startup: %v: %s", err, p.logWriter.LastLine())
	case <-time.After(600 * time.Millisecond):
	}
	p.refreshVersion()
	p.startTime = time.Now()
	p.exitErr = nil
	return nil
}

func (p *Process) Stop() error {
	if !p.IsRunning() {
		return errors.New("sing-box is not running")
	}
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	deadline := time.Now().Add(3 * time.Second)
	for p.IsRunning() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if p.IsRunning() {
		return p.cmd.Process.Kill()
	}
	return nil
}
