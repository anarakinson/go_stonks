package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("cmd", "/c", "echo Привет, мир!")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Ошибка при выполнении команды: %v\n", err)
	}
	fmt.Print(string(output))
}
