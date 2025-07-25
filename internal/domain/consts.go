package domain

// ISO 8601 weekday constants and mappings
const (
	Monday    = 1
	Tuesday   = 2
	Wednesday = 3
	Thursday  = 4
	Friday    = 5
	Saturday  = 6
	Sunday    = 7
)

// WeekdayNames maps ISO 8601 weekday numbers to their English names
var WeekdayNames = map[int]string{
	Monday:    "Monday",
	Tuesday:   "Tuesday",
	Wednesday: "Wednesday",
	Thursday:  "Thursday",
	Friday:    "Friday",
	Saturday:  "Saturday",
	Sunday:    "Sunday",
}

// WeekdayNumbers maps weekday numbers as strings to integers
var WeekdayNumbers = map[string]int{
	"1": Monday,
	"2": Tuesday,
	"3": Wednesday,
	"4": Thursday,
	"5": Friday,
	"6": Saturday,
	"7": Sunday,
}

// DefaultActiveDays represents Monday through Friday in ISO format
var DefaultActiveDays = []int{Monday, Tuesday, Wednesday, Thursday, Friday}

// DefaultRole is the default role name when none is configured
const DefaultRole = "On duty"
