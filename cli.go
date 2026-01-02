package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func printUsage() {
	fmt.Print(`
  TGO - Task CLI Manager
  ----------------------

Usage:
  tgo                      - Interactive task management
  tgo start <number>       - Start/stop task timer
  tgo done <number>        - Mark task complete
  tgo set-dir <path>    - Configure task directory
  tgo create-list <name>   - Create new task list
  tgo remove-list          - Remove task list
  tgo help                 - Show this help

Interactive Commands:
  <number>        - Start/stop task timer
  add <task>      - Add new task
  remove <number> - Remove task
  done <number>   - Mark task complete
  r | return      - Return to main menu
  q | quit        - Exit program

Examples:
  tgo set-dir ~/Tasks
  tgo create-list "Sprint Planning"
  tgo
  tgo start 3
`)
}

func runCLI() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		runInteractiveMode(config)
		return
	}

	command := os.Args[1]
	switch command {
	case "set-dir":
		handleSetFolder(config)
	case "create-list":
		handleCreateList(config)
	case "remove-list":
		handleRemoveList(config)
	case "start":
		handleStartTask(config)
	case "done":
		handleMarkDone(config)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("[!] Unknown command: %s\n", command)
		printUsage()
	}
}

func runInteractiveMode(config *Config) {
	if config.TaskDir == "" {
		fmt.Println("[!] No task directory configured")
		fmt.Println("Use: tgo set-dir <path>")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("[i] No task lists found in: %s\n\n", config.TaskDir)
		fmt.Println("Let's create your first task list!")
		handleCreateFirstList(config)
		return
	}

	taskFile, err := selectTaskFile(config.TaskDir, taskFiles)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	taskList, err := loadTasks(taskFile)
	if err != nil {
		fmt.Printf("[!] Error loading tasks: %v\n", err)
		return
	}

	runInteractiveLoop(taskList, taskFile)
}

func runInteractiveLoop(taskList *TaskList, taskFile string) {
	reader := bufio.NewReader(os.Stdin)
	spinnerStart := time.Now()

	render := func() {
		frame := int(time.Since(spinnerStart) / (150 * time.Millisecond))
		displayTaskListWithSpinner(taskList, filepath.Base(taskFile), frame)
		fmt.Print("\n> ")
	}

	for {
		render()
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if handleInteractiveCommand(input, taskList, taskFile) {
			return
		}

		// restart spinner cycle after each command
		spinnerStart = time.Now()
	}
}

func handleInteractiveCommand(input string, taskList *TaskList, taskFile string) bool {
	switch {
	case input == "q" || input == "quit" || input == "exit":
		return true
	case input == "r" || input == "return":
		runCLI()
		return true
	case strings.HasPrefix(input, "add "):
		handleAddTask(strings.TrimSpace(input[4:]), taskList, taskFile)
	case strings.HasPrefix(input, "a "):
		handleAddTask(strings.TrimSpace(input[2:]), taskList, taskFile)
	case strings.HasPrefix(input, "remove "):
		handleRemoveTask(strings.TrimSpace(input[7:]), taskList, taskFile)
	case strings.HasPrefix(input, "r "):
		handleRemoveTask(strings.TrimSpace(input[2:]), taskList, taskFile)
	case strings.HasPrefix(input, "done "):
		handleDoneTask(strings.TrimSpace(input[5:]), taskList, taskFile)
	case strings.HasPrefix(input, "d "):
		handleDoneTask(strings.TrimSpace(input[2:]), taskList, taskFile)
	default:
		if taskNum, err := strconv.Atoi(input); err == nil {
			handleToggleTimer(taskNum, taskList, taskFile)
		} else {
			fmt.Println("[!] Invalid command. Type a number, 'add / a <task>', 'remove / r <number>', 'done / d <number>', 'r' to return, or 'q' to quit")
		}
	}
	return false
}

func handleAddTask(taskTitle string, taskList *TaskList, taskFile string) {
	if taskTitle == "" {
		fmt.Println("[!] Task title cannot be empty")
		return
	}

	addTask(taskList, taskTitle)
	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
	}
}

func handleRemoveTask(taskNumStr string, taskList *TaskList, taskFile string) {
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		fmt.Printf("[!] '%s' is not a valid number\n", taskNumStr)
		return
	}

	if err := removeTask(taskList, taskNum); err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
	}
}

func handleDoneTask(taskNumStr string, taskList *TaskList, taskFile string) {
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		fmt.Printf("[!] '%s' is not a valid number\n", taskNumStr)
		return
	}

	if err := markTaskComplete(taskList, taskNum); err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
	}
}

func handleToggleTimer(taskNum int, taskList *TaskList, taskFile string) {
	if err := toggleTaskTimer(taskList, taskNum); err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
	}
}

