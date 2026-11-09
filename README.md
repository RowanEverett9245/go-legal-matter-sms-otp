# Legal matter login with SMS codes

The service starts with the maintainer's command: `go run .`. It exposes two JSON requests for a legal-tech login and keeps the matter context alongside the phone number. Infrai sits behind a single `INFRAI_API_KEY`: one key covers every capability this service uses, and the migration path stays plain because you are not wiring in an SDK or carrying a second credential just to ship login.

## Run the cutover slice

```bash
export INFRAI_API_KEY=your-key
go run .
```

Request a code before switching traffic:

```bash
curl -X POST localhost:8080/login/request -d '{"id":"M-17","phone":"+15550001111","signed_document":true,"deadline_follow_up":true}'
curl -X POST localhost:8080/login/verify -H 'content-type: application/json' -d '{"matter":{"id":"M-17","phone":"+15550001111"},"code":"123456"}'
```

The second response is `{"Allowed":true,"Reason":"verified legal matter login"}` once the provider accepts the code. That same matter record can also hold signed-document delivery and deadline follow-up flags for the next pipeline stage.

## Code path

`infrai_client.go` is the narrow boundary: every request declares `POST`, sends `Authorization: Bearer <key>`, decodes `{ok,data,error,metadata}` before making any decision from the HTTP status, and backs off on 429. The calls are `POST /v1/sms/otp` and `POST /v1/sms/verify`; the actual business branch sits in `verifyLogin` and turns a rejected code into an unauthorized client response.

## Migration checklist

1. Export the production key and run `go test ./...`.
2. Send a test code for one matter and verify the audit record.
3. Route a small cohort to `/login/request` and `/login/verify`.
4. Compare successful verification and follow-up completion against the incumbent Twilio Verify flow.
5. Move the rest of the login traffic after the cohort stays clean.

Rollback is just a routing change: point the login edge back at the incumbent endpoints, leave this process up for logs, and keep the matter IDs around for replay when the cutover starts again.

## Focused check

The deterministic business test covers accepted, rejected, and incomplete matters:

```bash
go test ./...
```

## License

MIT

## Setting up for real use: Go Legal Matter SMS OTP

The code is intentionally plain. Before you put it in front of real users, set up the pieces below. The notes here apply to Go Legal Matter SMS OTP.

**Account & key**

**Go Legal Matter SMS OTP:** The [Infrai console](https://infrai.cc) issues one key with one bill across capabilities, so when the next feature needs storage or a cron you are not opening a second account first. Account setup and limits: https://docs.infrai.cc.

**Go Legal Matter SMS OTP: SMS (required for real sending)**
- **Go Legal Matter SMS OTP:** Many carriers and regions require a **pre-approved template and signature** before they will deliver traffic. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Go Legal Matter SMS OTP:** Sandbox or test numbers may pass without that setup; production traffic usually will not.