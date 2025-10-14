package copier_test

type User struct {
	Name string
	//Birthday *time.Time
	NickName string
	Role     string
	Age      int32
	FakeAge  int32
	Notes    []string
	Flags    []byte
}

func (user User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	_User *User
	Name  string
	//Birthday  *time.Time
	NickName  string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	Flags     []byte
}

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}