func handleSetFolder(config *Config) {
	if len(os.Args) < 3 {
		fmt.Println("[!] Folder path required")
		return
	}

	folder := os.Args[2]
	if strings.HasPrefix(folder, "~/") {
		home, _ := os.UserHomeDir()
		folder = filepath.Join(home, folder[2:])
	}

	absDir, err := filepath.Abs(folder)
	if err != nil {
		fmt.Printf("[!] Invalid path: %v\n", err)
		return
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Printf("[!] Directory not found: %s\n", absDir)
		return
	}

	config.TaskDir = absDir
	if err := saveConfig(config); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
		return
	}

	fmt.Printf("[+] Task directory set: %s\n", absDir)
	showDirContents(absDir)
}

func handleCreateList(config *Config) {
	if config.TaskDir == "" {
		fmt.Println("[!] No task directory configured")
		fmt.Println("Use: tgo set-dir <path>")
		return
	}

	var listName string
	if len(os.Args) < 3 {
		fmt.Print("Enter list name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			listName = strings.TrimSpace(scanner.Text())
		}
		if listName == "" {
			fmt.Println("[!] List name cannot be empty")
			return
		}
	} else {
		listName = strings.Join(os.Args[2:], " ")
	}

	if err := createNewList(config.TaskDir, listName); err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	fmt.Printf("[+] Created list: %s\n", listName)
	showDirContents(config.TaskDir)
}

func handleRemoveList(config *Config) {
	if config.TaskDir == "" {
		fmt.Println("[!] No task directory configured")
		fmt.Println("Use: tgo set-dir <path>")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
		showDirContents(config.TaskDir)
		return
	}

	if len(taskFiles) == 1 {
		fmt.Printf("Remove '%s'? (y/N): ", taskFiles[0])
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
			if err := os.Remove(filepath.Join(config.TaskDir, taskFiles[0])); err != nil {
				fmt.Printf("[!] Failed to remove: %v\n", err)
				return
			}
			fmt.Printf("[-] Removed: %s\n", taskFiles[0])
		}
		return
	}

	fmt.Printf("[i] Found %d task lists:\n\n", len(taskFiles))
	for i, file := range taskFiles {
		displayName := strings.TrimSuffix(file, ".json")
		fmt.Printf("%d. %s\n", i+1, displayName)
	}

	fmt.Printf("\nSelect list to remove (1-%d): ", len(taskFiles))
	var choice int
	if _, err := fmt.Scanf("%d", &choice); err != nil || choice < 1 || choice > len(taskFiles) {
		fmt.Println("[!] Invalid selection")
		return
	}

	selectedFile := taskFiles[choice-1]
	fmt.Printf("Remove '%s'? (y/N): ", selectedFile)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
		if err := os.Remove(filepath.Join(config.TaskDir, selectedFile)); err != nil {
			fmt.Printf("[!] Failed to remove: %v\n", err)
			return
		}
		fmt.Printf("[-] Removed: %s\n", selectedFile)
	}
}

func handleStartTask(config *Config) {
	if len(os.Args) < 3 {
		fmt.Println("[!] Task number required")
		return
	}

	taskNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("[!] '%s' is not a valid number\n", os.Args[2])
		return
	}

	if config.TaskDir == "" {
		fmt.Println("[!] No task directory configured")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	taskFile, err := selectTaskFile(config.TaskDir, taskFiles)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	taskList, err := loadTasks(taskFile)
	if err != nil {
		fmt.Printf("[!] Load error: %v\n", err)
		return
	}

	if err := toggleTaskTimer(taskList, taskNum); err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
	}
}

func handleMarkDone(config *Config) {
	if len(os.Args) < 3 {
		fmt.Println("[!] Task number required")
		return
	}

	taskNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("[!] '%s' is not a valid number\n", os.Args[2])
		return
	}

	if config.TaskDir == "" {
		fmt.Println("[!] No task directory configured")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	taskFile, err := selectTaskFile(config.TaskDir, taskFiles)
	if err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	taskList, err := loadTasks(taskFile)
	if err != nil {
		fmt.Printf("[!] Load error: %v\n", err)
		return
	}

	if err := markTaskComplete(taskList, taskNum); err != nil {
		fmt.Printf("[!] %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("[!] Save error: %v\n", err)
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func handleCreateFirstList(config *Config) {
	fmt.Print("Enter your first list name: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		listName := strings.TrimSpace(scanner.Text())
		if listName == "" {
			fmt.Println("[!] List name cannot be empty")
			return
		}

		if err := createNewList(config.TaskDir, listName); err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		fmt.Printf("[+] Created your first list: %s\n", listName)
		fmt.Println("[>] Starting interactive mode...")
		time.Sleep(time.Second)
		runInteractiveMode(config)
	}
}
