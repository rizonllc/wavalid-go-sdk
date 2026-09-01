# wavalid-go-sdk

Go SDK for the [wavalid](https://wavalid.com) WhatsApp number validation API. wavalid checks whether a phone number is registered and active on WhatsApp; it does not send messages and does not store phone numbers.

## Install

```bash
go get github.com/rizonllc/wavalid-go-sdk
```

## Authentication

Every request needs an API key, created from the wavalid dashboard. Never hardcode it — read it from an environment variable.

```go
import wavalid "github.com/rizonllc/wavalid-go-sdk"

client := wavalid.NewClient(os.Getenv("WAVALID_API_KEY"), "https://wavalid.com")
```

`baseURL` is the root domain only — do not append `/api` or `/api/v1`, the client adds that path itself.

## `client.Validate`

Validates a single phone number.

```go
result, err := client.Validate(ctx, "+14155551234", nil)
// result.Status: "valid" | "invalid" | "limit"
```

## `client.ValidateBulk`

Validates up to 100 phone numbers in one request.

```go
bulk, err := client.ValidateBulk(ctx, []string{"+14155551234", "+447911123456"}, nil)
// bulk.Results, bulk.CreditsUsed, bulk.CreditsRemaining
```

For more than 100 numbers, split the list into chunks of 100 and call `ValidateBulk` once per chunk.

## Errors

Both methods return an `*wavalid.ApiError` on any non-2xx response — always check `err`, never assume success.

```go
result, err := client.Validate(ctx, "+14155551234", nil)
if err != nil {
	var apiErr *wavalid.ApiError
	if errors.As(err, &apiErr) {
		if apiErr.Code == "rateLimitExceeded" {
			// back off and retry later
		}
		log.Printf("wavalid error %d %s: %s", apiErr.StatusCode, apiErr.Code, apiErr.Message)
		return
	}
	log.Fatal(err)
}
```

| statusCode | code                  | meaning                                                                                |
| ---------- | --------------------- | -------------------------------------------------------------------------------------- |
| 400        | (none)                | `phoneNumber` is not a valid international number, or `phoneNumbers` is empty/over 100 |
| 401        | (none)                | missing or invalid API key                                                             |
| 402        | `insufficientCredits` | account has no credits left                                                            |
| 404        | (none)                | `batchId` does not exist for this account                                              |
| 429        | `rateLimitExceeded`   | too many requests; back off                                                            |
| 502        | (none)                | upstream WhatsApp check failed; safe to retry                                          |

## Links

- API reference: https://wavalid.com/product/api
- Source / issues: https://github.com/rizonllc/wavalid-go-sdk
- License: MIT
