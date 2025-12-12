package output

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/atotto/clipboard"
)

func HandleCodeBlock(action Action, code string) error {
	switch action {
	case ActionCopy:
		if err := clipboard.WriteAll(code); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to copy: %v\n", err)
			return err
		}
		fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")

	case ActionRun:
		fmt.Fprintln(os.Stderr, "\n✓ Running...")
		cmd := exec.Command("sh", "-c", code)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n✗ Command failed: %v\n", err)
			return err
		}

	case ActionEdit:
		edited, err := openInEditor(code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to open editor: %v\n", err)
			return err
		}

		fmt.Fprintf(os.Stderr, "\nEdited command:\n%s\n\nRun this? [y/N]: ", edited)
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			cmd := exec.Command("sh", "-c", edited)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n✗ Command failed: %v\n", err)
				return err
			}
		}

	case ActionSkip:
	}

	return nil
}

func openInEditor(content string) (string, error) {
	tmpfile, err := os.CreateTemp("", "gptcode-*.sh")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}
	tmpfile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	edited, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return string(edited), nil
}
