package node

import (
	"bufio"
	"bytes"
	"flutelake/fluteNAS/pkg/module/flog"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

/*
# cat /etc/os-release
PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"
NAME="Debian GNU/Linux"
VERSION_ID="12"
VERSION="12 (bookworm)"
VERSION_CODENAME=bookworm
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"
*/
func GetOS(host string) (osRelease string, version string) {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	osRelease, version = "Unknown", "Unknown"
	file, err := exec.Command("cat /etc/os-release")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			osRelease = strings.ToLower(strings.Trim(line[len("ID="):], `"`))
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(line[len("VERSION_ID="):], `"`)
		}
	}

	if err := scanner.Err(); err != nil {
		return
	}

	return
}

func GetKernelVersion(host string) string {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	output, err := exec.Command("uname -r")
	if err != nil {
		flog.Fatalf("Error getting kernel version: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func GetArch(host string) string {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	output, err := exec.Command("uname -m")
	if err != nil {
		flog.Fatalf("Error getting architecture: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func GetHostname(host string) string {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	output, err := exec.Command("hostname")
	if err != nil {
		flog.Fatalf("Error getting hostname: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func GetLocalHostSshPort() (string, error) {
	// 打开 sshd 配置文件
	file, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 默认端口是 22
	port := "22"

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// 忽略注释和空行
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// 查找 "Port" 配置项
		if strings.HasPrefix(line, "Port") {
			parts := strings.Fields(line)
			if len(parts) == 2 {
				// 解析端口号
				port = parts[1]
			}
			break // 找到第一个 Port 配置后退出
		}
	}

	// 检查是否有读取错误
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return port, nil
}

const OS_USER_FLUTE = "flute"

func GetFluteUIDGID() (uid, gid int) {
	u, err := user.Lookup(OS_USER_FLUTE)
	if err != nil {
		return 0, 0
	}
	uid, _ = strconv.Atoi(u.Uid)

	g, err := user.LookupGroup(OS_USER_FLUTE)
	if err != nil {
		return 0, 0
	}
	gid, _ = strconv.Atoi(g.Gid)

	return uid, gid
}

func CreateFluteUserAndGroup() error {
	// 检查 flute 用户是否存在
	_, uErr := user.Lookup(OS_USER_FLUTE)
	_, gErr := user.LookupGroup(OS_USER_FLUTE)

	if gErr != nil {
		// 创建 flute 组
		if _, err := NewExec().Command("groupadd " + OS_USER_FLUTE); err != nil {
			return err
		}
	}
	if uErr != nil {
		// 创建 flute 用户并加入 flute 组
		if _, err := NewExec().Command("useradd -g " + OS_USER_FLUTE + " " + OS_USER_FLUTE); err != nil {
			return err
		}
	}
	return nil
}

// detectSSHServiceName 检测系统上SSH服务的名称 (sshd 或 ssh)
func detectSSHServiceName() string {
	// 检查常见的SSH服务名称
	// 在Debian/Ubuntu系统中通常是ssh，在CentOS/RHEL系统中通常是sshd

	// 首先检查sshd服务是否存在
	_, err := NewExec().Command("systemctl status sshd")
	if err == nil {
		return "sshd"
	}

	// 然后检查ssh服务是否存在
	_, err = NewExec().Command("systemctl status ssh")
	if err == nil {
		return "ssh"
	}

	// 检查sshd服务文件是否存在
	if _, err := os.Stat("/etc/init.d/sshd"); err == nil {
		return "sshd"
	}

	// 检查ssh服务文件是否存在
	if _, err := os.Stat("/etc/init.d/ssh"); err == nil {
		return "ssh"
	}

	// 默认返回sshd (适用于大多数Linux发行版)
	return "sshd"
}

// AddShellOSC133Support 为不同的shell添加OSC 133支持
func AddShellOSC133Support() error {
	usr, err := user.Lookup("root")
	if err != nil {
		return err
	}
	homeDir := usr.HomeDir

	if err := addOSC133ToBash(homeDir); err != nil {
		return err
	}

	// 检查并为zsh添加OSC 133支持
	if err := addOSC133ToZsh(homeDir); err != nil {
		return err
	}

	// 检查并为fish添加OSC 133支持
	if err := addOSC133ToFish(homeDir); err != nil {
		return err
	}

	return nil
}

// addOSC133ToBash 为bash添加OSC 133支持
func addOSC133ToBash(homeDir string) error {
	bashRcPath := filepath.Join(homeDir, ".bashrc")
	bashRcFlutePath := filepath.Join(homeDir, ".bashrc_flute")

	if _, err := os.Stat(bashRcPath); os.IsNotExist(err) {
		return nil
	}

	bashRcContent, err := os.ReadFile(bashRcPath)
	if err != nil {
		return err
	}

	osc133Config := `# Added by FluteNAS for zsh terminal operation
# This won't be added again if you remove it.
if [[ $- == *i* ]]; then
	# 定义 OSC 133 转义序列
	_osc133_prompt_start=$'\e]133;A\a'
	_osc133_command_start=$'\e]133;B\a'
	_osc133_command_executed=$'\e]133;C\a'

	# 处理命令结束和退出状态
	_osc133_precmd() {
		local last_status=$?
		# 发送 D 序列，包含上一个命令的退出码
		printf "\e]133;D;%s\a" "$last_status"
	}

	# 1. 使用 PROMPT_COMMAND 在显示提示符前发送 "命令结束" 信号
	PROMPT_COMMAND="_osc133_precmd${PROMPT_COMMAND:+; $PROMPT_COMMAND}"

	# 2. 修改 PS1：在提示符前后包裹 A 和 B 序列
	# 注意：必须使用 \[ \] 包裹非打印字符，否则会导致换行计算错误
	PS1="\[$_osc133_prompt_start\]\u@\h:\w\$ \[$_osc133_command_start\]"

	# 3. 利用 DEBUG trap 捕获“命令即将开始执行”的瞬间 (即序列 C)
	_osc133_preexec() {
		# 避免在空命令或提示符渲染时触发
		[[ -n "$COMP_LINE" ]] && return
		printf "\e]133;C\a"
	}
	trap '_osc133_preexec' DEBUG
fi`

	if err := os.WriteFile(bashRcFlutePath, []byte(osc133Config), 0o644); err != nil {
		return err
	}

	if strings.Contains(string(bashRcContent), ".bashrc_flute") {
		return nil
	}

	osc133Include := `
# Added by FluteNAS for zsh terminal operation
# This won't be added again if you remove it.
if [ -f "$HOME/.bashrc_flute" ]; then
	. "$HOME/.bashrc_flute"
fi
	`

	updatedContent := string(bashRcContent) + osc133Include
	return os.WriteFile(bashRcPath, []byte(updatedContent), 0o644)
}

// addOSC133ToZsh 为zsh添加OSC 133支持
func addOSC133ToZsh(homeDir string) error {
	zshRcPath := filepath.Join(homeDir, ".zshrc")
	zshRcFlutePath := filepath.Join(homeDir, ".zshrc_flute")

	// 检查文件是否存在
	if _, err := os.Stat(zshRcPath); os.IsNotExist(err) {
		return nil // 文件不存在则跳过
	}

	zshRcContent, err := os.ReadFile(zshRcPath)
	if err != nil {
		return err
	}

	osc133Config := `
# Added by FluteNAS for zsh terminal operation
# This won't be added again if you remove it.

# OSC 133 Terminal Support
# 提示终端开始/结束命令执行
autoload -U add-zsh-hook

osc133_preexec() {
  printf '\033]133;A\007'
}

osc133_precmd() {
  local last_status=$?
  printf '\033]133;D;%s\007' "$last_status"
}

osc133_chpwd() {
  printf '\033]133;C\007'
}

if [ -t 1 ]; then
  case "$TERM" in
    xterm*|rxvt*|screen*|tmux*)
      add-zsh-hook -Uz preexec osc133_preexec
      add-zsh-hook -Uz precmd osc133_precmd
      add-zsh-hook -Uz chpwd osc133_chpwd
      ;;
  esac
fi`

	if err := os.WriteFile(zshRcFlutePath, []byte(osc133Config), 0o644); err != nil {
		return err
	}

	if strings.Contains(string(zshRcContent), ".zshrc_flute") {
		return nil
	}

	osc133Include := `
# Added by FluteNAS for zsh terminal operation
# This won't be added again if you remove it.
if [ -f "$HOME/.zshrc_flute" ]; then
  . "$HOME/.zshrc_flute"
fi
	`

	updatedContent := string(zshRcContent) + osc133Include
	return os.WriteFile(zshRcPath, []byte(updatedContent), 0o644)
}

// addOSC133ToFish 为fish添加OSC 133支持
func addOSC133ToFish(homeDir string) error {
	configDir := filepath.Join(homeDir, ".config", "fish")
	fishConfigPath := filepath.Join(configDir, "config.fish")
	fishFluteConfigPath := filepath.Join(configDir, "config_flute.fish")

	// 检查fish配置目录是否存在
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return nil // fish未安装或未配置
	}

	if _, err := os.Stat(fishConfigPath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(fishConfigPath)
	if err != nil {
		return err
	}

	osc133Config := "# OSC 133 Terminal Support\n"
	osc133Config += "# 提示终端开始/结束命令执行\n"
	osc133Config += "if test -t 1\n"
	osc133Config += "  switch $TERM\n"
	osc133Config += "    case xterm\\* rxvt\\* screen\\* tmux\\*\n"
	osc133Config += "      function osc133_cmd_executed --on-event fish_postexec\n"
	osc133Config += "        printf '\\033]133;D;%s\\007' $status\n"
	osc133Config += "      end\n"
	osc133Config += "      \n"
	osc133Config += "      function osc133_cmd_started --on-event fish_preexec\n"
	osc133Config += "        printf '\\033]133;A\\007'\n"
	osc133Config += "      end\n"
	osc133Config += "      \n"
	osc133Config += "      function osc133_dir_changed --on-variable PWD\n"
	osc133Config += "        printf '\\033]133;C\\007'\n"
	osc133Config += "      end\n"
	osc133Config += "      \n"
	osc133Config += "      # 初始化时设置当前目录\n"
	osc133Config += "      printf '\\033]133;C\\007'\n"
	osc133Config += "    case '*'\n"
	osc133Config += "  end\n"
	osc133Config += "end\n"

	if err := os.WriteFile(fishFluteConfigPath, []byte(osc133Config), 0o644); err != nil {
		return err
	}

	if strings.Contains(string(content), "config_flute.fish") {
		return nil
	}

	includeConfig := "\nif test -f ~/.config/fish/config_flute.fish\n"
	includeConfig += "  source ~/.config/fish/config_flute.fish\n"
	includeConfig += "end\n"

	updatedContent := string(content) + includeConfig
	return os.WriteFile(fishConfigPath, []byte(updatedContent), 0o644)
}

// 修改文件的所属权限为644，用户和组都为flute
func Belong2Flute(p string) error {
	uid, gid := GetFluteUIDGID()
	if uid == 0 && gid == 0 {
		return nil
	}
	if err := os.Chown(p, uid, gid); err != nil {
		return err
	}
	if err := os.Chmod(p, 0o644); err != nil {
		return err
	}
	return nil
}

// SetSshdConfig 配置sshd以允许设置环境变量LANG、PS1、HISTFILE
func SetSshdConfig() error {
	// 读取当前sshd配置
	content, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	acceptEnvFound := false
	var newLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// 检查是否已有 AcceptEnv 配置
		if strings.HasPrefix(trimmedLine, "AcceptEnv") {
			// 检查是否已经包含了我们需要的所有环境变量
			hasLang := strings.Contains(trimmedLine, "LANG")
			hasPs1 := strings.Contains(trimmedLine, "PS1")
			hasHistfile := strings.Contains(trimmedLine, "HISTFILE")

			if hasLang && hasPs1 && hasHistfile {
				// 已经包含了所有需要的环境变量
				acceptEnvFound = true
				newLines = append(newLines, line)
				continue
			} else {
				// 不包含所有需要的环境变量，跳过这一行（稍后添加新的）
				continue
			}
		}

		// 添加其他非AcceptEnv的行
		newLines = append(newLines, line)
	}

	// 添加 AcceptEnv 配置行，如果尚未存在包含所需变量的配置
	if !acceptEnvFound {
		newLines = append(newLines, "")
		newLines = append(newLines, "# Allow specific environment variables")
		newLines = append(newLines, "AcceptEnv LANG LC_*")
		newLines = append(newLines, "AcceptEnv PS1 HISTFILE")
	}

	// 写入更新后的配置
	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile("/etc/ssh/sshd_config", []byte(newContent), 0o644)
	if err != nil {
		return err
	}

	// 检测系统上SSH服务的名称 (sshd 或 ssh)
	serviceName := detectSSHServiceName()

	// 尝试重启SSH服务使配置生效
	_, err = NewExec().Command("systemctl restart " + serviceName)
	if err != nil {
		// 如果systemctl不可用，尝试使用service命令
		_, err = NewExec().Command("service " + serviceName + " restart")
		if err != nil {
			// 如果两种方式都失败，记录警告但不返回错误
			flog.Warnf("Warning: Could not restart %s service: %v", serviceName, err)
		}
	}

	return nil
}
