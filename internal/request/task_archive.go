package request

type BatchArchieveTask struct {
	IDs []string `json:"ids" validate:"required,min=1,max=20,dive,required"`
}
