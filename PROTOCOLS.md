# Protocols

## Push

Client MUST send a POST to /api/v1/io/push/prepare with a
proto.PushPrepareRequest body. The server will respond with a
proto.PushPrepareResponse. Client SHOULD send data using a POST to
/api/v1/io/push with a stream (long-running http request) of proto.Chunk in
body. Client MAY request a transfer log using a POST to /api/v1/io/push/log
with a proto.PushLogRequest.

## Pull

Client MUST send a POST to /api/v1/io/pull/prepare with a
proto.PullPrepareRequest body. The server will respond with a
proto.PullPrepareResponse. Client SHOULD send a POST to /api/v1/io/pull with a
proto.PullRequest. Server will respond with a stream of proto.Chunk.
