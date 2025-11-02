# Surfer

Surfer generates gcloud surfaces.

Input:
- protos from github.com/googleapis/googleapis
- service config from github.com/googleapis/googleapis
- custom config at gcloud.yaml

Testdata is at:

testdata/input/gcloud.yaml
testdata/input/googleapis/

Expected output is at:

testdata/output/

Surfer takes the input, uses internal/api to generate an internal
representation of the API, and then uses that to generate the output.
