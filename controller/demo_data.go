package controller

var (
	DemoUser = User{
		ID:            1,
		Name:          "TestUser",
		FollowCount:   0,
		FollowerCount: 0,
		IsFollow:      false,
	}

	usersLoginInfo = map[string]User{
		"zhangleidouyin": {
			ID:            1,
			Name:          "zhanglei",
			FollowCount:   10,
			FollowerCount: 5,
			IsFollow:      true,
		},
	}
)
