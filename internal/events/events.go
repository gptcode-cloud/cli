package events

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type EventType string

const (
	EventMessage     EventType = "message"
	EventStatus      EventType = "status"
	EventConfirm     EventType = "confirm"
	EventOpenFile    EventType = "open_file"
	EventOpenPlan    EventType = "open_plan"
	EventOpenSplit   EventType = "open_split"
	EventComplete    EventType = "complete"
	EventNotify      EventType = "notify"
)

type Event struct {
	Type EventType              `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type Emitter struct {
	writer io.Writer
}

func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{writer: w}
}

func (e *Emitter) Emit(eventType EventType, data map[string]interface{}) error {
	event := Event{
		Type: eventType,
		Data: data,
	}
	
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(e.writer, "__EVENT__%s__EVENT__\n", string(jsonBytes))
	if err != nil {
		return err
	}
	
	if f, ok := e.writer.(*os.File); ok {
		f.Sync()
	}
	return nil
}

func (e *Emitter) Message(content string) error {
	return e.Emit(EventMessage, map[string]interface{}{
		"content": content,
	})
}

func (e *Emitter) Status(status string) error {
	return e.Emit(EventStatus, map[string]interface{}{
		"status": status,
	})
}

func (e *Emitter) Confirm(prompt string, id string) error {
	return e.Emit(EventConfirm, map[string]interface{}{
		"prompt": prompt,
		"id":     id,
	})
}

func (e *Emitter) OpenFile(path string, split bool) error {
	return e.Emit(EventOpenFile, map[string]interface{}{
		"path":  path,
		"split": split,
	})
}

func (e *Emitter) OpenPlan(path string) error {
	return e.Emit(EventOpenPlan, map[string]interface{}{
		"path": path,
	})
}

func (e *Emitter) OpenSplit(testFile string, implFile string) error {
	return e.Emit(EventOpenSplit, map[string]interface{}{
		"test_file": testFile,
		"impl_file": implFile,
	})
}

func (e *Emitter) Complete() error {
	return e.Emit(EventComplete, map[string]interface{}{})
}

func (e *Emitter) Notify(message string, level string) error {
	return e.Emit(EventNotify, map[string]interface{}{
		"message": message,
		"level":   level,
	})
}
