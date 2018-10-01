package functions

import "github.com/tobyjsullivan/chalk/types"

type Application struct {
	Function  *Function
	Arguments []Argument
}

type Argument interface {
	Resolve() (types.Object, error)
}

func NewArgument(obj types.Object) Argument {
	return &objectArgument{obj}
}

func (app *Application) Resolve() (types.Object, error) {
	args := make([]types.Object, len(app.Arguments))

	var err error
	for i, arg := range app.Arguments {
		args[i], err = arg.Resolve()
		if err != nil {
			return nil, err
		}
	}

	return app.Function.Handler(args...)
}

type objectArgument struct {
	obj types.Object
}

func (arg *objectArgument) Resolve() (types.Object, error) {
	return arg.obj, nil
}
