# goorm cli

Goorm cli tool is a command line tool that generates, migrates, and manages Go ORM models as well as migrations.

## Installation

```bash
go install github.com/patrickkabwe/goorm/goorm-cli@latest
```

## Usage

```bash
goorm-cli -h

Usage:
  goorm-cli [flags]
  goorm-cli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  generate    Generates a typed model
  help        Help about any command
  migrate     Generate a migration file
  push        Pushes the models to the database

Flags:
  -h, --help   help for goorm-cli

Use "goorm-cli [command] --help" for more information about a command.

```

## Commands

### Generate

Generates a typed model from a database schema.

```bash
goorm-cli generate
```

### Migrate

Generates a migration file.

```bash
goorm-cli migrate --name create_tables
```

Apply migrations

```bash
goorm-cli migrate --apply
```

### Push

Pushes the models to the database that generated with the generate command.

```bash
goorm-cli push
```

## License

MIT License

Copyright (c) 2024 [Patrick Kabwe](https://github.com/patrickkabwe)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
