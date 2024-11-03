package goorm

type User struct {
	ID      int      `db_col:"id" goorm:"primary key,auto_increment,type:serial"`
	Name    string   `db_col:"name"`
	Email   string   `db_col:"email"`
	Age     int64    `db_col:"age" goorm:"default:30,check:(age > 0)"`
	Profile *Profile // Has one relationship
	Posts   []Post   // Has many relationship
}

type Profile struct {
	ID     int `db_col:"id" goorm:"primary key,auto_increment,type:serial"`
	UserID int    `db_col:"user_id"`
	Bio    string `db_col:"bio"`
	User   *User  // Belongs to relationship
}

type Post struct {
	ID      int `db_col:"id" goorm:"primary key,auto_increment,type:serial"`
	Title   string `db_col:"title" goorm:"unique,default:'asdasd'"`
	Content string `db_col:"content"`
	UserID  int    `db_col:"user_id"`
	User    *User  // Belongs to relationship
}

type Comment struct {
	ID      int `db_col:"id" goorm:"primary key,auto_increment,type:serial"`
	Content string `db_col:"content"`
}
