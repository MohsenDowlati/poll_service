package bootstrap

import "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"

type Application struct {
	Env   *Env
	Mongo mongo.Client
}

func App() Application {
	return AppWithEnv(NewEnv())
}

func AppWithEnv(env *Env) Application {
	app := &Application{Env: env}
	app.Mongo = NewMongoDatabase(app.Env)
	return *app
}

func (app *Application) CloseDBConnection() {
	CloseMongoDBConnection(app.Mongo)
}
