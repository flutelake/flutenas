package node

import (
	"bytes"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/util"
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
	cmd = cmd + " 2>&1"
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

// WriteFile 写入文件内容到远程主机或本地
func (x *Exec) WriteFile(path string, content []byte, perm os.FileMode) error {
	if x.isLocalHost() {
		return os.WriteFile(path, content, perm)
	}

	if x.client == nil {
		if err := x.Connect(); err != nil {
			return fmt.Errorf("connect error: %v", err)
		}
	}

	session, err := x.client.NewSession()
	if err != nil {
		return fmt.Errorf("create session error: %v", err)
	}
	defer session.Close()

	// 创建目标文件的父目录
	// if _, err := x.Command(fmt.Sprintf("mkdir -p %s", filepath.Dir(path))); err != nil {
	// 	return fmt.Errorf("create directory error: %v", err)
	// }

	// 使用 echo 和重定向来写入文件
	cmd := fmt.Sprintf("echo '%s' > %s", escapeContent(string(content)), path)
	if _, err := x.Command(cmd); err != nil {
		return fmt.Errorf("write file error: %v", err)
	}

	// 设置文件权限
	if _, err := x.Command(fmt.Sprintf("chmod %o %s", perm, path)); err != nil {
		return fmt.Errorf("chmod error: %v", err)
	}

	return nil
}

func (x *Exec) RemoveDir(p string) error {
	isEmpty, err := x.Command(fmt.Sprintf("ls -A %s | wc -l", p))
	if err != nil {
		return fmt.Errorf("检查目录 %s 是否为空失败: %v", p, err)
	} else {
		// 如果目录为空（输出为0），则删除该目录
		if util.Trim(string(isEmpty)) == "0" {
			_, err := x.Command(fmt.Sprintf("rmdir %s", p))
			if err != nil {
				return fmt.Errorf("删除空目录 %s 失败: %v", p, err)
			} else {
				flog.Infof("成功删除空目录: %s", p)
				return nil
			}
		}
	}
	return nil
}

// 解除挂载，解除挂载成功后，判断挂载点路径是否为空，如果为空则删除该路径目录(清理目录报错不返回错误)
func (x *Exec) UmountDir(p string) error {
	bs, err := x.Command(fmt.Sprintf("umount -f %s", p))
	if err != nil {
		// 解挂失败的问题，暂时不返回错误，等控制器来解挂
		return fmt.Errorf("umount -f %s failed: %v, stdout: %s", p, err, string(bs))
	}
	if err = x.RemoveDir(p); err != nil {
		flog.Errorf("remove empty mount point dir error: %v", err)
	}
	return nil
}

// escapeContent 转义文件内容中的特殊字符
func escapeContent(content string) string {
	// 转义单引号和反斜杠
	content = strings.ReplaceAll(content, `\`, `\\`)
	content = strings.ReplaceAll(content, `'`, `'\''`)
	return content
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
