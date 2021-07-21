package actionsdk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var (
	// args represents args that have been unmarshalled for the given action.
	// This only happens once and is read-only, therefore it's safe to keep this
	// in a single package-level variable.
	//
	// If args is nil, args has not yet been initialized.
	args *Args
)

type Args struct {
	ArgsVersion int
	Metadata    json.RawMessage
	Baggage     Baggage
}

type Baggage struct {
	EventWrapper EventWrapper `json:"WorkspaceEvent"`
	Actions      map[uint]map[string]interface{}
}

type EventWrapper struct {
	Event Event
}

type Event struct {
	Name      string                 `json:"name"`
	Data      map[string]interface{} `json:"data"`
	User      map[string]interface{} `json:"user,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Timestamp int64                  `json:"ts,omitempty"`
	Version   string                 `json:"v,omitempty"`
}

// WriteError writes an error to stdout with a standard format.  The error is
// added to a json object with an "error" key: {"error": err.Error()}.
//
// This does _not_ stop the action or workflow.
//
// To stop the action and prevent the workflow branch from continuing, exit
// with a non-zero status code (ie. `os.Exit(1)`).
//
// To stop the action but allow workflows to continue, exit with a zero status
// code (ie. `os.Exit(0)`)
func WriteError(err error) {
	byt, err := json.Marshal(map[string]interface{}{"error": err.Error()})
	if err != nil {
		log.Fatal(fmt.Errorf("unable to marshal error: %w", err))
	}

	_, err = fmt.Fprint(os.Stdout, string(byt))
	if err != nil {
		log.Fatal(fmt.Errorf("unable to write error: %w", err))
	}
}

// WriteResult writes the output as a JSON-encoded string to stdout.  Any data written
// here is captured as action output, which is added to the workflow context and can be
// used by future actions in the workflow.
//
// Even though this can be called many times the engine only supports one JSON-encoded
// object, so you really only want to write once.  This may be enforced in future versions
// of this SDK, and writing more than once may produce an error in the future.
//
// Note that this does _not_ stop the action.  To stop the action, call `os.Exit(0)` or
// return from your main function.
func WriteResult(i interface{}) error {
	if i == nil {
		_, err := fmt.Fprint(os.Stdout, "{}")
		return err
	}

	byt, err := json.Marshal(i)
	if err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}
	_, err = fmt.Fprint(os.Stdout, string(byt))
	return err
}

// GetMetadata returns the metadata for the action as configured within this specific workflow.
// The type for this struct must match the definitions within the action config (action.cue).
func GetMetadata(dest interface{}) error {
	args, err := GetArgs()
	if err != nil {
		return err
	}
	return json.Unmarshal(args.Metadata, dest)
}

// GetSecret returns the secret stored within the current workspace.  If no secret is found
// this returns an error.
func GetSecret(str string) (string, error) {
	if secret := os.Getenv(str); secret != "" {
		return secret, nil
	}
	return "", fmt.Errorf("secret not found: %s", str)
}

func GetArgs() (*Args, error) {
	if args != nil {
		return args, nil
	}

	// We pass in a JSON string as the first arugment.  This payload contains the action metadata,
	// workflow context, etc.
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("no arguments present")
	}

	args = &Args{}
	if err := json.Unmarshal([]byte(os.Args[1]), args); err != nil {
		return nil, fmt.Errorf("unable to parse arguments: %s", err)
	}

	return args, nil
}
