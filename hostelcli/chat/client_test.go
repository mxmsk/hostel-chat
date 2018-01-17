package chat

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testServer struct {
	r bytes.Buffer
	w bytes.Buffer
}

func (srv *testServer) Read(p []byte) (n int, err error) {
	return srv.r.Read(p)
}

func (srv *testServer) Write(p []byte) (n int, err error) {
	return srv.w.Write(p)
}

func TestClientAddSubscription_GivenRoomNick_SubscriptionAdded(t *testing.T) {
	cl := NewClient(nil)

	cl.AddSubscription("room1", "nick1")
	assert.Contains(t, cl.subscriptions.String(), "|room1:nick1")

	cl.AddSubscription("room2", "nick2")
	assert.Contains(t, cl.subscriptions.String(), "|room2:nick2")
}

func TestClientAddSubscription_GivenRoomNick_TracksDefaultRoom(t *testing.T) {
	cl := NewClient(nil)

	cl.AddSubscription("room1", "nick1")
	assert.Equal(t, "room1", cl.defaultRoom)

	cl.AddSubscription("room2", "nick2")
	assert.Equal(t, "room2", cl.defaultRoom)
}

func TestClientRun_HasSubscriptions_SubscribesToRooms(t *testing.T) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	srv := &testServer{}

	cl := NewClient(srv)
	cl.AddSubscription("room1", "nick1")
	cl.AddSubscription("room2", "nick2")

	cl.Run(in, out)

	assert.Equal(t, "subscribe|room1:nick1|room2:nick2\n", srv.w.String())
}

func TestClientRun_PayloadMessage_PublishedToServer(t *testing.T) {
	testCases := []struct {
		msg      string
		expected string
	}{
		{
			msg:      "/room1 msg1",
			expected: "publish|room1|msg1",
		}, {
			msg:      "/room2 msg2",
			expected: "publish|room2|msg2",
		}, {
			msg:      "msg3", // default room path
			expected: "publish|room2|msg3",
		},
	}

	for _, testCase := range testCases {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		srv := &testServer{}

		cl := NewClient(srv)
		cl.AddSubscription("room1", "nick1")
		cl.AddSubscription("room2", "nick2")

		in.WriteString(testCase.msg)
		cl.Run(in, out)

		s := strings.Split(srv.w.String(), "\n")
		assert.Len(t, s, 3)
		assert.Equal(t, testCase.expected, s[1])
	}
}

func TestClientRun_EmptyMessage_NotPublishedToServer(t *testing.T) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	srv := &testServer{}

	cl := NewClient(srv)
	cl.AddSubscription("room1", "nick1")
	cl.AddSubscription("room2", "nick2")

	in.WriteString("\n")
	cl.Run(in, out)

	s := strings.Split(srv.w.String(), "\n")
	assert.Len(t, s, 2)
}
