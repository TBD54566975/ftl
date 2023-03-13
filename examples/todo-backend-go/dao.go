package todo

import "errors"

// Copied from MIT licensed https://github.com/theramis/todo-backend-go-echo/tree/master

var DAO = NewInMemoryTodoDAO()

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Order     int    `json:"order"`
	Completed bool   `json:"completed"`
	URL       string `json:"url"`
}

type TodoDAO interface {
	Create(todo *Todo)
	GetAll() []*Todo
	Get(id int) (t *Todo, err error)
	Update(*Todo) (err error)
	DeleteAll()
	Delete(id int) (err error)
}

type InMemoryTodoDAO struct {
	Todos  []*Todo
	nextId int
}

func NewInMemoryTodoDAO() TodoDAO {
	t := new(InMemoryTodoDAO)
	t.Todos = make([]*Todo, 0)
	t.nextId = 1
	return t
}

func (r *InMemoryTodoDAO) Create(todo *Todo) {
	todo.ID = r.nextId
	r.Todos = append(r.Todos, todo)
	r.nextId++
}

func (r *InMemoryTodoDAO) GetAll() []*Todo {
	return r.Todos
}

func (r *InMemoryTodoDAO) DeleteAll() {
	r.Todos = make([]*Todo, 0)
}

func (r *InMemoryTodoDAO) Get(id int) (t *Todo, err error) {
	for _, t = range r.Todos {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("todo not found")
}

func (r *InMemoryTodoDAO) Delete(id int) (err error) {
	for i, t := range r.Todos {
		if t.ID == id {
			r.Todos = append(r.Todos[:i], r.Todos[i+1:]...)
			return nil
		}
	}
	return errors.New("todo not found")
}

func (r *InMemoryTodoDAO) Update(todo *Todo) (err error) {
	for i, t := range r.Todos {
		if t.ID == todo.ID {
			r.Todos[i] = todo
			return nil
		}
	}
	return errors.New("todo not found")
}
