# 🚀 goorm

Goorm is a QueryBuilder\ORM for Go that supports PostgreSQL, MySQL, and SQLite. Built with simplicity and flexibility in mind.

## 📦 Installation

```bash
go get github.com/patrickkabwe/goorm
```

> [!NOTE]
> 🔄 The Goorm ORM is yet to be implemented

## ⚠️ Disclaimer

> [!WARNING]
> This module is still under development. Use at your own risk. The API is experimental and subject to change. This module tries to create a simple and intuitive API for interacting with databases. However, it is not a replacement for a professional database management system (DBMS) like PostgreSQL, MySQL, or SQLite.

## 🎯 Usage

```go
package main

import (
	"fmt"
	"os"
	"github.com/patrickkabwe/goorm"
)

type User struct {
	ID      int64    `db:"id"`
	Name    string   `db:"name"`
	Email   string   `db:"email"`
}

func main() {
	db, err := sql.Open(
		string(goorm.Postgres),
		os.Getenv("POSTGRES_DSN"),
	)

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	qb = orm.NewQueryBuilder(db, &orm.PostgreSQL{}, nil)
	ctx := context.Background()
	email := "test@gmail.com"
	user := &User{}
	err := qb.
		InsertInto("users").
		Columns("name", "email").
		Values("patrick", email).
		Returning(ctx, user, "id", "email")

	fmt.Println(user)
}
```

## ✨ Features

- 🛠️ **Flexible Query Building**
  - SELECT with DISTINCT support
  - INSERT, UPDATE, DELETE operations
  - Complex JOIN operations
  - WHERE clauses with AND, OR, NOT
  - GROUP BY, HAVING, ORDER BY
  - LIMIT and OFFSET pagination
- 🔒 **Type Safety**
  - Strongly typed parameters
  - Struct mapping for results
- 📊 **Database Support**
  - PostgreSQL
  - MySQL
  - SQLite (coming soon)
- 🔄 **Advanced Features**
  - Transaction support
  - RETURNING clause
  - Custom logger integration
  - Nested struct mapping
  - Auto table name prefixing

## 📄 License

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

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Open issues for bugs or feature requests
- Submit pull requests
- Improve documentation
- Share feedback

## Made with ❤️ by [Patrick Kabwe](https://github.com/patrickkabwe)

⭐️ If you find this project helpful, please give it a star!