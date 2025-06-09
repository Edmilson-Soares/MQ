package runtime

type Runtime struct {
	ID      string
	app     map[string]string
	env     map[string]string
	stripts map[string]string
}
type RuntimeData struct {
	ID      string            `json:"id"`
	App     map[string]string `json:"app"`
	Env     map[string]string `json:"env"`
	Stripts map[string]string `json:"stripts"`
}
