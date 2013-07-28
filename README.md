
A simple caching proxy for a single upstream service.

    go get github.com/pranavraja/front

# Example

Say you run `npm install` a lot, and share code with your coworkers who also run `npm install`.
You want to front the npm registry with a caching server to speed up your builds, and stop wasting community resources.

Try this:

    front --upstream registry.npmjs.org

and then:

    npm install --registry http://localhost:8080

# Features

- Rewrite upstream URLs in responses, e.g. registry.npmjs.org => localhost:8080
- Cache responses based on a TTL (which can be overriden with the `--ttl` flag)

# Running the tests

Clone the repo, and:

    go test ./cache

