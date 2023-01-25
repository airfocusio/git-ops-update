package internal

type Augmenter interface {
	RenderMessage(config Config, change Change) (string, string, error)
}
