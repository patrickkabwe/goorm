# goorm

Goorm is a ORM for Go that supports PostgreSQL, MySQL, and SQLite.

## Installation

```bash
go get github.com/patrickkabwe/goorm
```

## CLI Installation

```bash
go install github.com/patrickkabwe/goorm/cmd@latest
```

> [!NOTE] You need to install both `goorm` and `goorm-cli` to use the CLI.
> The `goorm-cli` is used to generate the `goorm` code and to manage the migrations.

## Disclaimer

> [!WARNING] This module is still under development. Use at your own risk. The API is experimental and subject to change. This module tries to create a simple and intuitive API for interacting with databases
> However, it is not a replacement for a professional ORM like [gorm](https://github.com/go-gorm/gorm) or [ent](https://github.com/ent/ent).

## Usage

```go
package main

import (
	"fmt"

	"github.com/patrickkabwe/goorm"
)

type User struct {
	ID   int    `json:"id" db_col:"id"`
	Name string `json:"name" db_col:"name"`
}

func main() {
	db := goorm.New(&goorm.GoormConfig{
		DSN:    "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		Driver: goorm.Postgres,
		Logger: goorm.NewDefaultLogger(),
	})

	defer db.Close()

	users, err := db.User.FindMany(db.P{
		Where: db.Where(
			db.Eq("name", "John"),
		),
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}
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
