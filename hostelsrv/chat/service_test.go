package chat

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testClient struct {
	readBuf    bytes.Buffer
	writeBuf   bytes.Buffer
	closeCount int
	panic      interface{}
}

func (cl *testClient) Read(p []byte) (n int, err error) {
	if cl.panic != nil {
		panic(cl.panic)
	}
	return cl.readBuf.Read(p)
}

func (cl *testClient) Write(p []byte) (n int, err error) {
	return cl.writeBuf.Write(p)
}

func (cl *testClient) Close() error {
	cl.closeCount++
	return nil
}

type testCommand struct {
	handleArgs      string
	outgoingMessage message
	panic           interface{}
}

func (cmd *testCommand) Handle(user identity, args string, outgoing chan<- message) {
	if cmd.panic != nil {
		panic(cmd.panic)
	}
	if cmd.outgoingMessage != "" {
		outgoing <- cmd.outgoingMessage
	}
	cmd.handleArgs = args
}

type testUnsubscriber struct {
	invoked bool
}

func (u *testUnsubscriber) Unsubscribe(user identity) {
	u.invoked = true
}

func TestServiceHandleClient_CorrectInput_CommandInvoked(t *testing.T) {
	cl := &testClient{}
	fmt.Fprintln(&cl.readBuf, "cmd1|arg1")
	fmt.Fprintln(&cl.readBuf, "cmd2|arg21|arg22")

	cmd1 := testCommand{outgoingMessage: "message1"}
	cmd2 := testCommand{outgoingMessage: "message2"}
	cmds := map[string]Command{
		"cmd1": &cmd1,
		"cmd2": &cmd2,
	}

	s := NewService(cmds, &testUnsubscriber{})
	s.HandleClient(cl)

	assert.Equal(t, "arg1", cmd1.handleArgs)
	assert.Equal(t, "arg21|arg22", cmd2.handleArgs)
	assert.Contains(t, cl.writeBuf.String(), "message1")
	assert.Contains(t, cl.writeBuf.String(), "message2")
}

func TestServiceHandleClient_EmptyArgs_CommandInvokedWithEmptyArgs(t *testing.T) {
	cl := &testClient{}
	fmt.Fprintln(&cl.readBuf, "cmd1|")
	fmt.Fprintln(&cl.readBuf, "cmd2")

	cmd1 := testCommand{}
	cmd2 := testCommand{}
	cmds := map[string]Command{
		"cmd1": &cmd1,
		"cmd2": &cmd2,
	}

	s := NewService(cmds, &testUnsubscriber{})
	s.HandleClient(cl)

	assert.Empty(t, cmd1.handleArgs)
	assert.Empty(t, cmd2.handleArgs)
}

func TestServiceHandleClient_UnknownCommand_ErrorWritten(t *testing.T) {
	cl := &testClient{}
	fmt.Fprintln(&cl.readBuf, "cmd1|arg1")
	fmt.Fprintln(&cl.readBuf, "cmd2|arg21|arg22")

	s := NewService(nil, &testUnsubscriber{})
	s.HandleClient(cl)

	assert.Contains(t, cl.writeBuf.String(), "Unknown command: cmd1.")
	assert.Contains(t, cl.writeBuf.String(), "Unknown command: cmd2.")
}

func TestServiceHandleClient_CommandPaniced_ErrorWritten(t *testing.T) {
	cl := &testClient{}
	fmt.Fprintln(&cl.readBuf, "cmd1|arg1")

	cmds := map[string]Command{
		"cmd1": &testCommand{panic: true},
	}

	s := NewService(cmds, &testUnsubscriber{})
	s.HandleClient(cl)

	assert.Contains(t, cl.writeBuf.String(), "Unexpected server error!")
}

func TestServiceHandleClient_CorrectHandle_ClosesConn(t *testing.T) {
	cl := &testClient{}
	fmt.Fprintln(&cl.readBuf, "cmd1|arg1")
	fmt.Fprintln(&cl.readBuf, "cmd2|arg2")

	cmds := map[string]Command{
		"cmd1": &testCommand{},
		"cmd2": &testCommand{},
	}

	s := NewService(cmds, &testUnsubscriber{})
	s.HandleClient(cl)

	assert.Equal(t, 1, cl.closeCount)
}

func TestServiceHandleClient_ReadPaniced_ClosesConn(t *testing.T) {
	cl := &testClient{panic: true}
	defer func() {
		assert.NotNil(t, recover())
		assert.Equal(t, 1, cl.closeCount)
	}()

	s := NewService(nil, &testUnsubscriber{})
	s.HandleClient(cl)
}

func TestServiceHandleClient_CorrectHandle_UnsubscribesClient(t *testing.T) {
	cl := &testClient{}
	fmt.Fprintln(&cl.readBuf, "cmd1|arg1")
	fmt.Fprintln(&cl.readBuf, "cmd2|arg2")

	cmds := map[string]Command{
		"cmd1": &testCommand{},
		"cmd2": &testCommand{},
	}
	uns := testUnsubscriber{}

	s := NewService(cmds, &uns)
	s.HandleClient(cl)

	assert.True(t, uns.invoked)
}

func TestServiceHandleClient_ReadPaniced_UnsubscribesClient(t *testing.T) {
	cl := &testClient{panic: true}
	uns := testUnsubscriber{}
	defer func() {
		assert.NotNil(t, recover())
		assert.True(t, uns.invoked)
	}()

	s := NewService(nil, &uns)
	s.HandleClient(cl)
}
