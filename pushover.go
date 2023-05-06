package pushover

// Small, zero dependancy wrapper for sending pushover messages.
// (c) fpunkt@icloud.com

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Holding application and user/group keys to generate Messages. Use Load() or MustLoad() to
// initiate a Pushover structure, use the Message() function to generate messages.
type Pushover struct {
	App map[string]string
	Rec map[string]string
}

// Pushover Message for specific Application and Receiver keys.
// Message title and text are passed to the Send() method. A message can be reused
// to send arbritary number of messages. Messages can be throttled using Throttle().
type Message struct {
	app, rec string

	// Limit number of messages send to 1 message every throttle period
	throttle time.Duration
	lastsent time.Time
}

// Open app/usr database (typically like /usr/local/etc/pushover.json) or panic.
func MustLoad(fname string) Pushover {
	p, err := Load(fname)
	if err != nil {
		panic("Pushover Open: " + err.Error())
	}
	return p
}

// Load your application and receiver keys from a json-file
func Load(fname string) (Pushover, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return Pushover{}, err
	}
	p := Pushover{}
	err = json.Unmarshal(b, &p)
	return p, err
}

// Check if all apps are valid. Can be used for early error/typo discovery
func (p *Pushover) HasApp(keys ...string) bool {
	for _, k := range keys {
		if _, ok := p.App[k]; !ok {
			return false
		}
	}
	return true
}

// Check if all receivers are valid. Can be used for early error/typo discovery
func (p *Pushover) HasRec(keys ...string) bool {
	for _, k := range keys {
		if _, ok := p.Rec[k]; !ok {
			return false
		}
	}
	return true
}

// Check if all keys are valid, otherwise panic.
// Used for early bailout if you have typos in keys, like
//
// p := pushover.MustOpen("myfile.json").MustRec("a", "b", "c")
func (p *Pushover) MustRec(keys ...string) *Pushover {
	if !p.HasRec(keys...) {
		panic("Invalid receiver for pushover")
	}
	return p
}

// Create a Message for given Application and Receiver keys.
// The Message can be sent later with given title and text, a message can be sent multiple times.
// Message validates the pushover Application and Receiver key
//
//	p := pushover.MustOpen("/usr/local/etc/pushover.json")
//	m, _ := Message("HomeControl", "InfoGroup")
//	m.Send("Hello", "there")
func (p *Pushover) Message(app, receiver string) (Message, error) {
	a, aok := p.App[app]
	r, rok := p.Rec[receiver]
	m := Message{app: a, rec: r}
	if !aok {
		return m, fmt.Errorf("invalid pushover application: %s", app)
	}
	if !rok {
		return m, fmt.Errorf("invalid pushover receiver: %s", receiver)
	}
	return m, nil
}

// Create a Message, panics if application or receiver key cannot be found.
//
//	p := pushover.MustOpen("/usr/local/etc/pushover.json")
//	m := MustMessage("HomeControl", "InfoGroup")
//	m.Send("Hello", "there")
func (p *Pushover) MustMessage(app, receiver string) Message {
	m, err := p.Message(app, receiver)
	if err != nil {
		panic(fmt.Sprintf("pushover cannot create message for app=%s, rec=%s", app, receiver))
	}
	return m
}

// Error that is returned when messages are being send to fast and discarded.
var ErrThrottled = errors.New("pushover sending too fast - throttled")

// Reset throttle timer, next message will be sent unconditionally.
func (m *Message) ResetThrottle() { m.lastsent = time.Time{} }

// Limit messages to one message per specified intervall
func (m *Message) Throttle(d time.Duration) {
	m.throttle = d
	m.ResetThrottle()
}

func (m *Message) runThrottled(fn func() error) error {
	now := time.Now()
	if m.throttle > 0 && now.Sub(m.lastsent) < m.throttle {
		return ErrThrottled
	}
	m.lastsent = now
	return fn()
}

func (m *Message) pushover(title, message string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.PostForm("https://api.pushover.net/1/messages.json", url.Values{
		"token":   {m.app},
		"user":    {m.rec},
		"message": {message},
		"title":   {title},
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Only 500 errors will not respond a readable result
	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("internal server error")
	}
	_, err = io.ReadAll(resp.Body)
	return fmt.Errorf("cannot read body: %w", err)
}

// Send a message with timeout. This function blocks until the message is successfully
// sends and answer is received from the server.
// If throttled, the functions returns immediately without trying to send the
// message.
func (m *Message) SendAndWait(title, message string, timeout time.Duration) error {
	return m.runThrottled(func() error { return m.pushover(title, message, timeout) })
}

// Send message in background, return immediately. Network errors
// will only occur in background and are silently dropped.
// Only ErrThrottled is raised, if applicable
func (m *Message) Send(title, message string) error {
	return m.runThrottled(func() error { go m.pushover(title, message, 0); return nil })
}
