HTTP Playback
========================

Record HTTP sessions then play them back in the same order. This is useful for creating fake web services under
test conditions.

### Record

To add responses to a session use the /r/ endpoint with a session id:

POST `/r/[session id]`

```
{
    "Status": 200,
    "Headers": {},
    "Body": "bar",
    "Wait": 50
}
```

Explanation:

| Key             | Description                             |
| --------------- | --------------------------------------- |
| Status          | Response status                         |
| Headers         | Response headers                        |
| Body            | Response body                           |
| Wait            | Wait for N ms before returning response |


### Playback

Once responses have been added to a session they can be replayed using:

ANY `/p/[session id]/...`

Each time the session is called it will return a configured response. Responses
are returned in the same order they were configured. When no more
responses are available the response status is 0 and all other fields are
blank.

Note that `/p/` will accept any HTTP method and any uri AFTER the session id.
However the trailing slash after the session id is required.

## Example

```
go run main.go
```

meanwhile....

```
for body in {1..5}; do curl -X POST -d '{"Status": 200, "Body": "hi"}' localhost:8080/r/1; echo "\n"; done;

for num in {1..5}; do curl localhost:8080/p/1/; done;
```
