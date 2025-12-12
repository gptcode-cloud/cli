package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// RunREPL implements a REPL for command execution with follow-up support
type RunREPL struct {
	history *CommandHistory
	prompt  string
}

// NewRunREPL creates a new run REPL instance
func NewRunREPL(maxCommands int) *RunREPL {
	return &RunREPL{
		history: NewCommandHistory(maxCommands),
		prompt:  "> ",
	}
}

// Run starts the REPL loop
func (r *RunREPL) Run() error {
	fmt.Println("GPTCode Run REPL - Type /help for commands")
	fmt.Println("")

	// Initialize current directory
	if wd, err := os.Getwd(); err == nil {
		r.history.SetDirectory(wd)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Show prompt with current directory
		if r.history.CurrentDir != "" {
			fmt.Printf("%s:%s$ ", r.history.CurrentDir, strings.TrimSuffix(os.Args[0], "/chu"))
		} else {
			fmt.Print(r.prompt)
		}

		// Read input
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		// Trim whitespace
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle special commands starting with '/'
		if strings.HasPrefix(line, "/") {
			shouldContinue, shouldExit := r.handleCommand(line)
			if shouldExit {
				break
			}
			if shouldContinue {
				continue
			}
		}

		// Process command
		r.executeCommand(line)
	}

	return nil
}

// handleCommand processes REPL commands
func (r *RunREPL) handleCommand(cmd string) (bool, bool) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false, false
	}

	switch parts[0] {
	case "/exit", "/quit":
		fmt.Println("Goodbye!")
		return false, true

	case "/help":
		r.showHelp()
		return true, false

	case "/history":
		fmt.Print(r.history.String())
		return true, false

	case "/output":
		if len(parts) < 2 {
			fmt.Println("Usage: /output <command-id>")
			r.showRecentCommands()
		} else {
			if cmd, err := r.history.GetCommand(parts[1]); err != nil {
				fmt.Printf("Command not found: %s\n", parts[1])
			} else {
				if cmd.Error != "" {
					fmt.Printf("Command [%s] failed (exit %d):\n", cmd.ID, cmd.ExitCode)
					fmt.Printf("Command: %s\n", cmd.Command)
					fmt.Printf("Error: %s\n", cmd.Error)
				} else {
					fmt.Printf("Command [%s] succeeded:\n", cmd.ID)
					fmt.Printf("Command: %s\n", cmd.Command)
					if cmd.Output != "" {
						fmt.Printf("Output:\n%s\n", cmd.Output)
					}
				}
			}
		}
		return true, false

	case "/cd":
		if len(parts) < 2 {
			fmt.Println("Usage: /cd <directory>")
			if r.history.CurrentDir != "" {
				fmt.Printf("Current directory: %s\n", r.history.CurrentDir)
			}
		} else {
			if err := os.Chdir(parts[1]); err != nil {
				fmt.Printf("Failed to change directory: %v\n", err)
			} else {
				if wd, err := os.Getwd(); err == nil {
					r.history.SetDirectory(wd)
					fmt.Printf("Changed to: %s\n", wd)
				}
			}
		}
		return true, false

	case "/env":
		if len(parts) == 1 {
			fmt.Println("Environment variables:")
			for k, v := range r.history.Environment {
				fmt.Printf("  %s=%s\n", k, v)
			}
		} else if len(parts) == 2 && strings.Contains(parts[1], "=") {
			// Set environment variable
			envParts := strings.SplitN(parts[1], "=", 2)
			if len(envParts) == 2 {
				key, value := envParts[0], envParts[1]
				r.history.SetEnvironment(key, value)
				fmt.Printf("Set %s=%s\n", key, value)
			} else {
				fmt.Println("Usage: /env <key>=<value>")
			}
		} else if len(parts) == 2 {
			// Display specific env var
			value := r.history.GetEnvironment(parts[1])
			if value != "" {
				fmt.Printf("%s=%s\n", parts[1], value)
			} else {
				fmt.Printf("%s is not set\n", parts[1])
			}
		} else {
			fmt.Println("Usage: /env [key] or /env <key>=<value>")
		}
		return true, false

	default:
		fmt.Printf("Unknown command: %s (type /help for available commands)\n", parts[0])
		return true, false
	}
}

// executeCommand runs a shell command and captures its output
func (r *RunREPL) executeCommand(cmdStr string) {
	// Expand $last and $N references
	cmdStr = r.expandReferences(cmdStr)

	// Create command with current directory
	cmd := exec.Command("sh", "-c", cmdStr)

	// Set up environment
	cmd.Env = os.Environ()
	for k, v := range r.history.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set working directory
	if r.history.CurrentDir != "" {
		cmd.Dir = r.history.CurrentDir
	}

	// Capture output
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			fmt.Printf("Error executing command: %v\n", err)
			return
		}
	}

	// Add to history
	r.history.AddCommand(cmdStr, output, "", exitCode)

	// Display output
	if output != "" {
		fmt.Print(output)
	}

	// Show exit code if non-zero
	if exitCode != 0 {
		fmt.Printf("(exit code: %d)\n", exitCode)
	}
}

// expandReferences replaces $last and $N with previous command outputs
func (r *RunREPL) expandReferences(cmdStr string) string {
	result := cmdStr

	// Replace $last with the most recent command's output
	if lastCmd := r.history.GetLastCommand(); lastCmd != nil {
		result = strings.ReplaceAll(result, "$last", lastCmd.Command)
	}

	// Replace $N with command by ID
	for _, id := range r.history.GetCommandsList() {
		if cmd, err := r.history.GetCommand(id); err == nil {
			placeholder := "$" + id
			result = strings.ReplaceAll(result, placeholder, cmd.Command)
		}
	}

	return result
}

// showHelp displays available REPL commands
func (r *RunREPL) showHelp() {
	fmt.Println("REPL Commands:")
	fmt.Println("  /exit, /quit   - Exit the run session")
	fmt.Println("  /help          - Show this help")
	fmt.Println("  /history       - Show command history")
	fmt.Println("  /output <id>   - Show output of command ID")
	fmt.Println("  /cd <dir>      - Change directory")
	fmt.Println("  /env           - Show/set environment variables")
	fmt.Println("  /env <key>     - Show specific environment variable")
	fmt.Println("  /env <k>=<v>   - Set environment variable")
	fmt.Println("")
	fmt.Println("Commands can reference previous outputs:")
	fmt.Println("  $last          - Reference the last command")
	fmt.Println("  $1, $2, ...    - Reference command by ID")
	fmt.Println("")
	fmt.Println("All other input will be executed as a shell command.")
}

// showRecentCommands shows recent commands for reference
func (r *RunREPL) showRecentCommands() {
	recent := r.history.GetLastN(5)
	if len(recent) == 0 {
		fmt.Println("No commands executed yet.")
		return
	}

	fmt.Println("Recent commands:")
	for _, cmd := range recent {
		fmt.Printf("  [%s] %s", cmd.ID, cmd.Command)
		if cmd.ExitCode != 0 {
			fmt.Printf(" (failed)")
		}
		fmt.Println()
	}
}

// RunSingleShotCommand executes a single command without REPL mode
func RunSingleShotCommand(command string) error {
	// Create a temporary REPL and execute one command
	repl := NewRunREPL(10)
	repl.executeCommand(command)
	return nil
}
