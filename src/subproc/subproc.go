package subproc

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

const ETX = "\x03"

type Subproc struct {
	Inputs  []string
	Outputs []string
	Cmd     *exec.Cmd
	In      io.Writer
	Out     *bufio.Reader
}

func Init(path string, args ...string) *Subproc {
	cmd := exec.Command(path, args...)
	cmd.Stderr = os.Stderr

	si, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	so, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &Subproc{
		Inputs:  []string{},
		Outputs: []string{},
		Cmd:     cmd,
		In:      si,
		Out:     bufio.NewReader(so),
	}
}

func (p *Subproc) Send(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}
	_, err := p.In.Write(append([]byte(input), byte(4))) // U+0004 = EOT
	if err != nil {
		return "", err
	}
	p.Inputs = append(p.Inputs, input)
	output, err := p.Out.ReadString(byte(4)) // U+0004 = EOT
	output = strings.TrimSpace(output[:len(output)-1])
	if err != nil {
		return "", err
	}
	p.Outputs = append(p.Outputs, output)
	return output, nil
}

func (p *Subproc) Close() {
	p.In.Write([]byte{byte(3)}) // U+0003 = ETX
	p.Cmd.Wait()
}
