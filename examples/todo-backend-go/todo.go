package todo

import "context"

type GetAllRequest struct{}
type GetAllResponse struct {
	Todos []*Todo `json:"todos"`
}

//ftl:verb
func GetAll(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {
	todos := DAO.GetAll()
	return GetAllResponse{Todos: todos}, nil
}

type CreateResponse struct{}

//ftl:verb
func Create(ctx context.Context, req Todo) (CreateResponse, error) {
	DAO.Create(&req)
	return CreateResponse{}, nil
}
