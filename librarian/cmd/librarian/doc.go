/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUTHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Librarian is a tool for managing Google API client libraries.

Usage:

	librarian <command> [arguments]

The commands are:

	init      Initialize a new librarian-managed repository
	add       Add a library to be managed by librarian
	update    Regenerate client libraries
	config    Manage configuration
	remove    Remove a library from librarian management
	release   Release libraries

**Init Command**

The `init` command initializes a new `.librarian` directory in the current
working directory. It creates a `config.yaml` and an empty `state.yaml`.

**Add Command**

The `add` command adds a new library to the `state.yaml`. When `librarian add`
is run, it will run `git tag` to see if a version of the library exists that
matches the `release_tag_format` in the `config.yaml`. If so, it will populate
the `release` section of the library's state with the latest version and commit
hash.

**Update Command**

The `update` command regenerates the client libraries. It can be run for a
single library or for all libraries.

**Config Command**

The `config` command allows you to get and set configuration values.

**Remove Command**

The `remove` command removes a library from the `state.yaml`.

**Release Command**

The `release` command creates a new release for a library.
*/
package main