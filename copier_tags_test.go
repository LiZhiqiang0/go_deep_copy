package copier_test

type EmployeeTags struct {
	Name    string `copier:"must"`
	DOB     string
	Address string
	ID      int `copier:"-"`
}

type EmployeeTags2 struct {
	Name    string `copier:"must,nopanic"`
	DOB     string
	Address string
	ID      int `copier:"-"`
}

type EmployeeTags3 struct {
	Name    string
	DOB     string
	Address string
	ID      int
}

type User1 struct {
	Name    string
	DOB     string
	Address string `copier:"override"`
	ID      int
}

type User2 struct {
	DOB     string
	Address *string `copier:"override"`
	ID      int
}

type User3 struct {
	ID int
}

type Employee3 struct {
	ID int
}
