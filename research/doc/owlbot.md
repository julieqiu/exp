OwlBot Configuration Schema
===========================

OwlBot uses YAML configuration files to automate post-processing of generated code. The configuration file is typically named `.OwlBot.yaml` or `.OwlBot-hermetic.yaml`.

File Locations
--------------

| Language | Config File Location                  |
|----------|---------------------------------------|
| Node.js  | `.github/.OwlBot.yaml`                |
| Java     | `.OwlBot-hermetic.yaml` (per library) |
| Ruby     | `.OwlBot.yaml` (per library)          |
| PHP      | `.github/.OwlBot.yaml`                |

Schema
------

```yaml
# Docker image for OwlBot post-processor
docker:
  image: gcr.io/cloud-devrel-public-resources/owlbot-<language>:latest

# Patterns for copying generated code from staging to final locations
deep-copy-regex:
  - source: /google/cloud/secretmanager/v1/.*-java
    dest: /owl-bot-staging/$1/google-cloud-secretmanager/src

# Patterns for removing files
deep-remove-regex:
  - /owl-bot-staging

# Files to preserve (not overwrite during copy)
deep-preserve-regex:
  - /src/main/java/.*/package-info\.java$

# Begin and end markers for generated code sections
begin-after-commit-hash: abc123
```

Fields
------

### docker

Specifies the Docker image used for post-processing.

```yaml
docker:
  image: gcr.io/cloud-devrel-public-resources/owlbot-java:latest
```

### deep-copy-regex

List of source/destination patterns for copying generated files. Uses regex capture groups.

```yaml
deep-copy-regex:
  - source: /google/cloud/(.*)/(v[^/]+)/.*-java/proto-google-(cloud-.*)/src
    dest: /owl-bot-staging/$1/$2/proto-google-$3/src
  - source: /google/cloud/(.*)/(v[^/]+)/.*-java/grpc-google-(cloud-.*)/src
    dest: /owl-bot-staging/$1/$2/grpc-google-$3/src
```

### deep-remove-regex

Patterns for files/directories to remove after processing.

```yaml
deep-remove-regex:
  - /owl-bot-staging
```

### deep-preserve-regex

Patterns for files that should not be overwritten.

```yaml
deep-preserve-regex:
  - /src/main/java/.*/SomeHandwrittenFile\.java$
```

Language-Specific Examples
--------------------------

### Java (.OwlBot-hermetic.yaml)

```yaml
docker:
  image: gcr.io/cloud-devrel-public-resources/owlbot-java:latest

deep-copy-regex:
  - source: /google/cloud/secretmanager/(v[^/]+)/.*-java/proto-google-cloud-secretmanager-\1/src
    dest: /owl-bot-staging/$1/proto-google-cloud-secretmanager-$1/src
  - source: /google/cloud/secretmanager/(v[^/]+)/.*-java/grpc-google-cloud-secretmanager-\1/src
    dest: /owl-bot-staging/$1/grpc-google-cloud-secretmanager-$1/src
  - source: /google/cloud/secretmanager/(v[^/]+)/.*-java/gapic-google-cloud-secretmanager-\1/src
    dest: /owl-bot-staging/$1/google-cloud-secretmanager/src
```

### Node.js (.github/.OwlBot.yaml)

```yaml
docker:
  image: gcr.io/cloud-devrel-public-resources/owlbot-nodejs-mono-repo:latest

deep-copy-regex:
  - source: /google/cloud/secretmanager/v1/.*-nodejs/(.*)
    dest: /packages/google-cloud-secretmanager/src/$1
```

### Ruby (.OwlBot.yaml)

```yaml
docker:
  image: gcr.io/cloud-devrel-public-resources/owlbot-ruby:latest

deep-copy-regex:
  - source: /google/cloud/secret_manager/v1/.*-ruby/(.*)
    dest: /google-cloud-secret_manager-v1/$1
```

### PHP (.github/.OwlBot.yaml)

```yaml
docker:
  image: gcr.io/cloud-devrel-public-resources/owlbot-php:latest

deep-copy-regex:
  - source: /google/cloud/secretmanager/v1/.*-php/(.*)
    dest: /SecretManager/$1
```

Synthtool
---------

Synthtool (`github.com/googleapis/synthtool`) is a separate code generation automation tool that predates OwlBot. Some languages still use synthtool while others have migrated away.

### Still Using Synthtool

| Language | Tool Chain                         |
|----------|------------------------------------|
| Python   | synthtool + gapic-generator-python |
| PHP      | synthtool + gapic-generator-php    |
| Java     | synthtool + gapic-generator-java   |
| C#       | synthtool + gapic-generator-csharp |

These languages use `synth.py` files in their repositories, executed by Autosynth nightly.

### Migrated from Synthtool to OwlBot

| Language | Previous  | Current |
|----------|-----------|---------|
| Node.js  | synthtool | OwlBot  |
| Ruby     | synthtool | OwlBot  |

These languages previously used `synth.py` files with synthtool but have since migrated to OwlBot for post-processing.

### Never Used Synthtool

| Language | Alternative                   |
|----------|-------------------------------|
| Go       | Internal gapicgen + Librarian |
| C++      | Internal generator scripts    |
| Rust     | Sidekick (Librarian)          |

Deprecation Notice
------------------

OwlBot is deprecated and being phased out. Languages are migrating to:

-	**Go, Python, Rust:** Librarian (`github.com/googleapis/librarian`\)
-	**C#:** ReleaseManager tool
-	**C++:** Internal generator scripts
