package terminal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"flutelake/fluteNAS/pkg/module/flog"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type TerminalConn struct {
	// websocket 连接
	conn *websocket.Conn
	// 最后一次心跳时间
	lastHeatbeat int64
	// UniqueID
	uniqueID string
	// 节点ssh信息
	host *Host
	// 终端名称
	name string
	// 执行记录
	recorder *Recorder
}

type Host struct {
	Hostname   string
	Host       string
	Port       string
	Username   string
	Password   string
	PrivateKey string
}

func (h *Host) isLocalHost() bool {
	switch h.Host {
	case "localhost", "127.0.0.1":
		return true
	default:
		return false
	}
}

func newTerminalConn(conn *websocket.Conn, host *Host, uniqueID, name string) *TerminalConn {
	return &TerminalConn{
		conn:         conn,
		host:         host,
		uniqueID:     uniqueID,
		name:         name,
		lastHeatbeat: time.Now().Unix(),
		recorder:     NewRecorder(host.Hostname, fmt.Sprintf("%s_%s", uniqueID, name)),
	}
}

func (t *TerminalConn) Run() error {
	if t.host.isLocalHost() {
		return t.RunLocal()
	}
	hostport := fmt.Sprintf("%s:%s", t.host.Host, t.host.Port)

	config := &ssh.ClientConfig{
		User:            t.host.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if t.host.Password != "" {
		// 使用密码登录
		config.Auth = append(config.Auth, ssh.Password(t.host.Password))
	}
	if t.host.PrivateKey != "" {
		signers, err := readPrivateKeys(t.host.PrivateKey)
		if err != nil {
			return fmt.Errorf("unable to read private key: %v", err)
		}
		if len(signers) == 0 {
			return fmt.Errorf("not found any private key")
		}
		for _, signer := range signers {
			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		}
	}

	sshConn, err := ssh.Dial("tcp", hostport, config)
	if err != nil {
		return fmt.Errorf("SSH dial error: %v", err)
	}
	defer sshConn.Close()

	session, err := sshConn.NewSession()
	if err != nil {
		return fmt.Errorf("SSH session error: %v", err)
	}
	defer session.Close()

	sshOut, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("STDOUT pipe error: %v", err)
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("STDIN pipe error: %v", err)
	}

	// 需要sshd_config中配置了AcceptEnv HISTFILE，否则会设置失败
	if err := session.Setenv("HISTFILE", "/root/.bash_history"); err != nil {
		return fmt.Errorf("set env error: %v", err)
	}
	if err := session.Setenv("PS1", `➜ \W `); err != nil {
		return fmt.Errorf("set env PS1 error: %v", err)
	}

	if err := session.RequestPty("xterm", 24, 1080, ssh.TerminalModes{}); err != nil {
		return fmt.Errorf("request PTY error: %v", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("start shell error: %v", err)
	}

	// 缓存回写
	if t.recorder != nil {
		t.recorder.ReadLine(func(bs []byte) {
			t.conn.WriteMessage(websocket.BinaryMessage, bs)
		})
		t.conn.WriteMessage(websocket.BinaryMessage, []byte("\r\n"))
	}

	// ssh -> websocket
	go func() {
		defer session.Close()
		buf := make([]byte, 1024)
		for {
			n, err := sshOut.Read(buf)
			if err != nil {
				if err != io.EOF {
					flog.Errorf("read from SSH stdout error: %v", err)
				}
				return
			}
			if n > 0 {
				data := buf[:n]
				err = t.conn.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					flog.Errorf("write to WebSocket error: %v", err)
					return
				}
				if t.recorder != nil {
					t.recorder.Write(data)
				}
			}
		}
	}()

	// 解析执行的命令
	go func() {
		// 记录 & readline
		for {
			if t.recorder == nil || t.recorder.vt == nil {
				continue
			}
			line, err := t.recorder.vt.ReadLine()
			if err == io.EOF {
				continue
			}
			if err != nil {
				flog.Errorf("terminal read line failed, %v\n", err)
				return
			}
			if line == "" {
				continue
			}
			t.recorder.ParseCommandLine(line)
			// log.Infof("terminal line: [%s]\n", line)
		}
	}()

	// websocket -> ssh
	for {
		messageType, p, err := t.conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("read from WebSocket error: %v", err)
			} else {
				return nil
			}
		}
		dataLength := len(p)
		fmt.Printf("received %d message of size %v byte(s) from xterm.js with key sequence: %s\n", messageType, dataLength, string(p))

		// process
		if dataLength <= 0 {
			continue
		}

		// 1开头的为控制信号
		if p[0] == byte('1') {
			arr := strings.Split(string(p), ":")
			if len(arr) == 3 {
				rowsStr := arr[1]
				colsStr := arr[2]
				rows, err := strconv.Atoi(rowsStr)
				if err != nil {
					continue
				}
				cols, err := strconv.Atoi(colsStr)
				if err != nil {
					continue
				}
				fmt.Printf("resizing tty to use %d rows and %d columns...\n", rows, cols)
				if err := session.WindowChange(int(rows), int(cols)); err != nil {
					return fmt.Errorf("failed to resize tty, error: %v", err)
				}
				t.recorder.WriteSize(rows, cols)
				continue
			}
		}
		// 2 心跳信号
		if len(p) == 1 && p[0] == byte('2') {
			// 更新心跳时间
			t.lastHeatbeat = time.Now().Unix()
			continue
		}

		// 终端操作信息
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			sa := bytes.SplitN(p, []byte(":"), 3)
			if len(sa) != 3 {
				continue
			}
			_, err = sshIn.Write(sa[2])
			if err != nil {
				return fmt.Errorf("write to SSH stdin error: %v", err)
			}
		}
	}
}

