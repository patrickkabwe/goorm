package goorm

type User struct {
	ID      string   `db_col:"id" goorm:"primary key,auto_increment,type:serial,column:id"`
	Name    string   `db_col:"name"`
	Email   string   `db_col:"email"`
	Age     int64    `db_col:"age" goorm:"default:30,check:(age > 0)"`
	Profile *Profile // Has one relationship
	Posts   []Post   // Has many relationship
}

type Profile struct {
	ID     string `db_col:"id"`
	UserID string `db_col:"user_id"`
	Bio    string `db_col:"bio"`
	User   *User  // Belongs to relationship
}

type Post struct {
	ID      string `db_col:"id"`
	Title   string `db_col:"title" goorm:"unique,default:'asdasd'"`
	Content string `db_col:"content"`
	UserID  string `db_col:"user_id"`
	User    *User  // Belongs to relationship
}

type Comment struct {
	ID      string `db_col:"id"`
	Content string `db_col:"content"`
}
