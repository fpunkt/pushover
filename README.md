# pushover - a wrapper for using pushover.net service

Small, zero dependancy wrapper to send messages via pushover.net.
Application and receiver keys are read from a `.json` file
so you don't have to store them in your application or repository.
Typically you would keep them in a local directory on your machine, like `/usr/local/etc/pushover.json`.
The number of messages per time interval can be throttled.

Messages can be send in background (use `Send()`) or with error checking the result using `SendAndWait()`

## Example JSON config file

Use the `app` section for your pushover application keys, the `rec` section for receiver keys.

```json
{
    "app": {
        "a1": "app1",
        "a2": "app2"
    },
    "rec": {
        "r1": "rec1",
        "r2": "rec1"
    }
}
```

## Example Usage

```go
import "github.com/fpunkt/pushover"

main() {
    p := pushover.MustLoad("sample.json") // usually like "/usr/local/etc/pushover.json"
    m := p.MustMessage("a1", "r2") // prepare message for application a1 and receiver r2

    m.Send("Message Title", "for application a1 and receiver r2") // this message will be send

    m.Throttle(5 * time.Minute) // limit to one message every five minutes

    for i :=0; i<10; i++ {
        m.Send("Fancy Title", "only one message is gonna be send")
    }
}
```

## Author

fpunkt@icloud.com
