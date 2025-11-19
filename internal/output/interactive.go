package output

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

type Action int

const (
	ActionCopy Action = iota
	ActionRun
	ActionEdit
	ActionSkip
)

func PromptCodeBlock(block CodeBlock, total int) Action {
	fmt.Fprintf(os.Stderr, "\n┌─ Command %d/%d (%s) ", block.Index, total, block.Language)
	fmt.Fprintf(os.Stderr, "%s", strings.Repeat("─", 50-len(fmt.Sprintf("Command %d/%d (%s) ", block.Index, total, block.Language))))
	fmt.Fprintf(os.Stderr, "\n│ %s\n", strings.ReplaceAll(block.Code, "\n", "\n│ "))
	fmt.Fprintf(os.Stderr, "└")
	fmt.Fprintf(os.Stderr, "%s", strings.Repeat("─", 63))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "[c] Copy  [r] Run  [e] Edit & Run  [s] Skip: ")

	if err := keyboard.Open(); err != nil {
		return readFallback()
	}
	defer keyboard.Close()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return ActionSkip
		}

		if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC {
			fmt.Fprintln(os.Stderr, "\nCancelled")
			os.Exit(0)
		}

		switch char {
		case 'c', 'C':
			fmt.Fprintln(os.Stderr, "c")
			return ActionCopy
		case 'r', 'R':
			fmt.Fprintln(os.Stderr, "r")
			return ActionRun
		case 'e', 'E':
			fmt.Fprintln(os.Stderr, "e")
			return ActionEdit
		case 's', 'S':
			fmt.Fprintln(os.Stderr, "s")
			return ActionSkip
		}
	}
}

func readFallback() Action {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "c":
		return ActionCopy
	case "r":
		return ActionRun
	case "e":
		return ActionEdit
	default:
		return ActionSkip
	}
}
