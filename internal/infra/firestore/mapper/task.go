package mapper

import (
	"go-echo-demo/internal/infra/firestore/dto"
	"go-echo-demo/internal/model"

	"cloud.google.com/go/firestore"
)

// TaskMapper 将 Firestore DocumentSnapshot 映射为 model.Task
func TaskMapper(snap *firestore.DocumentSnapshot) (model.Task, error) {
	var d dto.Task
	if err := snap.DataTo(&d); err != nil {
		return model.Task{}, err
	}
	d.ID = snap.Ref.ID
	return *d.ToEntity(), nil
}
