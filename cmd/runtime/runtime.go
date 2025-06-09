package runtime

func New(app map[string]string, env map[string]string, stripts map[string]string) *Runtime {
	return &Runtime{
		ID:      app["id"],
		app:     app,
		env:     env,
		stripts: stripts,
	}
}

func (r *Runtime) GetEnv() map[string]string {
	return r.env
}
func (r *Runtime) GetApp() map[string]string {
	return r.app
}

func (r *Runtime) Run() error {

	return nil
}
