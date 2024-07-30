# indexer.sdk

## Installation

```go
go get github.com/carbonable-labs/indexer.sdk
```

## Usage

```go
import (
 "github.com/carbonable-labs/indexer.sdk"
)

var testConfig = Config{
 AppName: "test_config",
  StartBlock: 1,
 Contracts: []Contract{
  {
   Name:    "project",
   Address: "0x0516d0acb6341dcc567e85dc90c8f64e0c33d3daba0a310157d6bba0656c8769",
   Events: map[string]string{
    "URI": "project:uri",
   },
  },
}
}

func main() {
  res, err := sdk.Configure(testConfigs)
  if err != nil {
    log.Fatal(err)
  }

  duplicates := make(map[string]int)

  cancel, err := sdk.RegisterHandler(res.ApppName, fmt.Sprintf("%s.event.%s.>", res.Hash, "0x0516d0acb6341dcc567e85dc90c8f64e0c33d3daba0a310157d6bba0656c8769"), func(s string, u uint64, re sdk.RawEvent) error {
  slog.Debug("event received")
  _, ok := duplicates[re.EventId]
  if !ok {
   duplicates[re.EventId] = 1
  } else {
   duplicates[re.EventId] = duplicates[re.EventId] + 1
  }
  fmt.Println("========= DUPLICATES =========")
  // logique d'index
  fmt.Println(duplicates)
  return nil
 })
 if err != nil {
  slog.Error("failed to register handler", err)
  return
 }
 defer cancel()

 // Gracefully shutdown
 done := make(chan os.Signal, 1)
 signal.Notify(done, os.Interrupt, syscall.SIGTERM)
 <-done
}
```

