package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ije/gox/log"
	"github.com/ije/gox/utils"
)

type LoaderWorker struct {
	lock      sync.Mutex
	stdin     io.Writer
	stdout    io.Reader
	outReader *bufio.Reader
}

func (lw *LoaderWorker) Start(loaderjs []byte) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	jsPath := filepath.Join(homeDir, ".esm.sh", "run", fmt.Sprintf("loader@%d.js", VERSION))
	fi, err := os.Stat(jsPath)
	if (err != nil && os.IsNotExist(err)) || (err == nil && fi.Size() != int64(len(loaderjs))) || os.Getenv("DEBUG") == "1" {
		os.MkdirAll(filepath.Dir(jsPath), 0755)
		err = os.WriteFile(jsPath, loaderjs, 0644)
		if err != nil {
			return
		}
	}
	denoPath, err := getDenoPath()
	if err != nil {
		err = errors.New("deno not found, please install deno first")
		return
	}

	cmd := exec.Command(denoPath, "run", "--no-lock", "-A", jsPath)
	cmd.Stdin, lw.stdin = io.Pipe()
	lw.stdout, cmd.Stdout = io.Pipe()
	err = cmd.Start()
	if err != nil {
		lw.stdin = nil
		lw.stdout = nil
	} else {
		lw.outReader = bufio.NewReader(lw.stdout)
		if os.Getenv("DEBUG") == "1" {
			denoVersion, _ := exec.Command(denoPath, "-v").Output()
			fmt.Println(log.Grey(fmt.Sprintf("[debug] loader started (runtime: %s)", strings.TrimSpace(string(denoVersion)))))
		}
	}
	return
}

func (lw *LoaderWorker) Load(loaderType string, args []any) (lang string, code string, err error) {
	// only one load can be invoked at a time
	lw.lock.Lock()
	defer lw.lock.Unlock()

	if lw.outReader == nil {
		err = errors.New("loader not started")
		return
	}

	if os.Getenv("DEBUG") == "1" {
		start := time.Now()
		defer func() {
			if loaderType == "unocss" {
				fmt.Println(log.Grey(fmt.Sprintf("[debug] load 'uno.css' in %s", time.Since(start))))
			} else {
				fmt.Println(log.Grey(fmt.Sprintf("[debug] load '%s' in %s", args[0], time.Since(start))))
			}
		}()
	}

	loaderArgs := make([]any, len(args)+1)
	loaderArgs[0] = loaderType
	copy(loaderArgs[1:], args)
	err = json.NewEncoder(lw.stdin).Encode(loaderArgs)
	if err != nil {
		return
	}
	for {
		var line []byte
		line, err = lw.outReader.ReadBytes('\n')
		if err != nil {
			return
		}
		if len(line) > 3 {
			if bytes.HasPrefix(line, []byte(">>>")) {
				var s string
				t, j := utils.SplitByFirstByte(string(line[3:]), ':')
				err = json.Unmarshal([]byte(j), &s)
				if err != nil {
					return
				}
				if t == "error" {
					err = errors.New(s)
					return
				}
				lang = t
				code = s
				return
			}
		}
	}
}

var lock sync.Mutex

func getDenoPath() (denoPath string, err error) {
	lock.Lock()
	defer lock.Unlock()

	denoPath, err = exec.LookPath("deno")
	if err != nil {
		fmt.Println("Installing deno...")
		denoPath, err = installDeno()
	}
	return
}

func installDeno() (string, error) {
	isWin := runtime.GOOS == "windows"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if !isWin {
		denoPath := filepath.Join(homeDir, ".deno/bin/deno")
		fi, err := os.Stat(denoPath)
		if err == nil && fi.Mode().IsRegular() {
			return denoPath, nil
		}
	}
	installScriptUrl := "https://deno.land/install.sh"
	scriptExe := "sh"
	if isWin {
		installScriptUrl = "https://deno.land/install.ps1"
		scriptExe = "iex"
	}
	res, err := http.Get(installScriptUrl)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", errors.New("failed to get latest deno version")
	}
	defer res.Body.Close()
	cmd := exec.Command(scriptExe)
	cmd.Stdin = res.Body
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	if isWin {
		return exec.LookPath("deno")
	}
	return filepath.Join(homeDir, ".deno/bin/deno"), nil
}
