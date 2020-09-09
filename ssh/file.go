package file

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// publicKeyAuthFunc returns the ssh.AuthMethod by private key.
func publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// Create the signer for the private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func CreateClient(user, password, key, host string, port int) (*SSHClient, error) {
	if strings.HasPrefix(host, "http") {
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
	}
	if idx := strings.Index(host, ":"); idx > 0 {
		host = host[:idx]
	}

	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	if password != "" {
		auth = append(auth, ssh.Password(password))
	}

	if key != "" {
		keyAuth, err := publicKeyAuthFunc(key)
		if err != nil {
			return nil, err
		}
		auth = append(auth, keyAuth)
	}

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connect to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	client := &SSHClient{}

	if client.sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if client.sftpClient, err = sftp.NewClient(client.sshClient); err != nil {
		return nil, err
	}

	return client, nil
}

func (t *SSHClient) RemoteWrite(ruleFile string, data []byte) (err error) {

	dstFile, e := t.sftpClient.Create(ruleFile)
	if err = e; err != nil {
		logrus.Error("RemoteWrite create error: ", err.Error(), " rulefile:", ruleFile)
		return
	}
	defer dstFile.Close()

	if _, err = dstFile.Write(data); err != nil {
		logrus.Error("RemoteWrite error: ", err.Error())
	}

	return
}

func (t *SSHClient) Rename(oldname, newname string) (err error) {
	if err = t.sftpClient.Remove(newname); err != nil {
		logrus.Error("Rename create Remove error: ", err.Error(), " newname:", newname)
		return err
	}
	err = t.sftpClient.Rename(oldname, newname)
	if err != nil {
		logrus.Error("Rename create Rename error: ", err.Error(), " oldname:", oldname, " newname:", newname)
		return err
	}

	return err
}
func (t *SSHClient) Close() {
	t.sftpClient.Close()
	t.sshClient.Close()
}

func (t *SSHClient) Run(cmd string) (ret string, err error) {
	session, e := t.sshClient.NewSession()
	if err = e; err != nil {
		logrus.Error("SSHClient NewSession fail! err=", err)
		return
	}
	defer session.Close()
	var buf []byte
	if buf, err = session.CombinedOutput(cmd); err != nil {
		logrus.Error("SSHClient CombinedOutput fail! err=", err, " ret=", string(buf))
	}
	ret = string(buf)
	return
}
