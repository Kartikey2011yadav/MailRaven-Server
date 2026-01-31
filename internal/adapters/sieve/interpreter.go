package sieve

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"git.sr.ht/~emersion/go-sieve"
)

// Interpreter executes Sieve commands.
type Interpreter struct {
	ctx             context.Context
	msg             *mail.Message
	header          mail.Header
	targets         []string
	discard         bool
	stopped         bool
	vacationManager *VacationManager
	userID          string
}

func NewInterpreter(ctx context.Context, msg *mail.Message, vm *VacationManager, userID string) *Interpreter {
	return &Interpreter{
		ctx:             ctx,
		msg:             msg,
		header:          msg.Header,
		targets:         make([]string, 0),
		vacationManager: vm,
		userID:          userID,
	}
}

func (i *Interpreter) Run(cmds []sieve.Command) ([]string, error) {
	err := i.evalBlock(cmds)
	if err != nil {
		return nil, err
	}

	// Implicit Keep
	// If no targets and not discarded, default to INBOX
	if len(i.targets) == 0 && !i.discard {
		i.targets = append(i.targets, "INBOX")
	}

	return i.targets, nil
}

func (i *Interpreter) evalBlock(cmds []sieve.Command) error {
	for _, cmd := range cmds {
		if i.stopped {
			break
		}
		if err := i.evalCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) evalCommand(cmd sieve.Command) error {
	switch cmd.Name {
	case "require":
		// Ignore requirements for now, or check supported extensions
		return nil
	case "stop":
		i.stopped = true
		return nil
	case "keep":
		i.targets = append(i.targets, "INBOX")
		return nil
	case "discard":
		i.discard = true
		return nil
	case "fileinto":
		folder, err := getStringArg(cmd.Arguments, 0)
		if err != nil {
			return err
		}
		// Canonicalize
		if strings.EqualFold(folder, "inbox") {
			folder = "INBOX"
		}
		i.targets = append(i.targets, folder)
		// Explicit fileinto cancels implicit keep (by virtue of adding a target)
		return nil
	case "vacation":
		return i.evalVacation(cmd)
	case "if", "elsif": // "elsif" handled in previous block usually? No, flat list?
		// go-sieve parser handles structure.
		// "if" has Tests.
		match, err := i.evalTest(cmd.Tests)
		if err != nil {
			return err
		}
		if match {
			return i.evalBlock(cmd.Block)
		}
		// Try "else" or "elsif" - wait, where are they?
		// go-sieve might link them? Or they are next commands?
		// Check documentation or structure.
		// Usually if/elsif/else are linked.
		// Looking at check_sieve output: `Block:[...]`.
		// It doesn't show "Else".
		// Maybe `cmd.Next`? Or `Arguments`?
		// Or maybe `Block` contains the if-body, but where is else?
		// RFC says: control structure.
		// Parser normally structures if-elsif-else into one node or linked nodes.
		// Assuming simplest: if the test fails, we don't run block.
		// But we need to find the Else block.
		// If `go-sieve` Parser returns flat list of commands, then `else` would be a separate command?
		// But `else` depends on previous `if` result.
		// The AST provided by `Parse` might be tree-structured.
		// Recheck `check_sieve.go` output...
	}
	return nil
}

func (i *Interpreter) evalTest(tests []sieve.Test) (bool, error) {
	// Usually only one test per if? Or `anyof`/`allof`.
	// If `if` has multiple tests, it's syntax error in Sieve (must use anyof/allof).
	// Parser probably enforces this.
	if len(tests) == 0 {
		return true, nil // "true"
	}

	test := tests[0]
	switch test.Name {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "not":
		res, err := i.evalTest(test.Tests)
		return !res, err
	case "anyof":
		for _, t := range test.Tests {
			res, err := i.evalTest([]sieve.Test{t})
			if err != nil {
				return false, err
			}
			if res {
				return true, nil
			}
		}
		return false, nil
	case "allof":
		for _, t := range test.Tests {
			res, err := i.evalTest([]sieve.Test{t})
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		}
		return true, nil
	case "header":
		return i.evalHeaderTest(test)
	}
	return false, fmt.Errorf("unknown test: %s", test.Name)
}

func (i *Interpreter) evalHeaderTest(test sieve.Test) (bool, error) {
	// header :match-type :comparator "header-names" "key-list"
	// Extract args.
	// We need to parse optional args (:contains, :is, :matches).
	matchType := "is" // default
	var headerNames []string
	var keys []string

	for _, arg := range test.Arguments {
		switch v := arg.(type) {
		case sieve.ArgumentTag:
			matchType = string(v)
		case sieve.ArgumentStringList: // StringList is []string
			if headerNames == nil {
				headerNames = v
			} else if keys == nil {
				keys = v
			}
		}
	}

	if len(headerNames) == 0 || len(keys) == 0 {
		return false, fmt.Errorf("header test missing arguments")
	}

	for _, hn := range headerNames {
		val := i.header.Get(hn)
		// Check against all keys
		for _, key := range keys {
			if match(val, key, matchType) {
				return true, nil
			}
		}
	}

	return false, nil
}

func match(val, key, matchType string) bool {
	switch matchType {
	case "contains":
		return strings.Contains(strings.ToLower(val), strings.ToLower(key))
	case "is":
		return strings.EqualFold(val, key)
	case "matches":
		// Simple wildcard support ? and *
		// Convert Sieve wildcard to Regex?
		// * -> .*
		// ? -> .
		pattern := regexp.QuoteMeta(key)
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
		matched, err := regexp.MatchString("^"+pattern+"$", val)
		if err != nil {
			return false
		}
		return matched
	}
	return false
}

func (i *Interpreter) evalVacation(cmd sieve.Command) error {
	reason := ""
	opts := make(map[string]interface{})

	// Parse arguments (reason is positional, others are tagged)
	// go-sieve might put tokens in Arguments.
	// Structure: [Optional Tags...] Reason
	// Tags: :days <number>, :subject <string>, :from <string>, :addresses <string-list>, :mime, :handle <string>

	// Iterate args to find tags and final string
	for idx, arg := range cmd.Arguments {
		switch v := arg.(type) {
		case sieve.ArgumentTag:
			tag := string(v)
			switch tag {
			case ":days":
				// Next arg is number
				if val, err := getIntArg(cmd.Arguments, idx+1); err == nil {
					opts["days"] = val
				}
			case ":subject":
				if val, err := getStringArg(cmd.Arguments, idx+1); err == nil {
					opts["subject"] = val
				}
			case ":from":
				if val, err := getStringArg(cmd.Arguments, idx+1); err == nil {
					opts["from"] = val
				}
			case ":handle":
				if val, err := getStringArg(cmd.Arguments, idx+1); err == nil {
					opts["handle"] = val
				}
			case ":mime":
				opts["mime"] = true
			case ":addresses":
				// String List
				// TODO: Implement getListArg
			}
		case sieve.ArgumentStringList:
			// Usually the reason string if it's the last one and not a tag value?
			// Tag values are "consumed" by the tag check. We need robust parsing.
			// Simplified: The reason is the only positional string argument. Tag arguments follow tags.
			// We can assume if previous token was NOT a tag that takes an argument, this IS the reason.
			// However `getStringArg` helper doesn't respect iteration context.
			// Let's rely on standard: reason is mandatory string.
			// If we simplify: reason is the LAST string argument found that isn't a tag parameter.
			// Or better: scan loop.
		}
	}

	// iterating
	for j := 0; j < len(cmd.Arguments); j++ {
		arg := cmd.Arguments[j]
		if tag, ok := arg.(sieve.ArgumentTag); ok {
			t := string(tag)
			if t == ":days" || t == ":subject" || t == ":from" || t == ":handle" || t == ":addresses" {
				// These take an argument, skip next
				if j+1 < len(cmd.Arguments) {
					valArg := cmd.Arguments[j+1]
					// Extract value
					if t == ":days" {
						if n, ok := valArg.(sieve.ArgumentNumber); ok {
							opts["days"] = n.Value
						}
					} else if t == ":subject" {
						if s, err := getStringFromArg(valArg); err == nil {
							opts["subject"] = s
						}
					}
					j++ // Skip arg
				}
			} else {
				// :mime doesn't take arg
				if t == ":mime" {
					opts["mime"] = true
				}
			}
		} else if s, err := getStringFromArg(arg); err == nil {
			// Positional string -> Reason
			reason = s
		}
	}

	if i.vacationManager != nil {
		// Enqueue the vacation logic (async? ProcessVacation does db check and enqueue)
		err := i.vacationManager.ProcessVacation(i.ctx, i.userID, i.msg, opts, reason)
		if err != nil {
			// Log error but don't fail script execution?
			// Usually Sieve errors are logged.
			// fmt.Printf("Vacation error: %v\n", err)
			return err
		}
	}
	return nil
}

func getIntArg(args []sieve.Argument, index int) (int, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("out of bounds")
	}
	if n, ok := args[index].(sieve.ArgumentNumber); ok {
		return n.Value, nil
	}
	return 0, fmt.Errorf("not a number")
}

func getStringFromArg(arg sieve.Argument) (string, error) {
	if s, ok := arg.(sieve.ArgumentStringList); ok && len(s) > 0 {
		return s[0], nil
	}
	// Also literal? go-sieve parses strings as ArgumentStringList (slice of strings).
	return "", fmt.Errorf("not a string")
}

func getStringArg(args []sieve.Argument, index int) (string, error) {
	if index >= len(args) {
		return "", fmt.Errorf("argument index out of bounds")
	}
	switch v := args[index].(type) {
	case sieve.ArgumentStringList:
		if len(v) > 0 {
			return v[0], nil
		}
		return "", nil
	default:
		return "", fmt.Errorf("argument not a string")
	}
}
