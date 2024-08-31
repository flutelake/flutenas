package node

import "fmt"

func Shell(cmd string) ([]byte, error) {
	exec := NewExec().SetHost("10.0.1.10")
	defer exec.Close()

	output, err := exec.Command(cmd)
	if err != nil {
		return nil, fmt.Errorf("exec error: %s", err)
	}

	return output, nil
}
