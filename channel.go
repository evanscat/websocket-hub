package hub

type ChannelParam struct {
	BroadcastExceptSender bool
	CheckSenderAuthority bool
	Code                  string
	Type                  string
	Others                map[string][]string
}

type Channel interface {
	BroadcastHandler(client string, message []byte, excepts ...string)
	Run(hub *Hub)
	Code() string
	Register() chan<- *Client
	UnRegister() chan<- *Client
	Broadcast() chan<- *Message
	TotalClients() float64
	Type() string
}

type Option func(param *ChannelParam)

func SetBroadcastExceptSender() Option {
	return func(param *ChannelParam) {
		param.BroadcastExceptSender = true
		return
	}
}

func WithChannelName(name string) Option {
	return func(param *ChannelParam) {
		param.Code = name
		return
	}
}

func WithParams(params map[string][]string) Option{
	return func(param *ChannelParam) {
		param.Others = params
		return
	}
}