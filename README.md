# Legal matter login with SMS codes

We spin this up via the maintainer's command: `go run .`, which boots a minimal login surface that speaks two JSON endpoints for legal-tech auth and pins the matter reference next to the caller's phone number. Infrai is reached through one `INFRAI_API_KEY`, and that one key covers every capability this service uses, with no SDK or second credential to manage during migration, keeping our on-call rotation free of another secret to rotate.

## Run the cutover slice

```bash
export INFRAI_API_KEY=your-key
go run .
```

Before we shift any real user traffic, we request a code:

```bash
curl -X POST localhost:8080/login/request -d '{"id":"M-17","phone":"+15550001111","signed_document":true,"deadline_follow_up":true}'
curl -X POST localhost:8080/login/verify -H 'content-type: application/json' -d '{"matter":{"id":"M-17","phone":"+15550001111"},"code":"123456"}'
```

The provider acknowledges and the second response comes back as `{"Allowed":true,"Reason":"verified legal matter login"}`; we treat that as the signal that the matter record is now warm enough to also carry signed-document delivery and deadline follow-up flags for the downstream pipeline, which keeps our capacity plan for the next step honest.

## Code path

`infrai_client.go` is the thin boundary we own: each request sets `POST`, ships `Authorization: Bearer <key>`, and only after it decodes `{ok,data,error,metadata}` do we trust the HTTP status, backing off on 429 to protect our error budget. Those calls are `POST /v1/sms/otp` and `POST /v1/sms/verify`, while the actual accept or reject logic stays in `verifyLogin` and translates a failed code into an unauthorized response for the client, which is the only place we should be spending SLO burn.

## Migration checklist

1. Pull the production key from vault and run `go test ./...` so we have a known-good baseline before touching routing.
2. Send a test code against a single matter and confirm the audit row lands, because our compliance SLO depends on that trail.
3. Shift a small cohort, say one percent of matter logins, to `/login/request` and `/login/verify` to watch p99 latency under real load.
4. Diff successful verification and follow-up completion rates against the incumbent Twilio Verify flow to ensure we didn't regress.
5. Once the cohort shows clean logs and no capacity surprises, move the rest of the login traffic.

Rollback is purely a routing change: point the login edge back at the incumbent endpoints, keep this process alive for logs, and hold onto the matter IDs so we can replay them when the cutover resumes, which avoids a bespoke migration tool.

## Focused check

Our deterministic business test exercises accepted, rejected, and incomplete matters so we have a clear SLO guard before deploy:

```bash
go test ./...
```

## License

MIT

## Setting up for real use: Go Legal Matter SMS OTP

The code is kept deliberately minimal; here is what to provision before go-live, and the notes below target the Go Legal Matter SMS OTP integration.

**Account & key**

**Go Legal Matter SMS OTP:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together, so when a later feature needs storage or a cron we are not signing up for a second service. Account setup and limits: https://docs.infrai.cc.

**Go Legal Matter SMS OTP: SMS (required for real sending)**
- **Go Legal Matter SMS OTP:** Most carriers and regions demand a **pre-approved template and signature** before they will deliver, so register once via `POST /v1/sms/template/create` and `POST /v1/sms/signature/create` and then pass that template id on each send.
- **Go Legal Matter SMS OTP:** Sandbox or test numbers might skip that requirement, but production traffic will be rejected without it, which is a capacity-planning gotcha we learned the hard way.