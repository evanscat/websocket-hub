package hub

import "errors"

type ChannelBuilder func(option ...Option) Channel

type factory struct {
	registry map[string]ChannelBuilder
}

var defaultFactory = &factory{registry: make(map[string]ChannelBuilder)}

func Factory() *factory {
	return defaultFactory
}

func (f *factory) GetChannel(tp string, options ...Option) (Channel, error) {
	if builder, ok := f.registry[tp]; ok {
		return builder(options...), nil
	}
	return nil, errors.New("no such channel type")
}

func (f *factory) Register(tp string, builder ChannelBuilder) {
	f.registry[tp] = builder
}

func init() {
	Factory().Register("default", NewDefaultChannel)
	Factory().Register("multi",NewMultiChannel)
}
