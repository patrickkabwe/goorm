// Code generated by go generate; DO NOT EDIT.
package goorm

type GeneratedDB struct {
    *DB
    User *UserModel
    Profile *ProfileModel
    Post *PostModel
    Comment *CommentModel
}

func NewGoorm(cfg *GoormConfig) (*GeneratedDB, error) {
 	engine, err := NewEngine(cfg.Driver, cfg.DSN, cfg.Logger)
	if err != nil {
		return nil, err
	}
	db := NewDB(engine, cfg.Logger)	
    gdb := &GeneratedDB{
        DB: db,
        User: &UserModel{BaseModel: NewBaseModel[User](db)},
        Profile: &ProfileModel{BaseModel: NewBaseModel[Profile](db)},
        Post: &PostModel{BaseModel: NewBaseModel[Post](db)},
        Comment: &CommentModel{BaseModel: NewBaseModel[Comment](db)},
    }

    // Initialize relationships
    gdb.User.initRelations()
    gdb.Profile.initRelations()
    gdb.Post.initRelations()
    gdb.Comment.initRelations()

    return gdb, nil
}


type UserModel struct {
    *BaseModel[User]
}


func (m *UserModel) initRelations() {
    m.RegisterRelation(Relation{
        Name:       "Profile",
        Type:       "belongsTo",
        Model:      Profile{},
        ForeignKey: "ProfileID",
        References: "ID",
    })
    m.RegisterRelation(Relation{
        Name:       "Posts",
        Type:       "hasMany",
        Model:      Post{},
        ForeignKey: "UserID",
        References: "ID",
    })
}


type ProfileModel struct {
    *BaseModel[Profile]
}


func (m *ProfileModel) initRelations() {
    m.RegisterRelation(Relation{
        Name:       "User",
        Type:       "belongsTo",
        Model:      User{},
        ForeignKey: "UserID",
        References: "ID",
    })
}


type PostModel struct {
    *BaseModel[Post]
}


func (m *PostModel) initRelations() {
    m.RegisterRelation(Relation{
        Name:       "User",
        Type:       "belongsTo",
        Model:      User{},
        ForeignKey: "UserID",
        References: "ID",
    })
}


type CommentModel struct {
    *BaseModel[Comment]
}


func (m *CommentModel) initRelations() {
    m.RegisterRelation(Relation{
        Name:       "Post",
        Type:       "belongsTo",
        Model:      Post{},
        ForeignKey: "PostID",
        References: "ID",
    })
}


