package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kiishi/statemachine"
	"io"
	"os"
	"strings"
)

type Ruleset struct {
	NextState         string `json:"next_state"`
	ReplacementSymbol string `json:"replacement_symbol"`
	Direction         string `json:"direction"` // 'L' or 'R'
}

type TuringMachine struct {
	cursor      int
	InputString string
	Tape        []rune
	Controller  *statemachine.StateMachine
	AcceptState string //state name for the accept state
	RejectState string
	RuleSets map[string]map[string]Ruleset
}

func (t *TuringMachine) CurrentTapeValue() rune {
	if t.cursor >= len(t.InputString) {
		return '_'
	}
	return t.Tape[t.cursor]
}

func (t *TuringMachine) DecodeRuleSet(reader io.Reader) {
	decoder := json.NewDecoder(reader)
	var RuleSets map[string]map[string]Ruleset
	err := decoder.Decode(&RuleSets)
	if err != nil {
		panic("Invalid Json Format" + err.Error())
	}
	t.RuleSets = RuleSets
}

func (t *TuringMachine) ToRune(text string) rune {
	if len(text) > 1 {
		panic("Character cant be converted to rune ")
	}
	return []rune(text)[0]
}

func (t *TuringMachine) ProcessAndMoveCursor(character rune) {
	currentStateName := t.Controller.CurrentState.GetIdentifier()
	if val, ok := t.RuleSets[currentStateName]; ok {
		if characterConfiguration, ok := val[string(character)]; ok {
			if t.cursor >= len(t.InputString) {
				//treat as empty string _
				if characterConfiguration.Direction == "R" || characterConfiguration.Direction == "r" {
					//t.cursor--
					//send to reject state kos its going overboard
					err := t.Controller.SetState(t.RejectState)
					if err != nil {
						panic("Reject state invalid")
					}
				} else {
					t.cursor--
					t.Controller.Emit(currentStateName + characterConfiguration.NextState)
				}
				return
			}

			t.Tape[t.cursor] = t.ToRune(characterConfiguration.ReplacementSymbol)
			if (characterConfiguration.Direction == "L" || characterConfiguration.Direction == "l") && t.cursor != 0 {
				t.cursor--
			} else {
				t.cursor++
			}
			//	change state
			t.Controller.Emit(currentStateName + characterConfiguration.NextState)
		}
	}

}

//func (t *TuringMachine) UpdateCurrentS
type TuringMachineConfiguration struct {
	StateName            string
	ParentTurningMachine *TuringMachine
}

func (c *TuringMachineConfiguration) GetIdentifier() string {
	return c.StateName
}

func NewTuringConfiguration(parentMachine *TuringMachine, stateName string) *TuringMachineConfiguration {
	return &TuringMachineConfiguration{ParentTurningMachine: parentMachine, StateName: stateName}
}

//this takes in the states eg s1, s2, s3, s4
func (t *TuringMachine) InitializeController(states []string) {
	turingStateMachine := statemachine.NewMachine(&statemachine.Config{
		States:      nil,
		Transitions: nil,
	}, 0)

	for _, name := range states {
		state := NewTuringConfiguration(t, name)
		turingStateMachine.AddState(state)
	}

	//	converting ruleset to transitions
	for stateName, stateMaps := range t.RuleSets {
		for _, symbolRuleSet := range stateMaps {
			if stateName != symbolRuleSet.NextState {
				//	check if a transition rule exist between the 2 states
				transitionRule := statemachine.TransitionRule{
					EventName:        stateName + symbolRuleSet.NextState,
					CurrentState:     stateName,
					DestinationState: symbolRuleSet.NextState,
				}
				//this way a transition from s1 to s2 can be done by emitting "s1s2"
				turingStateMachine.AddTransition(&transitionRule)
			}
		}
	}

	t.Controller = turingStateMachine
}

func (t *TuringMachine) Prompt(prompt string, scanner *bufio.Scanner) string {
	fmt.Print(prompt)
	scanner.Scan()
	return scanner.Text()
}

func (t *TuringMachine) IsCompleted() bool {
	// we use x,y,z for tape symbols
	isComplete := true
	for _, val := range t.Tape {
		if val == 'x' || val == 'y' || val == 'z' || val == '_' {
			continue
		}
		isComplete = false
		break
	}
	return isComplete
}

func (t *TuringMachine) StartComputing() {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	fmt.Println("Loading rule set from ruleset.json...")
	file, err := os.Open("ruleset.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("missing ruleset.json")
			return
		}
		fmt.Println("Unknown error occurred")
		return
	}
	t.DecodeRuleSet(file)
	fmt.Println("Ruleset loaded completely")

	inputString := t.Prompt("Enter input string to test >>", scanner)

	t.Tape = []rune(inputString)
	t.InputString = inputString

	inputStates := t.Prompt("Enter States (eg 's1,s2,s3,sAccept,sReject') >>", scanner)

	inputStatesArr := strings.Split(inputStates, ",")

	t.AcceptState = t.Prompt("Which is the accept state? >>", scanner)

	t.RejectState = t.Prompt("Which is the reject state? >>", scanner)

	t.InitializeController(inputStatesArr)

	for {
		t.ProcessAndMoveCursor(t.CurrentTapeValue())
		if t.Controller.GetCurrentState().GetIdentifier() == t.AcceptState || t.Controller.GetCurrentState().GetIdentifier() == t.RejectState {
			break
		}
	}

	if t.Controller.GetCurrentState().GetIdentifier() == t.AcceptState {
		fmt.Println("String Accepted ✅")
		return
	}

	fmt.Println("String Rejected ❌")

}
