package node

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Exec struct {
	host   string
	port   string
	client *ssh.Client
}

func NewExec() *Exec {
	return &Exec{}
}

func (x *Exec) SetHost(host string) *Exec {
	x.host = host
	return x
}

func (x *Exec) SetPort(port string) *Exec {
	x.port = port
	return x
}

func (x *Exec) isLocalHost() bool {
	switch strings.ToLower(x.host) {
	case "localhost", "127.0.0.1", "::1", "":
		return true
	default:
		return false
	}
}

func (x *Exec) Connect() error {
	if x.isLocalHost() {
		return nil
	}

	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：在生产环境中应该更严格地验证主机密钥
	}
	signers, err := ReadPrivateKeys("/root/.ssh")
	if err != nil {
		return fmt.Errorf("unable to read private key: %v", err)
	}
	if len(signers) == 0 {
		return fmt.Errorf("not found any private key")
	}
	for _, signer := range signers {
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	// establish a connection to the remote server
	if x.port == "" {
		x.port = "22"
	}
	client, err := ssh.Dial("tcp", x.host+":"+x.port, config)
	if err != nil {
		return err
	}
	x.client = client
	return nil
}

func (x *Exec) Close() {
	if x.client != nil {
		x.client.Close()
	}
}

func (x *Exec) Command(cmd string) ([]byte, error) {
	if x.isLocalHost() {
		return x.localCommand(cmd)
	}
	if x.client == nil {
		if err := x.Connect(); err != nil {
			return nil, err
		}
	}

	session, err := x.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	if err := session.Setenv("LANG", "en_US.UTF-8"); err != nil {
		return nil, fmt.Errorf("set env error: %v", err)
	}

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	// run the command on the remote server
	err = session.Run(cmd)
	if err != nil {
		return nil, err
	}

	return stdoutBuf.Bytes(), nil
}

func (x *Exec) localCommand(cmd string) ([]byte, error) {
	command := exec.Command("sh", "-c", cmd)
	command.Env = append(os.Environ(), "LANG=en_US.UTF-8")
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// readPrivateKeys 读取私钥文件
func ReadPrivateKeys(path string) ([]ssh.Signer, error) {
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
			if strings.HasSuffix(file.Name(), ".pub") || !strings.HasPrefix(file.Name(), "id_") {
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
