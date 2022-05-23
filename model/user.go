package model

import (
	"log"
	initialize "simple-demo/initialize"
)

type User struct {
	Id            int64  `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	Password      string `json:"password"`
	Token         string `json:"token"`
	FollowCount   int64  `json:"follow_count,omitempty"`
	FollowerCount int64  `json:"follower_count,omitempty"`
}

func FindUserById(id int64) User {
	var user User
	initialize.GLOBAL_DB.Where(&User{
		Id: id,
	}).First(&user)
	return user
}

func FindUserByName(name string) User {
	var user User
	initialize.GLOBAL_DB.Where(&User{
		Name: name,
	}).First(&user)
	return user
}

func FindUserByToken(token string) User {
	var user User
	initialize.GLOBAL_DB.Where(&User{
		Token: token,
	}).First(&user)
	return user
}

func InsertUser(user User) int64 {
	result := initialize.GLOBAL_DB.Create(&user)
	log.Println(result)
	return user.Id
}
