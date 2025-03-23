package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

const (
	currentUser = "batazor"
)

type App struct {
	ctx       context.Context
	repo      repository.StateRepository
	loader    config.UseCaseLoader
	evaluator config.TriggerEvaluator
	executor  executor.UseCaseExecutor
	gameFSM   *fsm.GameFSM
	state     *domain.State
}

func NewApp() (*App, error) {
	ctx := context.Background()

	a := &App{
		ctx:       ctx,
		repo:      repository.NewFileStateRepository("db/state.yaml"),
		loader:    config.NewUseCaseLoader("usecases"),
		evaluator: config.NewTriggerEvaluator(),
		executor:  executor.NewUseCaseExecutor(),
		gameFSM:   fsm.NewGameFSM(),
	}

	// Try to load initial state
	state, err := a.repo.LoadState(a.ctx)
	if err != nil {
		// Create directories if they don't exist
		os.MkdirAll("db", 0755)

		// Initialize with default state if file doesn't exist
		a.state = &domain.State{
			Accounts: []domain.Account{
				{
					Email: "default@example.com",
					Characters: []domain.Gamer{
						{
							ID:       1,
							Nickname: "DefaultCharacter",
							Power:    100,
							Resources: domain.Resources{
								Wood: 100,
								Food: 100,
								Iron: 100,
								Meat: 100,
							},
							VIPLevel: 0,
							Heroes: domain.HeroesState{
								State: domain.HeroesStatus{
									IsHeroes: false,
								},
							},
							Messages: domain.MessagesState{
								State: domain.MessageStatus{
									IsNewMessage: false,
									IsNewReports: false,
								},
							},
							Alliance: domain.Alliance{
								State: domain.AllianceState{
									IsAlliance: false,
								},
							},
							Buildings: domain.Buildings{
								Items: make(map[string]domain.Building),
							},
							Researchs: domain.Researchs{
								Battle:  domain.Research{Level: 0},
								Economy: domain.Research{Level: 0},
							},
						},
					},
				},
			},
		}
		// Save the initial state
		if err := a.repo.SaveState(a.ctx, a.state); err != nil {
			return nil, fmt.Errorf("failed to save initial state: %w", err)
		}
	} else {
		a.state = state
	}

	return a, nil
}

func (a *App) Run() error {
	return a.interactiveMode()
}

func (a *App) listUsecases() error {
	usecases, err := a.loader.LoadAll(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load usecases: %w", err)
	}

	fmt.Println("Available Usecases:")

	for i, u := range usecases {
		fmt.Printf("%d) %s (Start Node: %s, Final Node: %s)\n",
			i+1, u.Name, u.Node, u.FinalNode)

		// Display steps summary
		if len(u.Steps) > 0 {
			fmt.Printf("   Steps: %d\n", len(u.Steps))
		}
	}

	fmt.Printf("\nFound %d usecases\n", len(usecases))

	return nil
}

