package reservation

type CheckboxBool bool

func (c *CheckboxBool) UnmarshalText(text []byte) error {
	*c = CheckboxBool(string(text) == "on")
	return nil
}

// todo split do vlastniho balicku?

type Handler struct { //todo prejmenovat package a struct
	Start       string       `form:"start_time" binding:"required"`
	End         string       `form:"end_time" binding:"required"`
	Area        int          `form:"area_id" binding:"required"` //todo tohle se asi na frontendu nenaplni
	Name        string       `form:"name" binding:"required"`
	Email       string       `form:"email" binding:"required"`
	Phone       string       `form:"phone" binding:"required"`
	PlayerCount int          `form:"players" binding:"required"`
	Notes       string       `form:"notes" binding:"required"`
	Reminder    CheckboxBool `form:"reminder" binding:"required"`
}