func (t *TerminalConn) RunLocal() error {
	terminal := "/bin/bash"
	cmd := exec.Command(terminal)
	// cmd := exec.Command("/usr/bin/ssh", "-S", "none", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "root@192.168.177.93")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TERM=xterm")
	cmd.Env = append(cmd.Env, "HISTFILE=$HOME/.bash_history")
	tty, err := pty.Start(cmd)
	if err != nil {
		message := fmt.Sprintf("failed to start tty: %s", err)
		t.conn.WriteMessage(websocket.TextMessage, []byte(message))
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = tty.Close() }() // Best effort.

	// 缓存回写
	if t.recorder != nil {
		t.recorder.ReadLine(func(bs []byte) {
			t.conn.WriteMessage(websocket.BinaryMessage, bs)
		})
		t.conn.WriteMessage(websocket.BinaryMessage, []byte("\r\n"))
	}

	// ssh -> websocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := tty.Read(buf)
			if err != nil {
				if err != io.EOF {
					flog.Errorf("read from SSH stdout error: %v", err)
				}
				return
			}
			if n > 0 {
				data := buf[:n]
				err = t.conn.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					flog.Errorf("write to WebSocket error: %v", err)
					return
				}
				if t.recorder != nil {
					t.recorder.Write(data)
				}
			}
		}
	}()

	// 解析执行的命令
	go func() {
		// 记录 & readline
		for {
			if t.recorder == nil || t.recorder.vt == nil {
				continue
			}
			line, err := t.recorder.vt.ReadLine()
			if err == io.EOF {
				continue
			}
			if err != nil {
				flog.Errorf("terminal read line failed, %v\n", err)
				return
			}
			if line == "" {
				continue
			}
			t.recorder.ParseCommandLine(line)
			// log.Infof("terminal line: [%s]\n", line)
		}
	}()

	// websocket -> ssh
	for {
		messageType, p, err := t.conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("read from WebSocket error: %v", err)
			} else {
				return nil
			}
		}
		dataLength := len(p)
		fmt.Printf("received %d message of size %v byte(s) from xterm.js with key sequence: %s\n", messageType, dataLength, string(p))

		// process
		if dataLength <= 0 {
			continue
		}

		// 1开头的为控制信号
		if p[0] == byte('1') {
			arr := strings.Split(string(p), ":")
			if len(arr) == 3 {
				rowsStr := arr[1]
				colsStr := arr[2]
				rows, err := strconv.Atoi(rowsStr)
				if err != nil {
					continue
				}
				cols, err := strconv.Atoi(colsStr)
				if err != nil {
					continue
				}
				fmt.Printf("resizing tty to use %d rows and %d columns...\n", rows, cols)
				if err := pty.Setsize(tty, &pty.Winsize{
					Rows: uint16(rows),
					Cols: uint16(cols),
				}); err != nil {
					return fmt.Errorf("failed to resize tty, error: %v", err)
				}
				t.recorder.WriteSize(rows, cols)
				continue
			}
		}
		// 2 心跳信号
		if len(p) == 1 && p[0] == byte('2') {
			// 更新心跳时间
			t.lastHeatbeat = time.Now().Unix()
			continue
		}

		// 终端操作信息
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			sa := bytes.SplitN(p, []byte(":"), 3)
			if len(sa) != 3 {
				continue
			}
			_, err = tty.Write(sa[2])
			if err != nil {
				return fmt.Errorf("write to SSH stdin error: %v", err)
			}
		}
	}
}

func (t *TerminalConn) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
	if t.recorder != nil {
		t.recorder.Close()
	}
}

// readPrivateKeys 读取私钥文件
func readPrivateKeys(path string) ([]ssh.Signer, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat private path error: %v", err)
	}
	var signers []ssh.Signer
	filePaths := []string{}
	if info.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("unable to read %s directory: %v", path, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if strings.HasSuffix(file.Name(), ".pub") || file.Name() == "authorized_keys" || file.Name() == "config" || file.Name() == "known_hosts" {
				continue // 跳过非私钥文件
			}

			filePath := filepath.Join(path, file.Name())
			filePaths = append(filePaths, filePath)
		}
	} else {
		filePaths = append(filePaths, path)
	}
	for _, filePath := range filePaths {
		key, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("unable to read private key file %s: %v\n", filePath, err)
			continue // 读取失败，尝试下一个文件
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			fmt.Printf("unable to parse private key file %s: %v\n", filePath, err)
			continue // 解析失败，尝试下一个文件
		}

		signers = append(signers, signer)
	}

	return signers, nil
}
