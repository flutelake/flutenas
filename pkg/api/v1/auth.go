package v1

import (
	"bufio"
	"encoding/base64"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

type AuthApi struct {
	publicKey  *util.LinkedRune
	privateKey *util.LinkedRune
	cache      cache.TinyCache
}

func NewAuthApi(privateKey *util.LinkedRune, publicKey *util.LinkedRune, c cache.TinyCache) *AuthApi {
	return &AuthApi{
		publicKey:  publicKey,
		privateKey: privateKey,
		cache:      c,
	}
}

func (a *AuthApi) Login(w *apiserver.Response, r *apiserver.Request) {
	in := model.LoginRequest{}
	if err := r.Unmarshal(&in); err != nil {
		w.Write(retcode.StatusError)
		return
	}
	pwdBs, err := base64.StdEncoding.DecodeString(in.Password)
	if err != nil {
		w.Write(retcode.StatusError)
		return
	}
	// flog.Infof(string(pwdBs))
	// testbs, err := util.RSAEncrypt(a.publicKey.String(), []byte("test123"))
	// if err != nil {
	// 	w.Write(retcode.StatusError)
	// 	return
	// }
	// flog.Infof(string(testbs))
	pwd, err := util.RSADecrypt(a.privateKey.String(), pwdBs)
	if err != nil {
		w.Write(retcode.StatusError)
		return
	}

	config := &ssh.ClientConfig{
		User: in.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	port, err := getSshPort()
	if err != nil {
		w.WriteError(err, nil)
		return
	}

	address := fmt.Sprintf("%s:%d", "127.0.0.1", port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		w.WriteError(err, nil)
		return
	}
	defer client.Close()

	// set cookie
	w.SetCookie(model.SessionUserInfo{
		Username: in.Username,
		Password: util.NewLinkedRune(pwd),
		IsAdmin:  true,
	})
	out := model.LoginResponse{}
	w.Write(retcode.StatusOK(out))
}

func (a *AuthApi) Logout(w *apiserver.Response, r *apiserver.Request) {
	cookie, _ := r.GetCookie()
	if cookie != nil {
		a.cache.Delete(apiserver.GenSessionCacheID(cookie.Value))
	}
	w.NullCookie()
	w.Write(retcode.StatusOK(nil))
}

func (a *AuthApi) GetKey(w *apiserver.Response, r *apiserver.Request) {
	key := base64.StdEncoding.EncodeToString([]byte(a.publicKey.String()))
	out := model.KeyResponse{
		Key: key,
	}

	w.Write(retcode.StatusOK(out))
}

func getSshPort() (int, error) {
	file, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		return 22, err // 默认返回端口 22
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Port") {
			parts := strings.Fields(line)
			if len(parts) == 2 {
				return strconv.Atoi(parts[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 22, err // 默认返回端口 22
	}

	return 22, nil // 如果未找到端口配置，默认返回端口 22
}