func (a *App) runUsecase(ucIndex, charIndex int) error {
	// Validate character index
	if charIndex < 0 || len(a.allCharacters()) <= charIndex {
		return fmt.Errorf("character index out of range")
	}
	char := a.allCharacters()[charIndex]

	// Load usecases
	usecases, err := a.loader.LoadAll(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load usecases: %w", err)
	}

	if ucIndex < 0 || ucIndex >= len(usecases) {
		return fmt.Errorf("usecase index out of range")
	}

	usecase := usecases[ucIndex]

	fmt.Printf("Running usecase: %s for character %s (ID: %d)\n",
		usecase.Name, char.Nickname, char.ID)
	fmt.Printf("Start node: %s, Final node: %s\n", usecase.Node, usecase.FinalNode)

	// Execute each step in the usecase
	for i, step := range usecase.Steps {
		fmt.Printf("Step %d/%d: Action type: %T\n", i+1, len(usecase.Steps), step)

		// Here you would call the executor to execute the step
		// For example: a.executor.Execute(a.ctx, a.state, step, char)

		// Simulate step execution time
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("Usecase completed. Final node: %s\n", usecase.FinalNode)

	// Save updated state
	if err := a.repo.SaveState(a.ctx, a.state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

func (a *App) showState() error {
	// Reload state from disk to ensure we have the latest
	state, err := a.repo.LoadState(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	a.state = state

	fmt.Println("Current Game State:")
	fmt.Printf("Accounts: %d\n", len(a.state.Accounts))

	totalChars := 0
	for _, acc := range a.state.Accounts {
		totalChars += len(acc.Characters)
	}
	fmt.Printf("Characters: %d\n", totalChars)

	return nil
}

func (a *App) listCharacters() error {
	chars := a.allCharacters()

	fmt.Println("All Characters:")
	fmt.Println("--------------------------------------")
	for i, char := range chars {
		fmt.Printf("%d) %s (ID: %d)\n", i+1, char.Nickname, char.ID)
		fmt.Printf("   Power: %d | VIP: %d | Alliance: %s\n",
			char.Power, char.VIPLevel, char.Alliance.Name)
		fmt.Printf("   Resources: Wood: %d | Food: %d | Iron: %d | Meat: %d\n",
			char.Resources.Wood, char.Resources.Food, char.Resources.Iron, char.Resources.Meat)
		fmt.Printf("   Research: Battle: %d | Economy: %d\n",
			char.Researchs.Battle.Level, char.Researchs.Economy.Level)
		fmt.Println("--------------------------------------")
	}

	return nil
}

func (a *App) showCharacter(idx int) error {
	chars := a.allCharacters()

	if idx < 0 || idx >= len(chars) {
		return fmt.Errorf("character index out of range")
	}

	char := chars[idx]

	fmt.Printf("Character Details: %s (ID: %d)\n", char.Nickname, char.ID)
	fmt.Println("======================================")
	fmt.Printf("Power: %d\n", char.Power)
	fmt.Printf("VIP Level: %d\n", char.VIPLevel)
	fmt.Printf("Gems: %d\n", char.Gems)

	fmt.Println("\nResources:")
	fmt.Printf("  Wood: %d\n", char.Resources.Wood)
	fmt.Printf("  Food: %d\n", char.Resources.Food)
	fmt.Printf("  Iron: %d\n", char.Resources.Iron)
	fmt.Printf("  Meat: %d\n", char.Resources.Meat)

	fmt.Println("\nAlliance:")
	if char.Alliance.State.IsAlliance {
		fmt.Printf("  Name: %s\n", char.Alliance.Name)
		fmt.Printf("  Power: %d\n", char.Alliance.Power)
		fmt.Printf("  Members: %d/%d\n", char.Alliance.Members.Count, char.Alliance.Members.Max)
		fmt.Printf("  Wars: %d\n", char.Alliance.State.IsWar)
		fmt.Printf("  Chests: %d\n", char.Alliance.State.IsChests)
	} else {
		fmt.Println("  Not in an alliance")
	}

	fmt.Println("\nBuildings:")
	if len(char.Buildings.Items) > 0 {
		for buildingType, building := range char.Buildings.Items {
			fmt.Printf("  %s - Level: %d, Power: %d\n", buildingType, building.Level, building.Power)
		}
	} else {
		fmt.Println("  No buildings")
	}

	fmt.Println("\nResearch:")
	fmt.Printf("  Battle: Level %d\n", char.Researchs.Battle.Level)
	fmt.Printf("  Economy: Level %d\n", char.Researchs.Economy.Level)

	return nil
}

func (a *App) resetState() error {
	// Initialize with default state
	a.state = &domain.State{
		Accounts: []domain.Account{
			{
				Email: "default@example.com",
				Characters: []domain.Gamer{
					{
						ID:       1,
						Nickname: "DefaultCharacter",
						Power:    100,
						Resources: domain.Resources{
							Wood: 100,
							Food: 100,
							Iron: 100,
							Meat: 100,
						},
						VIPLevel: 0,
						Heroes: domain.HeroesState{
							State: domain.HeroesStatus{
								IsHeroes: false,
							},
						},
						Messages: domain.MessagesState{
							State: domain.MessageStatus{
								IsNewMessage: false,
								IsNewReports: false,
							},
						},
						Alliance: domain.Alliance{
							State: domain.AllianceState{
								IsAlliance: false,
							},
						},
						Buildings: domain.Buildings{
							Items: make(map[string]domain.Building),
						},
						Researchs: domain.Researchs{
							Battle:  domain.Research{Level: 0},
							Economy: domain.Research{Level: 0},
						},
					},
				},
			},
		},
	}

	// Save the reset state
	if err := a.repo.SaveState(a.ctx, a.state); err != nil {
		return fmt.Errorf("failed to save reset state: %w", err)
	}

	fmt.Println("Game state has been reset to default.")
	return nil
}

func (a *App) interactiveMode() error {
	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		// Display header with exactly the specified format
		fmt.Printf("Current Date and Time (UTC - YYYY-MM-DD HH:MM:SS formatted): %s\n",
			time.Now().UTC().Format("2006-01-02 15:04:05"))
		fmt.Printf("Current User's Login: %s\n", currentUser)
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  l - List usecases")
		fmt.Println("  r <usecase> <character> - Run usecase for character")
		fmt.Println("  c - List all characters")
		fmt.Println("  v <index> - View character details")
		fmt.Println("  s - Show current state")
		fmt.Println("  x - Reset state")
		fmt.Println("  q - Quit")
		fmt.Print("> ")

		// Read user input
		var input string
		fmt.Scanln(&input)
		args := strings.Fields(input)

		if len(args) == 0 {
			continue
		}

		command := args[0]

		switch command {
		case "l":
			if err := a.listUsecases(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "r":
			if len(args) < 3 {
				fmt.Println("Usage: r <usecase_index> <character_index>")
				continue
			}

			ucIdx, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Printf("Error: Invalid usecase index: %v\n", err)
				continue
			}

			charIdx, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Printf("Error: Invalid character index: %v\n", err)
				continue
			}

			if err := a.runUsecase(ucIdx-1, charIdx-1); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "c":
			if err := a.listCharacters(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "v":
			if len(args) < 2 {
				fmt.Println("Usage: v <character_index>")
				continue
			}

			idx, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Printf("Error: Invalid character index: %v\n", err)
				continue
			}

			if err := a.showCharacter(idx - 1); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "s":
			if err := a.showState(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "x":
			fmt.Print("Are you sure you want to reset state? (y/n): ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm == "y" || confirm == "Y" {
				if err := a.resetState(); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				fmt.Println("Reset cancelled")
			}
		case "q":
			fmt.Println("Exiting...")
			return nil
		default:
			fmt.Println("Unknown command")
		}

		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
	}
}

// Helper function to get all characters across accounts
func (a *App) allCharacters() []domain.Gamer {
	var characters []domain.Gamer

	for _, acc := range a.state.Accounts {
		characters = append(characters, acc.Characters...)
	}

	return characters
}
