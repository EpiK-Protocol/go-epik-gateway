package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type SSHConfig struct {
	IP             string
	Port           uint64
	UserName       string
	Password       string
	PrivateKeyPath string
}

type SSHClient struct {
	conf   SSHConfig
	client *ssh.Client
	Result string
}

func NewSSH(conf SSHConfig) *SSHClient {
	return &SSHClient{
		conf: conf,
	}
}

func SSHRun(conf SSHConfig, shell string) (string, error) {
	ssh := NewSSH(conf)
	return ssh.Run(shell)
}

func SCPFile(conf SSHConfig, srcFile string, destFile string) error {
	s := NewSSH(conf)
	if s.client == nil {
		if err := s.connect(); err != nil {
			return err
		}
	}

	client, err := scp.NewClientBySSH(s.client)
	if err != nil {
		return err
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Open a file
	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	// Close the file after it has been copied
	defer f.Close()

	// Finaly, copy the file over
	// Usage: CopyFile(fileReader, remotePath, permission)
	err = client.CopyFromFile(*f, destFile, "0655")
	if err != nil {
		return err
	}
	return nil
}

func SCPFileFromRemote(conf SSHConfig, srcFile string, destFile string) error {
	s := NewSSH(conf)
	if s.client == nil {
		if err := s.connect(); err != nil {
			return err
		}
	}

	client, err := scp.NewClientBySSH(s.client)
	if err != nil {
		return err
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Open a file
	f, err := os.Create(destFile)
	if err != nil {
		return err
	}

	// Close the file after it has been copied
	defer f.Close()

	// Finaly, copy the file over
	// Usage: CopyFile(fileReader, remotePath, permission)
	err = client.CopyFromRemote(f, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSHClient) Run(shell string) (string, error) {
	if s.client == nil {
		if err := s.connect(); err != nil {
			return "", err
		}
	}
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(shell)
	if err != nil {
		return "", err
	}
	s.Result = strings.TrimSpace(string(buf))
	return s.Result, nil
}

func (s *SSHClient) connect() error {
	var auth ssh.AuthMethod
	if len(s.conf.Password) > 0 {
		auth = ssh.Password(s.conf.Password)
	} else {
		method, err := publicKeyAuthFunc(s.conf.PrivateKeyPath)
		if err != nil {
			return err
		}
		auth = method
	}
	config := ssh.ClientConfig{
		User:            s.conf.UserName,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         360 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", s.conf.IP, s.conf.Port)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return err
	}
	s.client = sshClient
	return nil
}

func publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func (s *SSHClient) RunTerminal(shell string, stdout, stderr io.Writer) error {
	if s.client == nil {
		if err := s.connect(); err != nil {
			return err
		}
	}
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(fd, oldState)

	session.Stdout = stdout
	session.Stderr = stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}
	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		return err
	}

	session.Run(shell)
	return nil
}
